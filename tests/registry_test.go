package tests

import (
	"errors"
	"fakery"
	"io/fs"
	"sort"
	"sync"
	"testing"
	"testing/fstest"
)

// countingFS wraps an fs.FS and counts how often each file is opened
type countingFS struct {
	fsys  fs.FS
	mu    sync.Mutex
	opens map[string]int
}

func newCountingFS(fsys fs.FS) *countingFS {
	return &countingFS{fsys: fsys, opens: make(map[string]int)}
}

func (c *countingFS) Open(name string) (fs.File, error) {
	c.mu.Lock()
	c.opens[name]++
	c.mu.Unlock()
	return c.fsys.Open(name)
}

func (c *countingFS) count(name string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.opens[name]
}

func file(content string) *fstest.MapFile {
	return &fstest.MapFile{Data: []byte(content)}
}

// A small, self contained data set used by the registry tests
func testData() fstest.MapFS {
	return fstest.MapFS{
		"en_US/names.json": file(`{
			"first_name_male": ["John"],
			"first_name_female": ["Jane"],
			"last_name": ["Doe"],
			"prefix_male": ["Mr."],
			"prefix_female": ["Ms."],
			"suffix_male": ["Jr."],
			"suffix_female": ["PhD"]
		}`),
		"en_GB/names.json": file(`{
			"first_name_male": ["Oliver"],
			"first_name_female": ["Amelia"],
			"last_name": ["Smith"],
			"prefix_male": ["Sir"],
			"prefix_female": ["Dame"],
			"suffix_male": ["OBE"],
			"suffix_female": ["MBE"]
		}`),
		"en_US/address.json": file(`{"states": ["Texas"], "state_abbrevs": ["TX"]}`),
		"generic/address.json": file(`{
			"city_prefixes": ["North"],
			"city_suffixes": ["ville"],
			"building_suffixes": ["Towers"],
			"street_suffixes": ["Street"],
			"countries": ["United States", "United Kingdom"],
			"country_codes": ["US", "GB"]
		}`),
		"generic/jobs.json": file(`{"jobs": ["Baker"]}`),
		"generic/words.json": file(`{
			"adjectives": ["quick"],
			"adjectives_positive": ["happy"],
			"adjectives_negative": ["sad"],
			"adverbs": ["quickly"]
		}`),
		"generic/currency.json": file(`[
			{"country": "India", "code": "INR", "currency": "Indian Rupee"}
		]`),
		"generic/cars/acme.json": file(`[{"model": "Roadrunner", "type": "Coupe"}]`),
		"generic/car.json": file(`{
			"car_makers": ["Acme"],
			"car_categories": ["Coupe"],
			"car_series": ["X"],
			"car_category_adjectives": ["Sporty"],
			"car_fuel_types": ["Petrol"]
		}`),
		"generic/internet.json": file(`{
			"common_tlds_weighted": [".com:0.75", ".org:0.25"],
			"bad_weights": [".com:0.5", ".org:0.2"],
			"no_weights": [".com", ".org"],
			"fake_email_domains": ["example.com"],
			"free_email_domains": ["mail.example"]
		}`),
		"generic/malformed.json":  file(`{"broken": [`),
		"generic/wrongshape.json": file(`{"not": ["an array"]}`),
		"en_GB/malformed.json":    file(`[1, 2`),
		"en_US/malformed.json":    file(`{"fine": ["yes"]}`),
	}
}

func newTestRegistry() (*fakery.Registry, *countingFS) {
	cfs := newCountingFS(testData())
	return fakery.NewRegistry(cfs), cfs
}

func newTestFakery(r *fakery.Registry, locale string) *fakery.Fakery {
	f := fakery.NewFromLocale(locale)
	f.SetRegistry(r)
	return f
}

func TestBuiltinResourcesRegistered(t *testing.T) {
	r := fakery.DefaultRegistry()

	builtin := fakery.BuiltinResources()
	registered := r.Resources()
	Expect(t, len(builtin), len(registered))
	Expect(t, true, sort.SliceIsSorted(registered, func(i, j int) bool {
		return registered[i].Name < registered[j].Name
	}))

	for _, spec := range builtin {
		got, ok := r.Spec(spec.Name)
		Expect(t, true, ok, spec.Name)
		Expect(t, spec.Kind, got.Kind, spec.Name)
	}

	spec, ok := r.Spec(fakery.ResourceCurrency)
	Expect(t, true, ok)
	Expect(t, fakery.MapResource, spec.Kind)

	// Car model files are covered by a pattern
	spec, ok = r.Spec("cars/ford.json")
	Expect(t, true, ok)
	Expect(t, fakery.MapResource, spec.Kind)
	Expect(t, fakery.ResourceCarModels, spec.Name)

	// Unknown resources are implicitly list based
	spec, ok = r.Spec("unknown.json")
	Expect(t, false, ok)
	Expect(t, fakery.ListResource, spec.Kind)
}

func TestRepeatedAccessReusesCachedData(t *testing.T) {
	r, cfs := newTestRegistry()

	Expect(t, false, r.Loaded(fakery.ResourceNames, "en_US"))
	first := r.Load(fakery.ResourceNames, "en_US")
	second := r.Load(fakery.ResourceNames, "en_US")

	Expect(t, true, first == second)
	Expect(t, true, r.Loaded(fakery.ResourceNames, "en_US"))
	Expect(t, 1, cfs.count("en_US/names.json"))
	Expect(t, nil, first.Err())

	// Generators go through the same cache
	f := newTestFakery(r, "en_US")
	for i := 0; i < 50; i++ {
		Expect(t, "Doe", f.LastName())
		name := f.Name()
		Expect(t, true, name == "John Doe" || name == "Jane Doe", name)
	}
	Expect(t, 1, cfs.count("en_US/names.json"))

	// A second Fakery sharing the registry shares the cache too
	g := newTestFakery(r, "en_US")
	Expect(t, "Doe", g.LastName())
	Expect(t, 1, cfs.count("en_US/names.json"))
}

func TestStructDataDecodedOnce(t *testing.T) {
	r, cfs := newTestRegistry()
	f := newTestFakery(r, "en_US")

	for i := 0; i < 20; i++ {
		Expect(t, "Acme", f.CarMake())
		Expect(t, "Roadrunner", f.CarModel())
		Expect(t, "Petrol", f.CarType())
	}
	Expect(t, 1, cfs.count("generic/car.json"))
	Expect(t, 1, cfs.count("generic/cars/acme.json"))
}

func TestDifferentLocalesAreIsolated(t *testing.T) {
	r, cfs := newTestRegistry()

	us := r.Load(fakery.ResourceNames, "en_US")
	gb := r.Load(fakery.ResourceNames, "en_GB")

	Expect(t, true, us != gb)
	Expect(t, "en_US", us.Locale())
	Expect(t, "en_GB", gb.Locale())
	Expect(t, "Doe", us.Get("last_name")[0])
	Expect(t, "Smith", gb.Get("last_name")[0])
	Expect(t, 1, cfs.count("en_US/names.json"))
	Expect(t, 1, cfs.count("en_GB/names.json"))

	fUS := newTestFakery(r, "en_US")
	fGB := newTestFakery(r, "en_GB")
	for i := 0; i < 20; i++ {
		Expect(t, "Doe", fUS.LastName())
		Expect(t, "Smith", fGB.LastName())
	}

	// Switching locale on an instance picks the other locale's data
	fUS.SetLocale("en_GB")
	Expect(t, "en_GB", fUS.Locale())
	Expect(t, "Smith", fUS.LastName())
}

func TestLocaleFallsBackToDefaultLocale(t *testing.T) {
	r, cfs := newTestRegistry()

	// en_GB has no address.json, en_US data is used
	gb := r.Load(fakery.ResourceAddress, "en_GB")
	Expect(t, nil, gb.Err())
	Expect(t, fakery.DefaultLocale, gb.Locale())
	Expect(t, true, gb == r.Load(fakery.ResourceAddress, fakery.DefaultLocale))
	Expect(t, "Texas", newTestFakery(r, "en_GB").State())

	// Lookup of the missing file is attempted only once
	r.Load(fakery.ResourceAddress, "en_GB")
	Expect(t, 1, cfs.count("en_GB/address.json"))
	Expect(t, 1, cfs.count("en_US/address.json"))
}

func TestLocaleFallsBackToGenericLocale(t *testing.T) {
	r, _ := newTestRegistry()

	data := r.Load(fakery.ResourceJobs, "fr_FR")
	Expect(t, nil, data.Err())
	Expect(t, fakery.GenericLocale, data.Locale())
	Expect(t, "Baker", data.Get("jobs")[0])
}

func TestGenericLocaleDoesNotFallBack(t *testing.T) {
	r, _ := newTestRegistry()

	// names.json exists for en_US but not for generic
	data, err := r.Lookup(fakery.ResourceNames, fakery.GenericLocale)
	Expect(t, true, errors.Is(err, fs.ErrNotExist))
	Expect(t, fakery.GenericLocale, data.Locale())
	Expect(t, 0, data.Len())
}

func TestDefaultDataLocaleFallback(t *testing.T) {
	// The bundled en_GB locale has names but no address data
	f := fakery.NewFromLocale("en_GB")
	Expect(t, true, len(f.State()) > 0)
	Expect(t, true, len(f.StateAbbr()) > 0)
	Expect(t, true, len(f.Name()) > 0)

	a := f.Address()
	Expect(t, true, len(a.PostalCode) > 0)
	Expect(t, "United Kingdom", a.Country)
}

func TestMissingResource(t *testing.T) {
	r, cfs := newTestRegistry()

	data, err := r.Lookup("missing.json", fakery.GenericLocale)
	Expect(t, true, err != nil)
	Expect(t, true, errors.Is(err, fs.ErrNotExist))
	Expect(t, true, data != nil)
	Expect(t, true, errors.Is(data.Err(), fs.ErrNotExist))
	Expect(t, true, errors.Is(data.Load(), fs.ErrNotExist))
	Expect(t, 0, data.Len())
	Expect(t, 0, len(data.Keys()))
	Expect(t, true, data.Get("anything") == nil)

	// Failures are cached as well
	again, err2 := r.Lookup("missing.json", fakery.GenericLocale)
	Expect(t, true, again == data)
	Expect(t, true, err2 == err)
	Expect(t, 1, cfs.count("generic/missing.json"))

	// Missing in every locale of the fallback chain
	data, err = r.Lookup("missing.json", "en_GB")
	Expect(t, true, errors.Is(err, fs.ErrNotExist))
	Expect(t, 0, data.Len())
}

func TestMissingResourceInGenerators(t *testing.T) {
	// An empty data source: generators return empty values, not panic
	f := fakery.New()
	f.SetRegistry(fakery.NewRegistry(fstest.MapFS{}))

	Expect(t, "", f.Adjective())
	Expect(t, "", f.LastName())
	Expect(t, "", f.CurrencyCode())
	Expect(t, "", f.CarMake())
	Expect(t, "", f.TLD())

	// Restoring the default registry restores the data
	f.SetRegistry(nil)
	Expect(t, true, f.Registry() == fakery.DefaultRegistry())
	Expect(t, true, len(f.Adjective()) > 0)
}

func TestMalformedJSON(t *testing.T) {
	r, cfs := newTestRegistry()

	data, err := r.Lookup("malformed.json", fakery.GenericLocale)
	Expect(t, true, err != nil)
	Expect(t, false, errors.Is(err, fs.ErrNotExist))
	Expect(t, true, data.Err() == err)
	Expect(t, 0, data.Len())
	Expect(t, true, data.Get("broken") == nil)

	r.Load("malformed.json", fakery.GenericLocale)
	Expect(t, 1, cfs.count("generic/malformed.json"))

	// Malformed data does not fall back, the error surfaces
	data, err = r.Lookup("malformed.json", "en_GB")
	Expect(t, true, err != nil)
	Expect(t, "en_GB", data.Locale())
	Expect(t, 0, cfs.count("en_US/malformed.json"))

	// Data in the wrong shape for its kind is malformed too
	Expect(t, nil, r.Register(fakery.ResourceSpec{Name: "wrongshape.json", Kind: fakery.MapResource}))
	data, err = r.Lookup("wrongshape.json", fakery.GenericLocale)
	Expect(t, true, err != nil)
	Expect(t, true, data.IsMap())
	Expect(t, 0, data.Len())
}

func TestMapBasedData(t *testing.T) {
	r, _ := newTestRegistry()

	data := r.Load(fakery.ResourceCurrency, fakery.GenericLocale)
	Expect(t, nil, data.Err())
	Expect(t, true, data.IsMap())
	Expect(t, 1, data.Len())

	f := newTestFakery(r, "en_US")
	item := data.RandomWeightedItem(f)
	Expect(t, "INR", item["code"])

	c := f.Currency()
	Expect(t, "Indian Rupee", c.Name)
	Expect(t, "INR", c.Code)
	Expect(t, "India", c.Country)
	Expect(t, "INR", f.CurrencyCode())
	Expect(t, "Indian Rupee", f.CurrencyName())
	Expect(t, "India", f.CurrencyCountry())

	// Pattern registered map data
	models := r.Load("cars/acme.json", fakery.GenericLocale)
	Expect(t, true, models.IsMap())
	Expect(t, "Roadrunner", models.RandomWeightedItem(f)["model"])

	// Decode into a typed value
	var decoded []struct {
		Code string `json:"code"`
	}
	Expect(t, nil, data.Decode(&decoded))
	Expect(t, "INR", decoded[0].Code)

	// Empty map data yields no item instead of panicking
	empty := r.Load("cars/missing.json", fakery.GenericLocale)
	Expect(t, true, empty.RandomWeightedItem(f) == nil)
}

func TestWeightedData(t *testing.T) {
	r, _ := newTestRegistry()
	data := r.Load(fakery.ResourceInternet, fakery.GenericLocale)

	arr, err := data.GetWeightedArray("common_tlds_weighted", ":")
	Expect(t, nil, err)
	Expect(t, 2, len(arr.Items))
	Expect(t, ".com", arr.Items[0].Item)
	Expect(t, 0.75, arr.Items[0].Weight)

	_, err = data.GetWeightedArray("bad_weights", ":")
	Expect(t, true, err != nil)

	_, err = data.GetWeightedArray("no_weights", ":")
	Expect(t, true, err != nil)

	f := newTestFakery(r, "en_US")
	for i := 0; i < 20; i++ {
		tld := f.TLD()
		Expect(t, true, tld == ".com" || tld == ".org", tld)
	}
}

func TestRegister(t *testing.T) {
	r, _ := newTestRegistry()

	// Idempotent
	Expect(t, nil, r.Register(fakery.ResourceSpec{Name: fakery.ResourceCurrency, Kind: fakery.MapResource}))
	Expect(t, nil, r.Register(fakery.ResourceSpec{Name: fakery.ResourceCarModels, Kind: fakery.MapResource}))

	// Conflicts
	err := r.Register(fakery.ResourceSpec{Name: fakery.ResourceCurrency, Kind: fakery.ListResource})
	Expect(t, true, errors.Is(err, fakery.ErrResourceConflict))
	err = r.Register(fakery.ResourceSpec{Name: fakery.ResourceCarModels, Kind: fakery.ListResource})
	Expect(t, true, errors.Is(err, fakery.ErrResourceConflict))

	// Invalid specs
	for _, spec := range []fakery.ResourceSpec{
		{Name: ""},
		{Name: "../escape.json"},
		{Name: "/abs.json"},
		{Name: "bad[pattern.json"},
		{Name: "ok.json", Kind: fakery.ResourceKind(42)},
	} {
		err = r.Register(spec)
		Expect(t, true, errors.Is(err, fakery.ErrInvalidResource), spec)
	}

	// Kind cannot change once data is cached
	r.Load("custom.json", fakery.GenericLocale)
	err = r.Register(fakery.ResourceSpec{Name: "custom.json", Kind: fakery.MapResource})
	Expect(t, true, errors.Is(err, fakery.ErrResourceLoaded))
	Expect(t, nil, r.Register(fakery.ResourceSpec{Name: "custom.json", Kind: fakery.ListResource}))

	// New registration before first use
	Expect(t, nil, r.Register(fakery.ResourceSpec{Name: "fresh.json", Kind: fakery.MapResource}))
	spec, ok := r.Spec("fresh.json")
	Expect(t, true, ok)
	Expect(t, fakery.MapResource, spec.Kind)
}

func TestInvalidKeys(t *testing.T) {
	r, _ := newTestRegistry()

	for _, key := range [][2]string{
		{"", "en_US"},
		{fakery.ResourceNames, ""},
		{"../names.json", "en_US"},
		{fakery.ResourceNames, ".."},
		{fakery.ResourceNames, "en_US/../generic"},
	} {
		data, err := r.Lookup(key[0], key[1])
		Expect(t, true, errors.Is(err, fakery.ErrInvalidResource), key)
		Expect(t, true, data != nil, key)
		Expect(t, 0, data.Len(), key)
	}
}

func TestDataLoaderCompatibility(t *testing.T) {
	// Zero value loader, as used by the previous API
	var loader fakery.DataLoader
	loader.Init(fakery.ResourceJobs)
	Expect(t, fakery.ResourceJobs, loader.FileName())

	f := fakery.New()
	jobs := f.LoadGenericLocale(&loader).Get("jobs")
	Expect(t, true, len(jobs) > 0)
	Expect(t, true, f.LoadGenericLocale(&loader) == fakery.DefaultRegistry().Load(fakery.ResourceJobs, fakery.GenericLocale))
	Expect(t, true, loader.Get(fakery.GenericLocale) == loader.EnsureLoaded(fakery.GenericLocale))

	preloaded := loader.Preload(fakery.GenericLocale)
	Expect(t, true, len(preloaded["jobs"].([]string)) > 0)

	names := fakery.DataLoader{}
	names.Init(fakery.ResourceNames)
	Expect(t, "en_US", f.LoadLocale(&names).Locale())

	// A zero loader follows the registry of the Fakery it is used with
	r, _ := newTestRegistry()
	tf := newTestFakery(r, "en_GB")
	Expect(t, "Smith", tf.LoadLocale(&names).Get("last_name")[0])

	// A registry bound loader, with map data configured through SetIsMap
	custom := r.Loader("cars/acme.json")
	custom.SetIsMap(fakery.GenericLocale)
	Expect(t, true, custom.EnsureLoaded(fakery.GenericLocale).IsMap())

	other := r.Loader("other_map.json")
	other.SetIsMap(fakery.GenericLocale)
	spec, _ := r.Spec("other_map.json")
	Expect(t, fakery.MapResource, spec.Kind)

	// Uninitialised LocaleData reports an error on Load
	var empty fakery.LocaleData
	Expect(t, true, empty.Load() != nil)
}

func TestSeededDeterminism(t *testing.T) {
	a := fakery.NewFromSeed(42)
	b := fakery.NewFromSeed(42)
	for i := 0; i < 20; i++ {
		Expect(t, a.Name(), b.Name())
		Expect(t, a.Currency().Code, b.Currency().Code)
		Expect(t, a.CarModel(), b.CarModel())
	}
}

func TestRandomStringExclDoesNotMutateInput(t *testing.T) {
	f := fakery.New()
	items := []string{"a", "b", "c"}
	for i := 0; i < 20; i++ {
		Expect(t, true, f.RandomStringExcl(items, "b") != "b")
	}
	Expect(t, "a", items[0])
	Expect(t, "b", items[1])
	Expect(t, "c", items[2])
	Expect(t, "", f.RandomStringExcl(nil, "b"))
	Expect(t, "", f.RandomString(nil))
}

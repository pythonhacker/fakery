package tests

import (
	"errors"
	"fakery"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

const testSource = "things.json"

// newTestRegistry builds an isolated registry over a temporary data root.
func newTestRegistry(t *testing.T) (*fakery.Registry, string) {
	t.Helper()
	root := t.TempDir()
	return fakery.NewRegistryWithRoot(root), root
}

func TestRegistryRepeatedAccessReusesCache(t *testing.T) {
	r, root := newTestRegistry(t)
	writeData(t, root, "en_US", testSource, `{"items": ["one", "two"]}`)
	r.Register(testSource)

	first := r.Load(testSource, "en_US")
	Expect(t, nil, first.Err())
	Expect(t, 2, len(first.Get("items")))

	// Change and then delete the file: cached data must be served
	// unchanged, which proves the file is not read again.
	writeData(t, root, "en_US", testSource, `{"items": ["changed"]}`)
	second := r.Load(testSource, "en_US")
	Expect(t, true, first == second)
	Expect(t, 2, len(second.Get("items")))

	if err := os.Remove(filepath.Join(root, "en_US", testSource)); err != nil {
		t.Fatal(err)
	}
	third := r.Load(testSource, "en_US")
	Expect(t, true, first == third)
	Expect(t, nil, third.Err())
	Expect(t, "one", third.Get("items")[0])
}

func TestRegistryRegisterIsIdempotent(t *testing.T) {
	r, _ := newTestRegistry(t)
	a := r.Register(testSource)
	b := r.Register(testSource)
	Expect(t, true, a == b)

	l, ok := r.Loader(testSource)
	Expect(t, true, ok)
	Expect(t, true, l == a)
	Expect(t, 1, len(r.Sources()))
}

func TestRegistryLocalesIsolated(t *testing.T) {
	r, root := newTestRegistry(t)
	writeData(t, root, "en_US", testSource, `{"items": ["us"]}`)
	writeData(t, root, "en_GB", testSource, `{"items": ["gb"]}`)
	r.Register(testSource)

	us := r.Load(testSource, "en_US")
	gb := r.Load(testSource, "en_GB")
	Expect(t, true, us != gb)
	Expect(t, "us", us.Get("items")[0])
	Expect(t, "gb", gb.Get("items")[0])

	// A locale without the file fails on its own; others are unaffected.
	fr := r.Load(testSource, "fr_FR")
	Expect(t, true, errors.Is(fr.Err(), os.ErrNotExist))
	Expect(t, "us", r.Load(testSource, "en_US").Get("items")[0])
}

// Two registries never share cached data, even for the same names.
func TestRegistriesAreIsolated(t *testing.T) {
	r1, root1 := newTestRegistry(t)
	r2, root2 := newTestRegistry(t)
	writeData(t, root1, "xx", testSource, `{"items": ["r1"]}`)
	writeData(t, root2, "xx", testSource, `{"items": ["r2"]}`)
	r1.Register(testSource)
	r2.Register(testSource)

	Expect(t, "r1", r1.Load(testSource, "xx").Get("items")[0])
	Expect(t, "r2", r2.Load(testSource, "xx").Get("items")[0])
}

func TestRegistryUnregisteredSource(t *testing.T) {
	r, root := newTestRegistry(t)
	writeData(t, root, "xx", testSource, `{"items": ["x"]}`)

	data := r.Load(testSource, "xx")
	Expect(t, true, data != nil)
	Expect(t, true, errors.Is(data.Err(), fakery.ErrSourceNotRegistered))
	Expect(t, true, data.Loaded())
	Expect(t, true, data.Get("items") == nil)

	_, ok := r.Loader(testSource)
	Expect(t, false, ok)
}

func TestRegistryMissingResource(t *testing.T) {
	r, _ := newTestRegistry(t)
	r.Register("absent.json")
	data := r.Load("absent.json", "xx")
	Expect(t, true, errors.Is(data.Err(), os.ErrNotExist))
	Expect(t, 0, data.Len())
	Expect(t, true, data.Get("key") == nil)
}

func TestRegistryMalformedJSON(t *testing.T) {
	r, root := newTestRegistry(t)
	writeData(t, root, "xx", "broken.json", `{"items": ["unterminated"`)
	writeData(t, root, "xx", "wrongshape.json", `{"items": "not-an-array"}`)
	writeData(t, root, "xx", "brokenmap.json", `[{"a": 1}]`)
	r.Register("broken.json")
	r.Register("wrongshape.json")
	r.Register("brokenmap.json", fakery.WithMapLocales("xx"))

	for _, name := range []string{"broken.json", "wrongshape.json", "brokenmap.json"} {
		data := r.Load(name, "xx")
		Expect(t, true, data.Err() != nil, name)
		Expect(t, true, data.Loaded(), name)
		Expect(t, 0, data.Len(), name)
	}

	// Failures are cached as well: fixing the file does not reload it.
	writeData(t, root, "xx", "broken.json", `{"items": ["fixed"]}`)
	Expect(t, true, r.Load("broken.json", "xx").Err() != nil)
}

func TestRegistryMapBasedData(t *testing.T) {
	r, root := newTestRegistry(t)
	records := `[{"code": "AAA", "name": "Alpha"}, {"code": "BBB", "name": "Beta"}]`
	writeData(t, root, "generic", "records.json", records)
	writeData(t, root, "en_US", "records.json", `{"codes": ["AAA"]}`)
	r.Register("records.json", fakery.WithMapLocales("generic"))

	generic := r.Load("records.json", "generic")
	Expect(t, nil, generic.Err())
	Expect(t, true, generic.IsMap())
	Expect(t, 2, generic.Len())

	f := fakery.NewFromSeed(1)
	for i := 0; i < 20; i++ {
		item := generic.RandomWeightedItem(f)
		Expect(t, true, item["code"] == "AAA" || item["code"] == "BBB")
	}

	var decoded []map[string]string
	Expect(t, nil, generic.Decode(&decoded))
	Expect(t, 2, len(decoded))
	Expect(t, "Beta", decoded[1]["name"])

	// The map mode is per locale: en_US is plain array data.
	us := r.Load("records.json", "en_US")
	Expect(t, nil, us.Err())
	Expect(t, false, us.IsMap())
	Expect(t, "AAA", us.Get("codes")[0])
}

func TestRegistryPrefixFamily(t *testing.T) {
	r, root := newTestRegistry(t)
	writeData(t, root, "generic", "cars/acme.json", `[{"model": "Roadrunner"}]`)
	writeData(t, root, "generic", "cars/special/x.json", `{"k": ["v"]}`)
	r.RegisterPrefix("cars/", fakery.WithMapLocales("generic"))
	// Longer prefixes win over shorter ones.
	r.RegisterPrefix("cars/special/")

	acme := r.Load("cars/acme.json", "generic")
	Expect(t, nil, acme.Err())
	Expect(t, "Roadrunner", acme.RandomWeightedItem(fakery.New())["model"])
	Expect(t, true, acme == r.Load("cars/acme.json", "generic"))

	special := r.Load("cars/special/x.json", "generic")
	Expect(t, nil, special.Err())
	Expect(t, "v", special.Get("k")[0])

	sources := r.Sources()
	Expect(t, true, sort.StringsAreSorted(sources))
	Expect(t, 2, len(sources))

	// A family member whose file is missing fails gracefully.
	missing := r.Load("cars/nope.json", "generic")
	Expect(t, true, errors.Is(missing.Err(), os.ErrNotExist))
	Expect(t, true, missing.RandomWeightedItem(fakery.New()) == nil)
}

func TestRegistryConcurrentSameResource(t *testing.T) {
	r, root := newTestRegistry(t)
	writeData(t, root, "en_US", testSource, `{"items": ["a", "b", "c"]}`)
	r.Register(testSource)

	results := make([]*fakery.LocaleData, raceWorkers)
	runTogether(raceWorkers, func(i int) {
		results[i] = r.Load(testSource, "en_US")
	})
	for _, data := range results {
		Expect(t, true, data == results[0], "all goroutines share one instance")
	}
	Expect(t, 3, len(results[0].Get("items")))
}

func TestRegistryConcurrentDifferentLocales(t *testing.T) {
	r, root := newTestRegistry(t)
	const locales = 8
	for i := 0; i < locales; i++ {
		writeData(t, root, fmt.Sprintf("l%d", i), testSource,
			fmt.Sprintf(`{"items": ["value-%d"]}`, i))
	}
	r.Register(testSource)

	results := make([]*fakery.LocaleData, raceWorkers)
	runTogether(raceWorkers, func(i int) {
		results[i] = r.Load(testSource, fmt.Sprintf("l%d", i%locales))
	})

	for i, data := range results {
		Expect(t, nil, data.Err())
		Expect(t, fmt.Sprintf("value-%d", i%locales), data.Get("items")[0])
		Expect(t, true, data == results[i%locales])
	}
}

func TestRegistryConcurrentRegisterAndFamilyLoad(t *testing.T) {
	r, root := newTestRegistry(t)
	for i := 0; i < 4; i++ {
		writeData(t, root, "generic", fmt.Sprintf("fam/%d.json", i), `[{"k": "v"}]`)
	}
	r.RegisterPrefix("fam/", fakery.WithMapLocales("generic"))

	runTogether(raceWorkers, func(i int) {
		switch i % 3 {
		case 0:
			r.Register(fmt.Sprintf("extra-%d.json", i%5))
		case 1:
			data := r.Load(fmt.Sprintf("fam/%d.json", i%4), "generic")
			if data.Len() != 1 {
				t.Errorf("unexpected length %d", data.Len())
			}
		default:
			_ = r.Sources()
		}
	})
	Expect(t, 9, len(r.Sources())) // 5 extra + 4 family members
}

type thingsData struct {
	Items []string `json:"items"`
}

func TestDataView(t *testing.T) {
	r, root := newTestRegistry(t)
	writeData(t, root, "generic", testSource, `{"items": ["x", "y"]}`)
	r.Register(testSource)

	view := fakery.NewDataView[thingsData](r, testSource, "generic")
	results := make([]*thingsData, raceWorkers)
	runTogether(raceWorkers, func(i int) {
		results[i] = view.Get()
	})
	for _, v := range results {
		Expect(t, true, v == results[0])
	}
	Expect(t, 2, len(results[0].Items))
}

func TestDataViewFailureYieldsZeroValue(t *testing.T) {
	r, _ := newTestRegistry(t)
	view := fakery.NewDataView[thingsData](r, "unknown.json", "generic")
	v := view.Get()
	Expect(t, true, v != nil)
	Expect(t, 0, len(v.Items))
}

func TestDefaultRegistryBuiltinSources(t *testing.T) {
	r := fakery.DefaultRegistry()
	for _, name := range []string{
		fakery.SourceAddress, fakery.SourceBeer, fakery.SourceBook,
		fakery.SourceCar, fakery.SourceColor, fakery.SourceCurrency,
		fakery.SourceInternet, fakery.SourceJobs, fakery.SourceWine,
		fakery.SourceWords,
	} {
		_, ok := r.Loader(name)
		Expect(t, true, ok, name)
		Expect(t, nil, r.Load(name, fakery.GenericLocale).Err(), name)
	}
	Expect(t, nil, r.Load(fakery.SourceNames, "en_US").Err())
	Expect(t, nil, r.Load(fakery.SourceNames, "en_GB").Err())
	Expect(t, nil, r.Load(fakery.SourceAddress, "en_US").Err())
	Expect(t, true, r.Load(fakery.SourceCurrency, fakery.GenericLocale).IsMap())
}

// A custom registry with built-ins registered behaves like the default.
func TestRegisterBuiltinSourcesOnCustomRegistry(t *testing.T) {
	r := fakery.RegisterBuiltinSources(fakery.NewRegistry())
	Expect(t, true, r != fakery.DefaultRegistry())
	data := r.Load(fakery.CarModelsSource("Toyota"), fakery.GenericLocale)
	Expect(t, nil, data.Err())
	Expect(t, true, data.IsMap())
	Expect(t, true, len(data.RandomWeightedItem(fakery.New())["model"]) > 0)
}

func TestCarModelsSource(t *testing.T) {
	Expect(t, "cars/mercedesbenz.json", fakery.CarModelsSource("Mercedes-Benz"))
	Expect(t, "cars/landrover.json", fakery.CarModelsSource("Land Rover"))
}

// Every maker listed in car.json has a model file.
func TestAllCarMakersHaveModels(t *testing.T) {
	r := fakery.DefaultRegistry()
	makers := r.Load(fakery.SourceCar, fakery.GenericLocale).Get("car_makers")
	Expect(t, true, len(makers) > 0)
	for _, maker := range makers {
		data := r.Load(fakery.CarModelsSource(maker), fakery.GenericLocale)
		Expect(t, nil, data.Err(), maker)
		Expect(t, true, data.Len() > 0, maker)
	}
}

// Generators share the default registry concurrently across locales.
// Each goroutine owns its Fakery, matching the documented contract.
func TestGeneratorsConcurrentAcrossLocales(t *testing.T) {
	locales := []string{"en_US", "en_GB"}
	runTogether(raceWorkers, func(i int) {
		f := fakery.NewFromLocale(locales[i%len(locales)])
		checks := map[string]string{
			"name":     f.Name(),
			"first":    f.FirstName(),
			"currency": f.CurrencyCode(),
			"car":      f.CarModel(),
			"beer":     f.Beer().Name,
			"wine":     f.Wine().Name,
			"color":    f.ColorName(),
			"book":     f.Book().Title,
			"tld":      f.TLD(),
			"job":      f.Job().Title,
			"adverb":   f.Adverb(),
			"country":  f.Country(),
		}
		for k, v := range checks {
			if v == "" {
				t.Errorf("locale %s: empty %s", locales[i%len(locales)], k)
			}
		}
		if i%len(locales) == 0 {
			if f.State() == "" {
				t.Errorf("empty state for en_US")
			}
		}
	})
}

// The public wrapper functions keep working with caller-owned loaders.
func TestFakeryLoadWrappers(t *testing.T) {
	loader := fakery.NewDataLoader(fakery.SourceNames)
	f := fakery.NewFromLocale("en_GB")
	data := f.LoadLocale(loader)
	Expect(t, "en_GB", data.Locale())
	Expect(t, nil, data.Err())

	generic := f.LoadGenericLocale(fakery.NewDataLoader(fakery.SourceWords))
	Expect(t, fakery.GenericLocale, generic.Locale())
	Expect(t, true, len(generic.Get("adjectives")) > 0)
}

package tests

import (
	"encoding/json"
	"errors"
	"fakery"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func writeLocaleFile(t *testing.T, baseDir, locale, relFile, content string) string {
	t.Helper()
	fullPath := filepath.Join(baseDir, locale, filepath.FromSlash(relFile))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(fullPath), err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", fullPath, err)
	}
	return fullPath
}

func TestDefaultRegistryBuiltInRegistration(t *testing.T) {
	reg := fakery.DefaultRegistry()
	Expect(t, true, reg != nil)

	expected := []string{
		fakery.ResourceAddress,
		fakery.ResourceBeer,
		fakery.ResourceBook,
		fakery.ResourceCar,
		fakery.ResourceColor,
		fakery.ResourceCurrency,
		fakery.ResourceInternet,
		fakery.ResourceJobs,
		fakery.ResourceNames,
		fakery.ResourceWine,
		fakery.ResourceWords,
	}

	registered := reg.RegisteredResources()
	for _, res := range expected {
		Expect(t, true, reg.Has(res), res)
		Expect(t, true, slices.Contains(registered, res), res)
	}

	// Map-based static resource configuration.
	currencyData := reg.Loader(fakery.ResourceCurrency).Get(fakery.GenericLocale)
	Expect(t, true, currencyData.IsMap())

	// Dynamic map-based rule for car model resources ("cars/<maker>.json").
	audiLoader := reg.Loader("cars/audi.json")
	Expect(t, true, audiLoader.Get(fakery.GenericLocale).IsMap())
}

func TestLazyLoadingAndCachedReuse(t *testing.T) {
	baseDir := t.TempDir()
	filePath := writeLocaleFile(t, baseDir, "en_US", "names.json", `{
		"first_name_male": ["Arthur"],
		"first_name_female": ["Guinevere"],
		"last_name": ["Pendragon"]
	}`)

	reg := fakery.NewDefaultLoaderRegistry(baseDir)
	loader := reg.Loader(fakery.ResourceNames)

	// Registration alone must not perform disk I/O.
	entry := loader.Get("en_US")
	Expect(t, false, entry.IsLoaded())
	Expect(t, int64(0), entry.LoadCount())

	// First load reads from disk once.
	loaded1 := reg.LoadLocale(fakery.ResourceNames, "en_US")
	Expect(t, true, loaded1 == entry)
	Expect(t, true, loaded1.IsLoaded())
	Expect(t, nil, loaded1.Err())
	Expect(t, int64(1), loaded1.LoadCount())
	Expect(t, "Arthur", loaded1.Get("first_name_male")[0])

	// Remove the underlying file on disk: repeated accesses must reuse cached memory
	// without re-reading disk.
	if err := os.Remove(filePath); err != nil {
		t.Fatalf("remove temp locale file: %v", err)
	}

	loaded2 := reg.LoadLocale(fakery.ResourceNames, "en_US")
	loaded3 := loader.EnsureLoaded("en_US")
	Expect(t, true, loaded1 == loaded2)
	Expect(t, true, loaded1 == loaded3)
	Expect(t, loaded1.ID(), loaded2.ID())
	Expect(t, int64(1), loaded2.LoadCount())
	Expect(t, "Pendragon", loaded2.Get("last_name")[0])

	// Mutating returned slice must not corrupt cached locale data.
	got := loaded2.Get("first_name_male")
	got[0] = "Mutated"
	Expect(t, "Arthur", loaded2.Get("first_name_male")[0])
}

func TestCarModelDynamicLoaderReusesCache(t *testing.T) {
	baseDir := t.TempDir()
	writeLocaleFile(t, baseDir, "generic", "car.json", `{
		"car_makers": ["Toyota"],
		"car_categories": ["Sedan"],
		"car_series": ["LE"],
		"car_category_adjectives": ["Compact"],
		"car_fuel_types": ["Hybrid"]
	}`)
	modelPath := writeLocaleFile(t, baseDir, "generic", "cars/toyota.json", `[
		{"model": "Camry", "type": "Sedan"}
	]`)

	reg := fakery.NewDefaultLoaderRegistry(baseDir)
	f := fakery.NewWithRegistry("en_US", reg)

	Expect(t, "Camry", f.CarModel())

	if err := os.Remove(modelPath); err != nil {
		t.Fatalf("remove car model file: %v", err)
	}

	// Subsequent calls must reuse the central registry's cached cars/toyota.json loader.
	for i := 0; i < 10; i++ {
		Expect(t, "Camry", f.CarModel())
	}

	toyotaData := reg.Loader("cars/toyota.json").Get(fakery.GenericLocale)
	Expect(t, int64(1), toyotaData.LoadCount())
}

func TestLocaleIsolationAcrossLocales(t *testing.T) {
	baseDir := t.TempDir()
	writeLocaleFile(t, baseDir, "en_US", "names.json", `{
		"first_name_male": ["John"],
		"first_name_female": ["Jane"],
		"last_name": ["Smith"]
	}`)
	writeLocaleFile(t, baseDir, "en_GB", "names.json", `{
		"first_name_male": ["Nigel"],
		"first_name_female": ["Imogen"],
		"last_name": ["Featherstone"]
	}`)
	writeLocaleFile(t, baseDir, "fr_FR", "names.json", `{
		"first_name_male": ["Jean"],
		"first_name_female": ["Camille"],
		"last_name": ["Dupont"]
	}`)

	reg := fakery.NewDefaultLoaderRegistry(baseDir)
	fUS := fakery.NewWithRegistry("en_US", reg)
	fGB := fakery.NewWithRegistry("en_GB", reg)
	fFR := fakery.NewWithRegistry("fr_FR", reg)

	for i := 0; i < 20; i++ {
		Expect(t, "Smith", fUS.LastName())
		Expect(t, "Featherstone", fGB.LastName())
		Expect(t, "Dupont", fFR.LastName())
	}

	usData := reg.LoadLocale(fakery.ResourceNames, "en_US")
	gbData := reg.LoadLocale(fakery.ResourceNames, "en_GB")
	frData := reg.LoadLocale(fakery.ResourceNames, "fr_FR")

	Expect(t, true, usData != gbData)
	Expect(t, true, gbData != frData)
	Expect(t, true, usData.ID() != gbData.ID())
	Expect(t, "en_US", usData.Locale())
	Expect(t, "en_GB", gbData.Locale())
	Expect(t, "fr_FR", frData.Locale())

	// Switching locale on an existing Fakery instance switches isolated datasets cleanly.
	fUS.SetLocale("en_GB")
	Expect(t, "Featherstone", fUS.LastName())
	fUS.SetLocale("fr_FR")
	Expect(t, "Dupont", fUS.LastName())

	// Built-in en_US vs en_GB datasets in repository also remain distinct.
	realUS := fakery.DefaultRegistry().LoadLocale(fakery.ResourceNames, "en_US")
	realGB := fakery.DefaultRegistry().LoadLocale(fakery.ResourceNames, "en_GB")
	Expect(t, true, realUS != realGB)
	Expect(t, true, slices.Contains(realGB.Get("first_name_male"), "Aadam"))
}

func TestLocaleFallbackBehaviour(t *testing.T) {
	baseDir := t.TempDir()
	writeLocaleFile(t, baseDir, "en_US", "names.json", `{
		"first_name_male": ["DefaultMale"],
		"first_name_female": ["DefaultFemale"],
		"last_name": ["DefaultLast"]
	}`)
	writeLocaleFile(t, baseDir, "fr", "names.json", `{
		"first_name_male": ["BaseFrenchMale"],
		"first_name_female": ["BaseFrenchFemale"],
		"last_name": ["BaseFrenchLast"]
	}`)
	writeLocaleFile(t, baseDir, "generic", "only_generic.json", `{
		"items": ["GenericValue"]
	}`)

	reg := fakery.NewDefaultLoaderRegistry(baseDir)

	// Regional locale "fr_CA" falls back to language base "fr".
	frCA := fakery.NewWithRegistry("fr_CA", reg)
	Expect(t, "BaseFrenchLast", frCA.LastName())
	Expect(t, "fr", reg.LoadLocale(fakery.ResourceNames, "fr_CA").Locale())

	// Unknown locale "de_DE" falls back to DefaultLocale ("en_US").
	deDE := fakery.NewWithRegistry("de_DE", reg)
	Expect(t, "DefaultLast", deDE.LastName())
	Expect(t, "en_US", reg.LoadLocale(fakery.ResourceNames, "de_DE").Locale())

	// Empty locale falls back to DefaultLocale ("en_US").
	emptyLoc := fakery.NewWithRegistry("", reg)
	Expect(t, "DefaultLast", emptyLoc.LastName())

	// Resource present only in generic locale resolves via generic fallback.
	genResolved := reg.LoadLocale("only_generic.json", "de_DE")
	Expect(t, fakery.GenericLocale, genResolved.Locale())
	Expect(t, "GenericValue", genResolved.Get("items")[0])

	// Real en_GB Address falls back to en_US for states while preserving GB postal code & country.
	gbAddr := fakery.NewFromLocale("en_GB").Address()
	Expect(t, true, gbAddr != nil)
	Expect(t, true, len(gbAddr.State) > 0)
	Expect(t, true, len(gbAddr.PostalCode) > 0)
	Expect(t, "", gbAddr.ZipCode)
	Expect(t, "United Kingdom", gbAddr.Country)
}

func TestMissingResources(t *testing.T) {
	baseDir := t.TempDir()
	reg := fakery.NewDefaultLoaderRegistry(baseDir)
	f := fakery.NewWithRegistry("en_US", reg)

	// Missing resource via ResolveLocale / LoadLocale.
	missing := reg.LoadLocale("does_not_exist.json", "en_US")
	Expect(t, true, missing != nil)
	Expect(t, true, missing.IsLoaded())
	Expect(t, true, errors.Is(missing.Err(), os.ErrNotExist))
	Expect(t, int64(1), missing.LoadCount())
	Expect(t, true, missing.Get("missing_key") == nil)
	Expect(t, true, missing.Entries() == nil)
	Expect(t, true, missing.RandomWeightedItem(f) == nil)

	// Repeated access to missing resource must not re-attempt disk reads.
	missingAgain := reg.LoadLocale("does_not_exist.json", "en_US")
	Expect(t, true, missing == missingAgain)
	Expect(t, int64(1), missingAgain.LoadCount())

	// Path traversal outside baseDir is rejected safely.
	traversal := reg.Loader("../../etc/passwd").EnsureLoaded("generic")
	Expect(t, true, errors.Is(traversal.Err(), os.ErrNotExist))

	// Uninitialized LocaleData returns an initialization error.
	var uninit fakery.LocaleData
	Expect(t, true, uninit.Load() != nil)
}

func TestMalformedJSONHandling(t *testing.T) {
	baseDir := t.TempDir()

	// Default locale has valid JSON, while bad_LOC has broken JSON syntax.
	writeLocaleFile(t, baseDir, "en_US", "names.json", `{
		"last_name": ["ValidDefault"]
	}`)
	writeLocaleFile(t, baseDir, "bad_LOC", "names.json", `{"last_name": ["Unclosed"`)
	writeLocaleFile(t, baseDir, "generic", "currency.json", `[{"code": "USD"`)
	writeLocaleFile(t, baseDir, "type_MISMATCH", "names.json", `[{"not": "a slice map"}]`)
	writeLocaleFile(t, baseDir, "generic", "cars/broken.json", `{"not": ["a map list"]}`)

	reg := fakery.NewDefaultLoaderRegistry(baseDir)
	f := fakery.NewWithRegistry("bad_LOC", reg)

	// Syntax error must be surfaced and NOT masked by falling back to en_US.
	badNames := reg.LoadLocale(fakery.ResourceNames, "bad_LOC")
	Expect(t, "bad_LOC", badNames.Locale())
	Expect(t, true, badNames.Err() != nil)
	var syntaxErr *json.SyntaxError
	Expect(t, true, errors.As(badNames.Err(), &syntaxErr))
	Expect(t, true, badNames.Get("last_name") == nil)
	Expect(t, "", f.LastName())

	// Syntax error in map-based resource.
	badCurrency := reg.LoadGeneric(fakery.ResourceCurrency)
	Expect(t, true, badCurrency.Err() != nil)
	Expect(t, true, errors.As(badCurrency.Err(), &syntaxErr))
	Expect(t, true, badCurrency.Entries() == nil)
	Expect(t, "", f.CurrencyCode())

	// Shape mismatch: array of maps given to key->[]string loader.
	mismatchKeyMap := reg.LoadLocale(fakery.ResourceNames, "type_MISMATCH")
	Expect(t, true, mismatchKeyMap.Err() != nil)
	Expect(t, true, mismatchKeyMap.Get("not") == nil)

	// Shape mismatch: object given to map-slice loader.
	mismatchMapSlice := reg.LoadGeneric("cars/broken.json")
	Expect(t, true, mismatchMapSlice.Err() != nil)
	Expect(t, true, mismatchMapSlice.Entries() == nil)
}

func TestMapBasedDataAndWeightedArrays(t *testing.T) {
	baseDir := t.TempDir()
	writeLocaleFile(t, baseDir, "generic", "currency.json", `[
		{"country": "Japan", "code": "JPY", "currency": "Japanese Yen"},
		{"country": "Switzerland", "code": "CHF", "currency": "Swiss Franc"}
	]`)
	writeLocaleFile(t, baseDir, "generic", "internet.json", `{
		"valid_weighted": [".com:0.75", ".org:0.25"],
		"bad_sep": ["com_0.5", "org_0.5"],
		"bad_float": [".com:abc", ".org:0.5"],
		"bad_sum": [".com:0.20", ".org:0.20"]
	}`)

	reg := fakery.NewDefaultLoaderRegistry(baseDir)
	f := fakery.NewWithRegistry("en_US", reg)

	// Map-based decoding & defensive copy checks.
	currData := reg.LoadGeneric(fakery.ResourceCurrency)
	Expect(t, nil, currData.Err())
	Expect(t, true, currData.IsMap())
	entries := currData.Entries()
	Expect(t, 2, len(entries))
	entries[0]["code"] = "HACKED"
	Expect(t, "JPY", currData.Entries()[0]["code"])

	for i := 0; i < 10; i++ {
		c := f.Currency()
		Expect(t, true, c.Code == "JPY" || c.Code == "CHF")
		if c.Code == "JPY" {
			Expect(t, "Japan", c.Country)
			Expect(t, "Japanese Yen", c.Name)
		} else {
			Expect(t, "Switzerland", c.Country)
			Expect(t, "Swiss Franc", c.Name)
		}
		picked := currData.RandomWeightedItem(f)
		picked["code"] = "MUTATED"
	}
	Expect(t, "JPY", currData.Entries()[0]["code"])

	// Weighted array decoding & error cases.
	netData := reg.LoadGeneric(fakery.ResourceInternet)
	wa, err := netData.GetWeightedArray("valid_weighted", ":")
	Expect(t, nil, err)
	Expect(t, 2, len(wa.Items))
	item, err := f.RandomWeightedItem(wa)
	Expect(t, nil, err)
	Expect(t, true, item == ".com" || item == ".org")

	// Mutating returned WeightedArray must not corrupt cached WeightedArray.
	wa.Items[0].Weight = 99.0
	wa2, err := netData.GetWeightedArray("valid_weighted", ":")
	Expect(t, nil, err)
	ok, _ := wa2.Validate()
	Expect(t, true, ok)

	_, err = netData.GetWeightedArray("bad_sep", ":")
	Expect(t, true, err != nil)

	_, err = netData.GetWeightedArray("bad_float", ":")
	Expect(t, true, err != nil)

	_, err = netData.GetWeightedArray("bad_sum", ":")
	Expect(t, true, err != nil)

	_, err = netData.GetWeightedArray("nonexistent_key", ":")
	Expect(t, true, err != nil)

	// DataLoader.Preload + ConvertMapToStruct compatibility.
	preloaded := reg.Loader(fakery.ResourceInternet).Preload(fakery.GenericLocale)
	Expect(t, true, len(preloaded) > 0)
	var target struct {
		ValidWeighted []string `json:"valid_weighted"`
	}
	Expect(t, nil, fakery.ConvertMapToStruct(preloaded, &target))
	Expect(t, 2, len(target.ValidWeighted))
}

func TestWordGenerators(t *testing.T) {
	f := fakery.New()
	Expect(t, true, len(f.Adjective()) > 0)
	Expect(t, true, len(f.AdjectivePositive()) > 0)
	Expect(t, true, len(f.AdjectiveNegative()) > 0)
	Expect(t, true, len(f.Adverb()) > 0)
}

func TestConcurrentSingleFlightAndMultiLocaleIsolation(t *testing.T) {
	baseDir := t.TempDir()
	const localeCount = 16

	for i := 0; i < localeCount; i++ {
		loc := fmt.Sprintf("loc_%02d", i)
		writeLocaleFile(t, baseDir, loc, "names.json", fmt.Sprintf(`{
			"first_name_male": ["Male_%02d"],
			"first_name_female": ["Female_%02d"],
			"last_name": ["Last_%02d"]
		}`, i, i, i))
	}
	writeLocaleFile(t, baseDir, "generic", "cars/tesla.json", `[
		{"model": "Model S", "type": "Sedan"},
		{"model": "Model 3", "type": "Sedan"}
	]`)

	reg := fakery.NewDefaultLoaderRegistry(baseDir)

	// 1. Concurrent cold requests for the exact same (resource, locale) must load disk once.
	results := make([]*fakery.LocaleData, raceWorkers)
	runTogether(raceWorkers, func(i int) {
		results[i] = reg.LoadLocale(fakery.ResourceNames, "loc_00")
		if got := results[i].Get("last_name"); len(got) != 1 || got[0] != "Last_00" {
			t.Errorf("worker %d got unexpected data: %v", i, got)
		}
	})

	for i := 1; i < raceWorkers; i++ {
		Expect(t, true, results[0] == results[i])
		Expect(t, results[0].ID(), results[i].ID())
	}
	Expect(t, int64(1), results[0].LoadCount())

	// 2. Concurrent requests across many distinct locales + dynamic map resources.
	runTogether(raceWorkers, func(i int) {
		idx := i % localeCount
		loc := fmt.Sprintf("loc_%02d", idx)
		expectedLast := fmt.Sprintf("Last_%02d", idx)

		f := fakery.NewWithRegistry(loc, reg)
		if got := f.LastName(); got != expectedLast {
			t.Errorf("locale %s isolation violated: expected %s, got %s", loc, expectedLast, got)
		}

		carData := reg.LoadGeneric("cars/tesla.json")
		picked := carData.RandomWeightedItem(f)
		if !strings.HasPrefix(picked["model"], "Model ") {
			t.Errorf("unexpected car model: %v", picked)
		}
	})

	for i := 0; i < localeCount; i++ {
		loc := fmt.Sprintf("loc_%02d", i)
		Expect(t, int64(1), reg.Loader(fakery.ResourceNames).Get(loc).LoadCount())
	}
	Expect(t, int64(1), reg.Loader("cars/tesla.json").Get(fakery.GenericLocale).LoadCount())
}

func TestConcurrentAllGeneratorsAndSharedFakery(t *testing.T) {
	shared := fakery.New()
	locales := []string{"en_US", "en_GB", "fr_FR"}

	runTogether(raceWorkers, func(i int) {
		// Exercise per-goroutine Fakery instances across locales.
		f := fakery.NewFromLocale(locales[i%len(locales)])
		_ = f.Person()
		_ = f.Address()
		_ = f.Beer()
		_ = f.Book()
		_ = f.Car()
		_ = f.Color()
		_ = f.ColorName()
		_ = f.Currency()
		_ = f.Email()
		_ = f.UserName()
		_ = f.TLD()
		_ = f.Adjective()
		_ = f.Adverb()
		_ = f.Job()
		_ = f.Wine()

		// Also exercise concurrent calls on a single shared Fakery instance.
		shared.SetLocale(locales[(i+1)%len(locales)])
		_ = shared.Name()
		_, _ = shared.CarMakeAndModel()
		_ = shared.CurrencyCode()
	})
}

func TestBackwardCompatibleAPIsAndDeterministicSeed(t *testing.T) {
	// Standalone DataLoader constructors and FileName metadata.
	namesLoader := fakery.NewDataLoader(fakery.ResourceNames)
	Expect(t, fakery.ResourceNames, namesLoader.FileName())

	currencyLoader := fakery.NewMapDataLoader(fakery.ResourceCurrency)
	Expect(t, fakery.ResourceCurrency, currencyLoader.FileName())
	Expect(t, true, currencyLoader.Get(fakery.GenericLocale).IsMap())
	Expect(t, fakery.ResourceCurrency, currencyLoader.Get(fakery.GenericLocale).FileName())

	// Zero-value DataLoader works safely without panicking even before Init.
	var zeroLoader fakery.DataLoader
	zeroLoader.SetIsMap("en_US")
	zeroData := zeroLoader.Get("en_US")
	Expect(t, true, zeroData != nil)
	Expect(t, true, zeroData.IsMap())
	Expect(t, true, zeroData.Load() != nil)

	// Nil LocaleData receiver safety.
	var nilData *fakery.LocaleData
	Expect(t, "", nilData.Locale())
	Expect(t, "", nilData.FileName())
	Expect(t, false, nilData.IsMap())
	Expect(t, false, nilData.IsLoaded())
	Expect(t, int64(0), nilData.LoadCount())
	Expect(t, nil, nilData.Err())
	Expect(t, true, nilData.Get("k") == nil)
	Expect(t, true, nilData.Entries() == nil)
	Expect(t, true, nilData.RandomWeightedItem(fakery.New()) == nil)

	// Fakery.LoadLocale and Fakery.LoadGenericLocale compatibility wrappers.
	f := fakery.NewFromSeed(42)
	Expect(t, true, f.LoadLocale(nil) == nil)
	Expect(t, true, f.LoadGenericLocale(nil) == nil)
	Expect(t, true, len(f.LoadLocale(namesLoader).Get("first_name_male")) > 0)
	Expect(t, true, len(f.LoadGenericLocale(currencyLoader).Entries()) > 0)

	// Registry / SetRegistry switching and idempotent Register.
	baseDir := t.TempDir()
	writeLocaleFile(t, baseDir, "generic", "custom.json", `[{"k": "v"}]`)
	customReg := fakery.NewLoaderRegistry(baseDir)
	customReg.RegisterRule(nil)
	l1 := customReg.Loader("custom.json")
	l2 := customReg.Register(fakery.ResourceConfig{
		FileName:   "custom.json",
		IsMap:      true,
		MapLocales: []string{fakery.GenericLocale},
	})
	Expect(t, true, l1 == l2)
	Expect(t, "v", customReg.LoadGeneric("custom.json").Entries()[0]["k"])

	f.SetRegistry(customReg)
	Expect(t, true, f.Registry() == customReg)
	f.SetRegistry(nil)
	Expect(t, true, f.Registry() == fakery.DefaultRegistry())

	// Deterministic output with NewFromSeed across lazy-loaded generators.
	s1 := fakery.NewFromSeed(12345)
	s2 := fakery.NewFromSeed(12345)
	Expect(t, s1.Name(), s2.Name())
	Expect(t, s1.BeerName(), s2.BeerName())
	Expect(t, s1.WineName(), s2.WineName())
	Expect(t, s1.CarMake(), s2.CarMake())
	Expect(t, s1.ColorName(), s2.ColorName())
	Expect(t, s1.CurrencyCode(), s2.CurrencyCode())
	Expect(t, s1.BookGenre(), s2.BookGenre())
	Expect(t, s1.TLD(), s2.TLD())
}

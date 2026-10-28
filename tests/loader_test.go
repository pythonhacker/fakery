package tests

import (
	"errors"
	"fakery"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// writeData creates <root>/<locale>/<fileName> with the given contents.
func writeData(t *testing.T, root, locale, fileName, contents string) {
	t.Helper()
	full := filepath.Join(root, locale, filepath.FromSlash(fileName))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

// A zero-value DataLoader (as used historically by generator files) must
// still work after Init, and must not panic on SetIsMap before Init.
func TestDataLoaderZeroValue(t *testing.T) {
	var loader fakery.DataLoader
	loader.SetIsMap("generic") // must not panic on nil maps
	Expect(t, true, loader.IsMap("generic"))

	loader.Init("names.json")
	Expect(t, "names.json", loader.FileName())
	// Init resets configuration.
	Expect(t, false, loader.IsMap("generic"))

	data := loader.EnsureLoaded(fakery.DefaultLocale)
	Expect(t, nil, data.Err())
	Expect(t, true, data.Loaded())
	Expect(t, true, len(data.Get("first_name_male")) > 0)
}

func TestDataLoaderGetReturnsSamePointer(t *testing.T) {
	loader := fakery.NewDataLoader("names.json")
	a := loader.Get("en_US")
	b := loader.Get("en_US")
	Expect(t, true, a == b)
	Expect(t, false, a.Loaded())

	// EnsureLoaded hands back the very same instance, now loaded.
	c := loader.EnsureLoaded("en_US")
	Expect(t, true, a == c)
	Expect(t, true, a.Loaded())
}

func TestDataLoaderLocaleIsolation(t *testing.T) {
	loader := fakery.NewDataLoader("names.json")
	us := loader.EnsureLoaded("en_US")
	gb := loader.EnsureLoaded("en_GB")

	Expect(t, true, us != gb)
	Expect(t, "en_US", us.Locale())
	Expect(t, "en_GB", gb.Locale())
	Expect(t, nil, us.Err())
	Expect(t, nil, gb.Err())
	Expect(t, true, len(us.Get("last_name")) > 0)
	Expect(t, true, len(gb.Get("last_name")) > 0)
}

func TestDataLoaderPreloadReturnsCopy(t *testing.T) {
	loader := fakery.NewDataLoader("beer.json")
	first := loader.Preload(fakery.GenericLocale)
	Expect(t, true, len(first) > 0)

	delete(first, "beer_names")
	second := loader.Preload(fakery.GenericLocale)
	_, ok := second["beer_names"]
	Expect(t, true, ok, "mutating a preload result must not affect the cache")
}

func TestDataLoaderPreloadMissing(t *testing.T) {
	loader := fakery.NewDataLoader("does-not-exist.json")
	data := loader.Preload(fakery.GenericLocale)
	Expect(t, true, data != nil)
	Expect(t, 0, len(data))
}

func TestDataLoaderMissingFile(t *testing.T) {
	loader := fakery.NewDataLoader("does-not-exist.json")
	data := loader.EnsureLoaded(fakery.GenericLocale)

	Expect(t, true, data != nil)
	Expect(t, true, data.Loaded())
	Expect(t, true, errors.Is(data.Err(), os.ErrNotExist))
	Expect(t, true, data.Get("anything") == nil)
	Expect(t, 0, data.Len())
	Expect(t, true, data.RandomWeightedItem(fakery.New()) == nil)

	// The error is sticky and returned on every call.
	Expect(t, true, errors.Is(data.Load(), os.ErrNotExist))
}

func TestLocaleDataZeroValueNotInitialized(t *testing.T) {
	var data fakery.LocaleData
	Expect(t, nil, data.Err(), "no error before a load attempt")
	Expect(t, true, errors.Is(data.Load(), fakery.ErrLoaderNotInitialized))
	Expect(t, true, errors.Is(data.Err(), fakery.ErrLoaderNotInitialized))
}

func TestDataLoaderUninitialized(t *testing.T) {
	var loader fakery.DataLoader
	data := loader.EnsureLoaded("en_US")
	Expect(t, true, errors.Is(data.Err(), fakery.ErrLoaderNotInitialized))
}

func TestLocaleDataAccessorsBeforeLoad(t *testing.T) {
	loader := fakery.NewDataLoader("names.json")
	data := loader.Get("en_US")
	// Accessors are safe (and empty) before the data is loaded.
	Expect(t, false, data.Has("last_name"))
	Expect(t, 0, len(data.Keys()))
	Expect(t, 0, data.Len())
	Expect(t, false, data.IsMap())
	Expect(t, nil, data.Err())
}

func TestLocaleDataKeysAndHas(t *testing.T) {
	data := fakery.NewDataLoader("words.json").EnsureLoaded(fakery.GenericLocale)
	Expect(t, true, data.Has("adverbs"))
	Expect(t, false, data.Has("no_such_key"))
	Expect(t, data.Len(), len(data.Keys()))
	Expect(t, true, data.Get("no_such_key") == nil)
}

func TestDataLoaderMapBased(t *testing.T) {
	loader := fakery.NewDataLoader("currency.json")
	loader.SetIsMap(fakery.GenericLocale)
	data := loader.EnsureLoaded(fakery.GenericLocale)

	Expect(t, nil, data.Err())
	Expect(t, true, data.IsMap())
	Expect(t, true, data.Len() > 0)
	item := data.RandomWeightedItem(fakery.New())
	Expect(t, true, len(item["code"]) > 0)
	Expect(t, true, len(item["currency"]) > 0)
}

// Without SetIsMap an array-of-objects file cannot be decoded as an object
// of arrays: the error surfaces through Err instead of a panic.
func TestDataLoaderMapDataWithoutMapConfig(t *testing.T) {
	data := fakery.NewDataLoader("currency.json").EnsureLoaded(fakery.GenericLocale)
	Expect(t, true, data.Err() != nil)
	Expect(t, false, data.IsMap())
	Expect(t, 0, data.Len())
}

// SetIsMap only matters before the first load of that locale.
func TestDataLoaderSetIsMapAfterLoadHasNoEffect(t *testing.T) {
	loader := fakery.NewDataLoader("words.json")
	data := loader.EnsureLoaded(fakery.GenericLocale)
	loader.SetIsMap(fakery.GenericLocale)
	again := loader.EnsureLoaded(fakery.GenericLocale)

	Expect(t, true, data == again)
	Expect(t, nil, again.Err())
	Expect(t, false, again.IsMap())
	Expect(t, true, len(again.Get("adverbs")) > 0)
}

func TestLocaleDataGetWeightedArray(t *testing.T) {
	root := t.TempDir()
	writeData(t, root, "xx", "w.json", `{
		"good": ["a:0.5", "b:0.25", "c:0.25"],
		"bad_weight": ["a:0.5", "b:0.1"],
		"bad_number": ["a:zero"],
		"bad_format": ["a"]
	}`)
	r := fakery.NewRegistryWithRoot(root)
	r.Register("w.json")
	data := r.Load("w.json", "xx")
	Expect(t, nil, data.Err())

	arr, err := data.GetWeightedArray("good", ":")
	Expect(t, nil, err)
	Expect(t, 3, len(arr.Items))
	Expect(t, "a", arr.Items[0].Item)
	Expect(t, 0.5, arr.Items[0].Weight)

	for _, key := range []string{"bad_weight", "bad_number", "bad_format"} {
		arr, err = data.GetWeightedArray(key, ":")
		Expect(t, true, err != nil, key)
		Expect(t, true, arr == nil, key)
	}
}

// Every goroutine must get the same loaded instance, and the result must
// be correct.
func TestDataLoaderEnsureLoadedConcurrentSameInstance(t *testing.T) {
	loader := fakery.NewDataLoader("names.json")
	results := make([]*fakery.LocaleData, raceWorkers)

	runTogether(raceWorkers, func(i int) {
		results[i] = loader.EnsureLoaded("en_US")
	})

	for _, r := range results {
		Expect(t, true, r == results[0])
	}
	Expect(t, nil, results[0].Err())
	Expect(t, true, len(results[0].Get("first_name_female")) > 0)
}

// Readers that use Get (not loaded) concurrently with a loader must not
// race with it.
func TestLocaleDataReadDuringLoad(t *testing.T) {
	loader := fakery.NewDataLoader("names.json")
	data := loader.Get("en_US")

	runTogether(raceWorkers, func(i int) {
		if i%2 == 0 {
			_ = data.Load()
		} else {
			_ = data.Has("last_name")
			_ = data.Keys()
			_ = data.Len()
			_ = data.Err()
		}
	})
	Expect(t, true, data.Has("last_name"))
}

func TestDataLoaderConcurrentInitAndGet(t *testing.T) {
	loader := fakery.NewDataLoader("names.json")
	var wg sync.WaitGroup
	for i := 0; i < raceWorkers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%8 == 0 {
				loader.Init("names.json")
			}
			_ = loader.EnsureLoaded("en_US").Get("last_name")
			_ = loader.FileName()
		}(i)
	}
	wg.Wait()
}

// Check if the loader reversts to the base locale for a requested
// locale that doesn't exist.
func TestResolveLocaleDoesNotFallbackToBaseLanguage(t *testing.T) {
	dataRoot := t.TempDir()

	// A base-language resource exists, but the requested regional
	// locale resource does not.
	baseLocaleDir := filepath.Join(dataRoot, "fr")
	if err := os.MkdirAll(baseLocaleDir, 0o755); err != nil {
		t.Fatal(err)
	}

	resource := filepath.Join(baseLocaleDir, "names.json")
	if err := os.WriteFile(
		resource,
		[]byte(`{"firstName":["Jean"]}`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	// Anthropic's DataLoader supports a custom data root.
	loader := fakery.NewDataLoader("names.json")
	loader.SetRoot(dataRoot)

	// LoadLocale should request exactly "fr_CA".
	f := fakery.NewFromLocale("fr_CA")
	got := f.LoadLocale(loader)

	if got == nil {
		t.Fatal("LoadLocale returned nil")
	}

	if got.Locale() != "fr_CA" {
		t.Errorf(
			"LoadLocale(%q) returned locale %q; want exact locale %q",
			"fr_CA",
			got.Locale(),
			"fr_CA",
		)
	}

	if err := got.Err(); !errors.Is(err, os.ErrNotExist) {
		t.Errorf(
			"LoadLocale(%q) error = %v; want os.ErrNotExist",
			"fr_CA",
			err,
		)
	}
}

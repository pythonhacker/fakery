# Fakery locale data architecture

This document describes how fakery finds, loads, caches and serves its data
files (`data/locales/<locale>/<resource>.json`).

## Goals

* One place that knows about every data resource and how to decode it.
* Lazy loading: a file is read only when a generator first needs it.
* Each `(resource, locale)` pair is read and decoded **at most once**,
  even when many goroutines ask for it at the same moment.
* Locales are isolated from each other; missing locale data falls back
  in a predictable way.
* Keep the public generator API (`f.Name()`, `f.Address()`, ...) and the
  existing loader API (`DataLoader`, `LocaleData`, `LoadLocale`, ...)
  working the same way.

## Components

```
                +-------------------+
 generators --> |  Fakery           |  rng (mutex-guarded source), locale,
 (address.go,   |  localeData(name) |  optional *Registry
  person.go...) |  genericData(name)|
                +---------+---------+
                          |
                          v
                +-------------------+       +----------------------+
                |  Registry         |------>|  fs.FS data source   |
                |  specs / patterns |       |  (default: os.DirFS  |
                |  cache[(name,loc)]|       |   of data/locales)   |
                +---------+---------+       +----------------------+
                          |
                          v
                +-------------------+
                |  LocaleData       |  immutable decoded data and
                |  (+ typed views)  |  lazily built typed struct views
                +-------------------+
```

### ResourceSpec and the resource catalogue (`resources.go`)

A `ResourceSpec{Name, Kind}` describes one data source:

* `Name` is a slash-separated path relative to a locale directory, for
  example `address.json`. A name with glob characters is a **pattern**:
  `cars/*.json` covers every car maker's model file.
* `Kind` selects the decoder:
  * `ListResource` (default): a JSON object of string arrays,
    `{"states": ["Alabama", ...]}`.
  * `MapResource`: a JSON array of string maps,
    `[{"code": "INR", "currency": "Indian Rupee"}, ...]`.

All bundled resources are listed once in `BuiltinResources()` and exposed
as `Resource*` constants (`ResourceAddress`, `ResourceCurrency`,
`ResourceCarModels`, ...). That list is the single source of truth
for what data exists and how it is decoded. Generators refer only to those
constants.

### Registry (`registry.go`)

`Registry` owns the specs and the cache.

* `NewRegistry(fsys fs.FS)` creates a registry over any `fs.FS` whose root
  contains one directory per locale, and registers the builtin resources.
* `DefaultRegistry()` is the process-wide registry over the bundled
  `data/locales` directory. It is found relative to the package source
  file, as before.
* `Register(spec)` adds a spec. Registering the same spec twice does
  nothing. Registering a name again with a different kind returns
  `ErrResourceConflict`. Changing the kind of a resource that already has
  cached data returns `ErrResourceLoaded`. Invalid names return
  `ErrInvalidResource`.
* `Spec(name)` resolves the effective spec: exact name first, then the
  first matching pattern. Otherwise it returns an implicit `ListResource`
  spec, so any file can be loaded without registering it first, as before.
* `Lookup(name, locale) (*LocaleData, error)` returns the data and the
  load error. `Load(name, locale) *LocaleData` returns only the data. It
  never returns nil. On failure it returns an empty `LocaleData` that
  carries the error in `Err()`, and the error is logged once.
* `Loaded`, `Resources` and `Loader(name)` support inspection and
  compatibility.

#### Concurrency model

The cache is a `map[(name, locale)]*cacheEntry`, guarded by a
`sync.RWMutex`. Each entry has its own `sync.Once`:

1. Find or create the entry. There is a fast path under the read lock, and
   the entry is created under the write lock with a double check.
2. Run `entry.once.Do(load)` **outside** the registry lock.

This means:

* Concurrent requests for the same `(resource, locale)` wait for one
  load and then all get the same `*LocaleData` pointer. There is no
  duplicate I/O or decoding.
* Requests for different keys never block each other while I/O runs.
  Only the short map access is serialised.
* Results, including failures, are written once inside the `Once` and
  are read-only after that. This gives the happens-before relationship
  between the writer and every reader.
* `LocaleData` is immutable after it is published. Typed struct views are
  built lazily by a second per-type `sync.Once`, stored in a `sync.Map` on
  the `LocaleData`.

`Fakery` wraps its `math/rand` source in a mutex-guarded `rand.Source64`.
This lets one instance be shared between goroutines without data races.
The sequence of values is the same as with the plain source, so seeded
instances stay deterministic.

#### Locale fallback

Generic data (`GenericLocale`, i.e. `generic/`) never falls back. For any
other locale, if the file **does not exist**, the lookup continues one step
along the chain:

```
requested locale  ->  DefaultLocale (en_US)  ->  GenericLocale
```

Every step is itself a cached `Lookup`, so fallback targets are shared,
not loaded again. The returned `LocaleData.Locale()` reports the locale the
data actually came from. Only *missing* files trigger a fallback.
Malformed data is reported as an error, so it is not silently hidden.

Paths are checked with `fs.ValidPath`. Locales cannot contain `/`, so
requests cannot leave the data root.

### LocaleData (`localedata.go`)

This is the decoded, immutable content of one file:

* `Get(key)`, `GetWeightedArray(key, sep)` and `RandomWeightedItem(f)`
  read the data, as before.
* `Decode(v)` converts the data into any typed value through JSON.
* `Locale()`, `FileName()`, `IsMap()`, `Len()`, `Keys()` and `Err()`
  inspect the data.
* `Load()` is kept for compatibility. It returns the load error.

Generators that work with a typed struct (beer, book, car, color, wine)
use `genericView[T](f, name)`. It decodes the struct once per
`LocaleData` and type, then shares it. This replaces the old eager
`init()` + `ConvertMapToStruct` + package global pattern.

### DataLoader (`loader.go`)

`DataLoader` is now a thin, stateless **handle** (`fileName` plus an
optional `*Registry`) into a registry. `Init`, `SetIsMap`, `Preload`, `Get`,
`EnsureLoaded`, `Fakery.LoadLocale` and `Fakery.LoadGenericLocale` keep
their signatures and delegate to the registry. A zero-value loader used
with a `Fakery` follows that Fakery's registry. Used on its own, it uses
the default registry.

## Using a custom data source

```go
r := fakery.NewRegistry(os.DirFS("/path/to/my/locales")) // or embed.FS, fstest.MapFS
_ = r.Register(fakery.ResourceSpec{Name: "pets.json", Kind: fakery.ListResource})

f := fakery.NewFromLocale("de_DE")
f.SetRegistry(r) // nil restores the default registry
```

## Compatibility considerations

* **Public generator API is unchanged.** All exported generator methods
  keep their signatures. Exported types (`BeerData`, `CarData`,
  `NamesData`, `WeightedArray`, ...) are unchanged.
* **Loader API kept.** `DataLoader`, `LocaleData` and their methods still
  exist with the same signatures. Behaviour changes:
  * `DataLoader.Get(locale)` now returns *loaded* data. It is the same as
    `EnsureLoaded`. Before, it could return an unloaded placeholder.
  * `SetIsMap(locale)` sets the kind for the **resource**, so it applies to
    all locales. Before, it was stored per locale, and only the generic
    locale was ever used this way. Changing the kind after data was
    loaded is rejected and logged.
  * `LocaleData.Load()` only reports the load error. Data from a registry
    is always loaded already.
* **Package-level state removed.** The unexported globals (`addressLoader`,
  `beerData`, ...) and all `init()` functions are gone. Importing the
  package no longer reads beer, book, car, color or wine data at start-up.
  That data is read on first use.
* **Fallback is new for missing locale files.** Before, a locale without a
  file (for example `en_GB/address.json`) logged an error and returned
  empty data. Most generators then panicked with `rand.Intn(0)`. Now the
  data comes from `DefaultLocale`, then from `GenericLocale`.
* **Empty data no longer panics.** `RandomString`, `RandomStringExcl`,
  `RandomWeightedItem` and `LocaleData.RandomWeightedItem` return a zero
  value or an error when there is no data. `RandomStringExcl` no longer
  changes its input slice, which could corrupt shared cached data.
* **The `Fakery.data` field and the `LocaleData` struct tags were
  removed.** Both were unexported and unused. The struct tags on
  unexported fields were what made `go vet` fail.
* **`github.com/google/uuid` dependency dropped.** It was only used for an
  internal `LocaleData.id` field.
* **Thread safety.** Registries, `LocaleData` and the default data are safe
  for concurrent use. A single `Fakery` can be shared across goroutines for
  generation. Changing its configuration (`SetLocale`, `SetRegistry`) at
  the same time as generation is still the caller's responsibility.

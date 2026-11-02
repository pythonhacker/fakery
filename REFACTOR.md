# Refactor: centralised, concurrency-safe locale data loading

## Motivation

Before the refactor, each generator file had its own data plumbing:

* Generator files declared a package-level `DataLoader` (`addressLoader`,
  `personLoader`, `netLoader`, `wordsLoader`, `jobLoader`,
  `currencyLoader`, `beerLoader`, `bookLoader`, `carLoader`,
  `colorLoader`, `wineLoader`), each with its own `init()` to set the file
  name.
* Five of them (beer, book, car, color, wine) read their file eagerly at
  import time. They converted the result into a package-global struct
  with `ConvertMapToStruct` and ignored any errors.
* Map-shaped data was configured by hand in each place. `currency.go`
  called `SetIsMap` in its `init()`. `car.go` built a brand-new loader on
  **every** `CarModel()` call, so the maker's JSON file was read and
  decoded again on each call, with no caching.
* `DataLoader.Get` read and wrote a plain Go map with no lock, and the
  `loaded` flag was written outside its `sync.Once`. Concurrent use of
  fakery was a data race. This was confirmed with `go test -race` on the
  original code.
* `LocaleData` had JSON struct tags on unexported fields, which made
  `go vet ./...` fail. It also pulled in `github.com/google/uuid` for an id
  that nothing read.
* Locales without a given file (for example `en_GB` has no `address.json`)
  returned empty data, and the generators then panicked inside
  `rand.Intn(0)`.

## The new design in one sentence

A single **Registry** describes every data resource once (name + decoding
kind), lazily loads each `(resource, locale)` pair exactly once behind a
per-entry `sync.Once`, and serves immutable `LocaleData` values to
generators. The generators only name the resource they need.

See `DESIGN.md` for the full architecture.

## Changes by area

### New building blocks

* **`registry.go`** contains:
  * `ResourceKind` (`ListResource` / `MapResource`) and `ResourceSpec`.
    Specs can be patterns (`cars/*.json`).
  * The `Registry` type: spec registration with conflict detection, spec
    resolution (exact name, then pattern, then implicit list), a
    `(name, locale)` cache, and the fallback chain
    `locale -> DefaultLocale -> GenericLocale`.
  * The default registry over the bundled `data/locales`, reached through
    `fs.FS`.
  * Sentinel errors: `ErrInvalidResource`, `ErrResourceConflict`,
    `ErrResourceLoaded`.
* **`resources.go`** is the catalogue of bundled resources. It holds the
  `Resource*` name constants and `BuiltinResources()`, the one place that
  says which file is map-shaped (`currency.json`, `cars/*.json`). It also
  has the small `Fakery` helpers generators use: `localeData`,
  `genericData` and `genericView[T]`.
* **`localedata.go`** holds the `LocaleData` type, moved out of
  `helpers.go`. It is now immutable once published and records its load
  error. It has inspection accessors and `Decode`, plus lazily built,
  per-type cached struct views.
* **`loader.go`** holds `DataLoader`, which is now a thin handle into a
  registry. It keeps its full original method set for compatibility.

### Generators

* All package-level loader variables, the package-global data structs and
  every `init()` function were removed from `address.go`, `person.go`,
  `internet.go`, `word.go`, `job.go`, `currency.go`, `beer.go`, `book.go`,
  `car.go`, `color.go` and `wine.go`.
* Generators now use one uniform style:
  * locale-specific data: `f.localeData(ResourceX)`
  * generic data: `f.genericData(ResourceX)`
  * struct-shaped generic data: a small accessor such as
    `f.beerData()`, backed by `genericView[BeerData]`.
* Car models are resolved through the registered `cars/*.json` pattern.
  They are now cached per maker instead of being read from disk on every
  call.

### Fakery

* `Fakery` has an optional `registry` field, with `Registry()`,
  `SetRegistry()` and `Locale()` accessors. The unused `data` field was
  removed.
* The random source is wrapped in a mutex-guarded `rand.Source64`. Shared
  instances are race-free, and seeded sequences are unchanged.
* `RandomString`, `RandomStringExcl` and `RandomWeightedItem` handle empty
  input without panicking. `RandomStringExcl` no longer changes its
  argument, which matters now that slices are shared from the cache.

### Housekeeping

* `go vet ./...` now passes. The invalid struct tags were removed.
* `github.com/google/uuid` was removed from `go.mod` and `go.sum`.
* The `NormalizeString` regexp is compiled once instead of on every call.
* All sources and tests are `gofmt`-clean. This includes two existing test
  files whose imports were not sorted.

## Behavioural compatibility

* All exported generator methods and exported types keep their
  signatures and output semantics. Seeded `Fakery` instances produce the
  same sequences as before.
* `DataLoader`, `LocaleData`, `Fakery.LoadLocale` and
  `Fakery.LoadGenericLocale` keep their signatures. The differences are
  small and listed in `DESIGN.md` ("Compatibility considerations"). In
  short: `Get` returns loaded data, `SetIsMap` applies per resource, and
  `Load` reports the cached load error.
* New behaviour applies only where the old code returned empty data or
  panicked: the fallback for missing locale files, and zero values instead
  of panics on empty data.
* Data is no longer read at import time. A program that never asks for
  beer data never reads `beer.json`.

## Tests

The existing tests in `tests/` are unchanged apart from gofmt import
ordering, and they still pass. New tests were added:

* `tests/registry_test.go` uses an in-memory `fstest.MapFS`, wrapped in an
  fs that counts file opens. It covers:
  * the builtin catalogue and pattern registration
  * cache reuse, checked by pointer identity and a single file open, both
    for direct loads and through generators
  * struct views decoded once
  * locale isolation and `SetLocale`
  * fallback to the default locale and to generic data, using both test
    data and the bundled `en_GB` data
  * the generic locale not falling back
  * missing resources: errors, cached failures, and generators returning
    empty values instead of panicking
  * malformed JSON and data of the wrong shape, with no fallback on
    malformed data
  * map-based data and `Decode`
  * weighted data, including invalid sums and missing separators
  * `Register` conflicts, invalid input and kind changes after loading
  * invalid keys and path traversal
  * `DataLoader` compatibility
  * seeded determinism
  * `RandomStringExcl` not changing its input
* `tests/concurrency_test.go` covers:
  * many goroutines loading the same key: one read, one shared pointer
  * concurrent loads across locales and resources, including fallbacks,
    missing and malformed files: each file read at most once, results
    isolated per locale
  * every generator run concurrently, both with one instance per goroutine
    and with a single shared instance
  * concurrent `Register` and `Load`

The checks used were `gofmt -l .` (no output), `go vet ./...`,
`go test ./...` and `go test -race ./...`.

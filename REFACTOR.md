# Refactor Notes: Locale Data Loading

## Motivation

Before this change, each generator file (`address.go`, `beer.go`, `book.go`,
`car.go`, `color.go`, `currency.go`, `internet.go`, `job.go`, `person.go`,
`wine.go`, `word.go`) declared its own package-level `DataLoader` and an
`init()` to configure it. Five of them also decoded their whole data file into
a struct at program start. This caused several problems:

* **Data races.** `DataLoader.Get` read and wrote its locale map with no
  synchronisation. `SetIsMap` wrote a config map that `Get` read at the same
  time. `EnsureLoaded` read the `loaded` flag outside its `sync.Once`. The
  explicit race test (`tests/racecondition_test.go`) failed under
  `go test -race`, and concurrent first access could crash with
  "concurrent map writes".
* **Duplicated setup.** The same "declare loader, `Init` in `init()`,
  sometimes `SetIsMap`, sometimes `Preload` + `ConvertMapToStruct`" code
  appeared in eleven places.
* **No caching for car models.** `makeFromModel` built a new loader on every
  call, so every `CarModel()` re-read and re-parsed a JSON file.
* **`go vet` failed.** `LocaleData` had JSON struct tags on unexported fields.
* **Hard to test.** The data root was fixed, so missing files, malformed JSON
  and map-based decoding could not be tested in isolation.

## What changed

### A dedicated loading subsystem

The loading code was moved out of `helpers.go` into three focused files:

* **`loader.go`**: `LocaleData` (one resource) and `DataLoader` (all locales
  of one file), rewritten to be concurrency-safe:
  * `DataLoader` is protected by a `sync.RWMutex`, with a double-checked
    get-or-create so every caller gets the same `*LocaleData` per locale.
  * `LocaleData` loads inside `sync.Once` and publishes the result through an
    `atomic.Bool`. After loading, its payload never changes, so it can be read
    without locks.
  * The decoding mode (map vs. array) is read from the loader when the load
    runs, which removes the `SetIsMap`/`Get` race.
  * Load errors are stored and returned by `Err()`, and are logged only once.
  * The zero value works without panicking. `NewDataLoader` was added.
* **`registry.go`**: a `Registry` keyed by data file name. It is the only owner
  of loaders and their caches. It supports exact registrations, prefix families
  for runtime-named files (`cars/<maker>.json`), configurable data roots, and
  `DataView[T]` for lazily decoded typed views.
* **`sources.go`**: one declaration table for all bundled sources (map-based
  ones flagged via `WithMapLocales`), the `cars/` family, exported `Source*`
  identifiers, and the package-level `defaultRegistry` built from that table.

### Generators use one access pattern

Every generator now gets its data in one of three ways:

1. `f.localeSource(SourceX)` for locale-specific data,
2. `f.genericSource(SourceX)` for generic data, or
3. `xData.Get()` on a `DataView[T]` for struct-shaped generic data.

All per-file `DataLoader` variables and `init()` functions were removed. The
eager `init()` decoding of beer, book, car, color and wine data became lazy,
concurrency-safe `DataView`s. Car model lookup now goes through the
registered `cars/` family, so each maker file is loaded once and then cached.

### Robustness fixes found on the way

* `RandomWeightedItem` returns `nil` on empty map data instead of panicking.
* `GetWeightedArray` reports malformed `item:weight` entries as errors instead
  of panicking with an index out of range.
* `Preload` returns a copy, so callers cannot change the shared cache.
* `Registry.Register` applies options to a new loader before publishing it.
  The concurrent family-registration test showed that publishing first let
  another goroutine decode a map-based file before its map mode was set.

### Cleanups that follow from the design

* The unused `LocaleData.id` field was removed, and with it the
  `github.com/google/uuid` dependency.
* The ineffective JSON tags on unexported `LocaleData` fields were removed,
  so `go vet ./...` passes.
* The unused `Fakery.data` field and `namesData` variable were removed.
* `getModuleFile` and the `prefix` constant were replaced by
  `resolveDataPath` and a `sync.OnceValue`-cached module data root.
* Pre-existing test files with unsorted imports were gofmt'd.

## Public API impact

* **Kept (same signatures and behaviour):** all generator methods,
  `Fakery.LoadLocale`, `Fakery.LoadGenericLocale`, `DataLoader.{Init, SetIsMap,
  Get, EnsureLoaded, Preload}`, `LocaleData.{Get, Load, RandomWeightedItem,
  GetWeightedArray}`, `WeightedArray`, `WeightedItem`, `ConvertMapToStruct`.
* **Added:** registry, source identifiers, typed views, sentinel errors and
  inspection helpers (`Err`, `Loaded`, `IsMap`, `Has`, `Keys`, `Len`,
  `Decode`, …). See `DESIGN.md` for the full list.
* **Behavioural notes:** struct data is decoded lazily instead of in `init()`,
  car model files are cached, and `DataLoader` must not be copied after use
  because it now holds a mutex.

## Tests

The existing tests pass unchanged, including `tests/racecondition_test.go`
under `-race`. New tests:

* **`tests/loader_test.go`** (`DataLoader` / `LocaleData`): zero-value
  usability, pointer identity and caching, locale isolation, preload copy
  semantics, missing files, uninitialised loaders, accessors before load,
  map-based decoding, wrong-mode decoding, `SetIsMap` after load, weighted
  arrays (valid, bad weight, bad number, bad format), concurrent
  `EnsureLoaded` returning one instance, reads during a concurrent load, and
  concurrent `Init`/`Get`.
* **`tests/registry_test.go`** (`Registry` / `DataView` / generators):
  repeated access reuses the cache (proved by changing and then deleting the
  file after the first load), idempotent registration, locale isolation,
  isolation between registries, unregistered sources, missing resources,
  malformed JSON (syntax error, wrong shape, bad map values) with cached
  failures, map-based data per locale, prefix families with longest-prefix
  match, concurrent loads of one resource and of many locales, concurrent
  registration and family loads, `DataView` concurrency and failure handling,
  built-in source coverage, a model file for every listed car maker,
  generators running concurrently across locales, and the public `Fakery`
  wrappers.

Verified with `gofmt -l .` (clean), `go vet ./...`, `go test ./...` and
`go test -race ./...`.

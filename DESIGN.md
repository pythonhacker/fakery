# Locale Data Loading: Design

This document describes how fakery finds, loads, caches and serves locale data,
and what that means for compatibility.

## Goals

1. **One place to declare data sources.** Every data file is declared once in a
   central registry, along with how to decode it (an object of string arrays,
   or an array of maps).
2. **Lazy, at-most-once loading.** A `(file, locale)` resource is read from disk
   the first time it is needed and never again.
3. **Safe for concurrent use.** Any number of goroutines may ask for any mix of
   resources and locales without data races, duplicate loads or a corrupted
   cache.
4. **Same behaviour as before.** Locale and generic lookup rules, weighted data,
   map-based data and every public generator method work as they did.

## Architecture

The subsystem has three layers. Each layer returns one shared instance per
key, so callers asking for the same thing always get the same cached object.

```
                    ┌──────────────────────────────────────────────┐
 Fakery generators  │ f.localeSource(SourceNames)   (f.locale)     │
 (person.go, …)     │ f.genericSource(SourceWords)  ("generic")    │
                    │ carData.Get()  ← DataView[CarData]           │
                    └───────────────────────┬──────────────────────┘
                                            │
                    ┌───────────────────────▼──────────────────────┐
 Registry           │ fileName → *DataLoader     (sync.RWMutex)    │
 (registry.go)      │ prefix families: "cars/" → options           │
                    │ Load(fileName, locale) *LocaleData           │
                    └───────────────────────┬──────────────────────┘
                                            │
                    ┌───────────────────────▼──────────────────────┐
 DataLoader         │ locale → *LocaleData       (sync.RWMutex)    │
 (loader.go)        │ locale → isMap config                        │
                    └───────────────────────┬──────────────────────┘
                                            │
                    ┌───────────────────────▼──────────────────────┐
 LocaleData         │ sync.Once load → immutable payload           │
 (loader.go)        │ atomic "loaded" flag guards all readers      │
                    └──────────────────────────────────────────────┘
                     reads  <root>/<locale>/<fileName>
                     (root defaults to the module's data/locales)
```

### 1. `LocaleData`: one resource

* Holds the decoded data of one `(fileName, locale)` pair: either
  `map[string][]string` (most files) or `[]map[string]string` (map-based files
  such as `currency.json` and `cars/*.json`).
* **Loads once.** `Load()` runs the file read and decode inside a `sync.Once`.
  Concurrent callers block until that first load finishes and then all see the
  same result, including the same error.
* **Never changes after loading.** The payload is written only inside the
  `Once` and is never modified afterwards, so reading it needs no lock.
* **Safe to read before loading.** Every accessor (`Get`, `Has`, `Keys`, `Len`,
  `IsMap`, `Err`, `RandomWeightedItem`, `Decode`) first checks an
  `atomic.Bool` that is set only after the payload is written. This gives the
  needed happens-before edge, so even a `*LocaleData` obtained from
  `DataLoader.Get` (not yet loaded) can be read safely while another goroutine
  loads it. Until then, accessors return empty values.
* **Failures are values.** A missing file or bad JSON leaves the data empty and
  is reported through `Err()`. The failure is logged once, by the call that
  did the load. Failures are cached just like successes.
* **Decoding mode is read at load time.** The map/array mode is taken from the
  owning `DataLoader` when the load runs, not when the `LocaleData` is created.
  So a `SetIsMap` call made before the first load always applies.

### 2. `DataLoader`: every locale of one file

* `Get(locale)` uses a read-locked fast path, then a write-locked
  double-check, so concurrent first requests for a locale all receive **the
  same** `*LocaleData`. Together with the `Once` above, this means each locale
  is read from disk at most once.
* `SetIsMap`, `IsMap`, `Init` and `FileName` are all mutex-protected.
* The zero value is usable: the internal maps are allocated on first need,
  so `var l DataLoader; l.SetIsMap(...)` no longer panics. `NewDataLoader(name)`
  is a convenience constructor.
* `EnsureLoaded(locale)` = `Get` + one-time `Load`, logging the error once.
* `Preload(locale)` returns a *copy* of the top-level map, so callers cannot
  change the shared cache.

### 3. `Registry`: every file, declared once

* `Register(fileName, opts...)` creates the loader for a file, or returns the
  existing one. It is idempotent, so registering a name twice gives the same
  loader and the same cache. A new loader is configured *before* it is
  published in the registry, so concurrent readers never see it half set up.
* `WithMapLocales(locales...)` is the option that selects map-based decoding
  for specific locales.
* `RegisterPrefix(prefix, opts...)` declares a **family** of files whose names
  are only known at runtime, e.g. `cars/` for per-maker model files. When a
  name is first requested, it is registered on the spot with the family's
  options. If several prefixes match, the longest one wins.
* `Load(fileName, locale)` is the single entry point used by generators. It
  never returns nil: an unknown name gives an empty `LocaleData` whose `Err()`
  wraps `ErrSourceNotRegistered`.
* `NewRegistryWithRoot(dir)` serves files from another data root, for tests or
  for applications that ship their own data. Registries never share caches.

### Source declarations (`sources.go`)

`builtinSources` and `builtinFamilies` form one declarative table of all
bundled data, with exported identifiers (`SourceNames`, `SourceCurrency`, …,
`SourceCarModelsPrefix`). `RegisterBuiltinSources(r)` applies the table to any
registry. The package-level `defaultRegistry` is built by a variable
initialiser, so it exists before any `init()` or dependent variable runs. Adding
a data file means adding one line to the table.

### Typed views (`DataView[T]`)

Several generators work with a struct form of a generic data file (`BeerData`,
`BooksData`, `CarData`, `ColorData`, `WineData`). `DataView[T]` decodes the
resource into `T` lazily, exactly once (`sync.Once`), and returns the same
read-only `*T` from then on. This replaces the old eager `init()` +
`Preload` + `ConvertMapToStruct` code: data is now loaded only when a
generator first needs it, and the load is concurrency-safe.

### `Fakery` accessors

Generators call two small internal helpers:

* `f.localeSource(source)`: the data for the instance's locale (`f.locale`).
* `f.genericSource(source)`: the locale-independent data (`GenericLocale`).

The public `LoadLocale(*DataLoader)` and `LoadGenericLocale(*DataLoader)`
wrappers are kept unchanged for callers that manage their own loaders.

## Concurrency guarantees (summary)

| Scenario | Mechanism | Result |
|---|---|---|
| N goroutines, same file + locale, first access | Registry RW lock → DataLoader double-checked RW lock → `sync.Once` | One `LocaleData`, one disk read; every caller waits for it and sees the same data/error |
| N goroutines, same file, different locales | DataLoader write lock only around map insert; separate `Once` per locale | Locales load independently and each is cached separately |
| Runtime family registration (`cars/x.json`) | `Register` re-checks under the registry write lock and fully configures a new loader (options applied) *before* publishing it | One loader per name, even if many goroutines register it at once; no goroutine can load it before its map mode is set |
| `SetIsMap` concurrently with `Get` | Both behind the DataLoader mutex; mode read at load time | No race; mode applies if set before the first load |
| Readers vs. loader on the same `LocaleData` | `atomic.Bool` published after the payload write | No race; readers see either empty or fully loaded data |
| Typed struct access | `DataView` `sync.Once` | Decoded once; shared read-only pointer |

`Fakery` itself is **not** goroutine-safe because it owns a `*rand.Rand`. Use
one `Fakery` per goroutine. They all share the concurrency-safe default
registry.

## Behaviour preserved

* **Locale rules.** Locale-specific sources (names, address states) are looked
  up in `f.locale`. Generic sources always come from `generic`. As before,
  there is **no implicit cross-locale fallback**: if a locale lacks a file
  (e.g. `en_GB/address.json`), the result is empty data plus a logged error,
  exactly as in the previous implementation. A future fallback chain (e.g.
  `en_GB → en_US → generic`) would belong in `Registry.Load`, so generators
  would not need to change.
* **Weighted data.** `GetWeightedArray(key, sep)` and the `WeightedArray`
  validation are unchanged. Malformed entries (no separator) now return an
  error instead of panicking with an index out of range.
* **Map-based data.** `currency.json` (generic) and `cars/*.json` are
  configured as map based in the declaration table. `RandomWeightedItem` keeps
  its behaviour and returns `nil` on empty data instead of panicking (indexing
  a nil map yields `""`).
* **Public API.** Every exported generator method has the same signature and
  output. `DataLoader` (`Init`, `SetIsMap`, `Get`, `EnsureLoaded`, `Preload`),
  `LocaleData` (`Get`, `Load`, `RandomWeightedItem`, `GetWeightedArray`),
  `ConvertMapToStruct`, `WeightedArray` and `WeightedItem` are all kept.

## Compatibility notes

* **`DataLoader` must not be copied after first use** because it now holds a
  mutex (`go vet` copylocks will flag copies). The old code only ever used it
  through pointers, so no existing usage is affected.
* **The struct-based data is no longer loaded eagerly.** It used to be decoded
  in `init()`; now it is decoded on first use. Programs that never generate
  beer/book/car/color/wine data no longer pay for loading it at startup.
* **Car model files are now cached.** `CarModel()` used to build a fresh loader
  and re-read `cars/<maker>.json` on *every* call. It now reads each maker file
  once.
* **`LocaleData` lost its unused `id` field and its JSON struct tags** on
  unexported fields. Those tags never had any effect and made `go vet` fail.
  As a result the `github.com/google/uuid` dependency was removed.
* **The unused `Fakery.data` field and package variable `namesData` were
  removed.** Both were unexported.
* **New exported API (additive only):** `NewDataLoader`, `DataLoader.IsMap`,
  `DataLoader.FileName`, `LocaleData.{Locale, FileName, Loaded, Err, IsMap,
  Has, Keys, Len, Decode}`, `Registry`, `NewRegistry`, `NewRegistryWithRoot`,
  `SourceOption`, `WithMapLocales`, `DataView`, `NewDataView`,
  `DefaultRegistry`, `RegisterBuiltinSources`, `CarModelsSource`, the
  `Source*` constants, `ErrLoaderNotInitialized` and `ErrSourceNotRegistered`.
* **Requires Go ≥ 1.21** for generics and `sync.OnceValue`. The module already
  declares `go 1.23.4`.

## File map

| File | Responsibility |
|---|---|
| `loader.go` | `LocaleData`, `DataLoader`, path resolution |
| `registry.go` | `Registry`, source options and families, `DataView[T]` |
| `sources.go` | Built-in source declarations, `defaultRegistry`, source name helpers |
| `fakery.go` | `Fakery` accessors (`LoadLocale`, `LoadGenericLocale`, `localeSource`, `genericSource`) |
| `helpers.go` | General helpers only (weighted arrays, string utilities); no loading code |
| generator files | Use only `localeSource` / `genericSource` / `DataView`; no loader state |

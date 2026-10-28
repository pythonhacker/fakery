# Locale-Data Loading Architecture (`DESIGN.md`)

## 1. Overview & Design Goals

`fakery` generates realistic synthetic data backed by JSON locale datasets stored under `data/locales/<locale>/<resource>.json` (both locale-specific datasets such as `en_US/names.json` and `en_GB/names.json`, and locale-independent datasets under `generic/`).

The redesigned locale-data loading subsystem establishes a **single, centralized, lazy, and concurrency-safe architecture** governed by four core principles:

1. **Centralized Resource Registration (`LoaderRegistry`)**: Every static resource and dynamic resource rule (including map-decoded datasets) is registered once in a central registry keyed by resource filename rather than scattered across per-generator global `DataLoader` variables and `init()` functions.
2. **True Lazy Single-Flight Loading**: Importing `fakery` performs zero filesystem I/O or JSON parsing. Each `(resource, locale)` pair is decoded from disk at most once on first demand via `sync.Once`, and reused from memory on all subsequent requests.
3. **End-to-End Concurrency Safety**: Every layer—from registry lookup, per-locale loader cache indexing, single-flight JSON decoding, and weighted-array caching to defensive copies on read—is synchronized against concurrent reads and writes (`go test -race`).
4. **Full Backward Compatibility**: Existing public generator signatures, exported data structures, locale fallback semantics, generic data access, weighted-item parsing, and legacy `DataLoader` / `LocaleData` entry points remain intact.

---

## 2. Layered Architecture

The subsystem is organized into three clearly separated layers:

	+-------------------------------------------------------------------+
	|                     Generator Methods (*Fakery)                   |
	|   Person, Address, Beer, Book, Car, Color, Currency, Internet...  |
	|        f.LoadResource(name)   /   f.LoadGenericResource(name)     |
	+---------------------------------+---------------------------------+
                                  |
								  v
	+-------------------------------------------------------------------+
	|                   LoaderRegistry (registry.go)                    |
	|  - Keyed by resource filename ("names.json", "cars/audi.json")    |
	|  - Pre-registers built-in ResourceConfigs + dynamic ResourceRules |
	|  - RWMutex double-checked lookup -> one *DataLoader per resource  |
	+---------------------------------+---------------------------------+
                                  |
                                  v
	+-------------------------------------------------------------------+
	|                      DataLoader (loader.go)                       |
	|  - Scoped to a single resource filename & base locales directory  |
	|  - Maps locale -> *LocaleData under sync.RWMutex                  |
	|  - Implements locale fallback resolution (ResolveLocale)          |
	+---------------------------------+---------------------------------+
                                  |
                                  v
	+-------------------------------------------------------------------+
	|                      LocaleData (loader.go)                       |
	|  - Scoped to a single (resource, locale) pair                     |
	|  - Single-flight disk read & JSON unmarshal via sync.Once         |
	|  - Stores key->[]string, []map[string]string, & weighted caches   |
	|  - Returns defensive copies to preserve cache immutability        |
	+-------------------------------------------------------------------+

### 2.1 `LoaderRegistry` (`registry.go`)

`LoaderRegistry` is the single source of truth for resource metadata and `DataLoader` lifecycle management:

- **Static Resource Registration**: `newDefaultRegistry` registers all built-in resource descriptors (`ResourceAddress`, `ResourceBeer`, `ResourceBook`, `ResourceCar`, `ResourceColor`, `ResourceCurrency`, `ResourceInternet`, `ResourceJobs`, `ResourceNames`, `ResourceWine`, `ResourceWords`) once at package initialization without touching the filesystem.
- **Map-Decoded Resources**:
  - `currency.json` is registered centrally with `IsMap: true` (`[]map[string]string` decoding).
  - Dynamic car model catalogs (`cars/<maker>.json`, e.g., `cars/toyota.json`) are governed by a registered `ResourceRule` matching `cars/*.json` with `IsMap: true`. When a car make is requested for the first time, `LoaderRegistry.Loader` instantiates and caches its `DataLoader` once so subsequent requests reuse both the loader and its decoded `LocaleData`.
- **Custom & Test Registries**: While `DefaultRegistry()` serves package-wide default requests, `NewLoaderRegistry(baseDir)` and `NewDefaultLoaderRegistry(baseDir)` (paired with `NewWithRegistry` / `Fakery.SetRegistry`) allow isolated registries backed by custom directories for testing and multi-tenant embeddings.

### 2.2 `DataLoader` (`loader.go`)

Each `DataLoader` manages the per-locale cache (`map[string]*LocaleData`) and decoding mode (`defaultIsMap` and per-locale `configIsMap`) for one resource file:

- `Get(locale)` performs double-checked locking (`RLock` fast path, `Lock` slow path with re-check) so concurrent first requests for the same locale always receive the exact same `*LocaleData` instance and distinct locales remain isolated in `localeDataMap`.
- `SetIsMap(locale)` and `SetDefaultIsMap(isMap)` synchronize configuration updates under `loader.mu` and atomically propagate `isMap` to any unloaded `LocaleData` entry.
- `EnsureLoaded(locale)` loads the exact requested locale without fallback and logs any load failure at most once per `(resource, locale)`.
- `ResolveLocale(locale)` implements locale fallback resolution for locale-aware generators.

### 2.3 `LocaleData` (`loader.go`)

`LocaleData` encapsulates the decoded payload and load state for one `(resource, locale)` pair:

- **Single-Flight Lazy Loading**: `Load()` executes `l.load()` inside `l.once.Do(...)`. All concurrent callers requesting the same `(resource, locale)` block on `l.once.Do` until the single filesystem read and JSON decode finish, eliminating duplicate I/O and preventing partially initialized reads.
- **Atomic Decode & Error Caching**: JSON payloads are unmarshaled into local variables (`map[string][]string` or `[]map[string]string`) before acquiring `l.mu.Lock()` and publishing them to `l.data` / `l.dataMap`. If reading or decoding fails, cached fields remain `nil` and `l.loadErr` is recorded so all concurrent and future callers observe the exact same error without re-reading disk.
- **Cache Immutability**: `Get(key)`, `Entries()`, `RandomWeightedItem(f)`, and `GetWeightedArray(key, sep)` return defensive copies so caller mutations can never corrupt shared cached slices, maps, or weights.
- **Weighted-Array Caching**: `GetWeightedArray(key, sep)` validates and caches parsed `*WeightedArray` instances per `(key, sep)` tuple under `l.mu`, avoiding repeated string splitting and float parsing on hot paths such as `TLD()`.
- **Path Safety**: Relative paths are normalized and checked before joining with `baseDir` so traversal paths outside the locale root are rejected with `os.ErrNotExist`.

---

## 3. Concurrency Model & Synchronization Invariants

| Shared State | Synchronization Primitive | Guarantee |
| --- | --- | --- |
| `LoaderRegistry.loaders`, `configs`, `rules` | `sync.RWMutex` (`r.mu`) | Lock-free-style concurrent read scaling via `RLock`; single `DataLoader` instantiation per resource under `Lock`. |
| `DataLoader.localeDataMap`, `configIsMap`, `defaultIsMap` | `sync.RWMutex` (`loader.mu`) | Safe concurrent access for identical and distinct locales; safe concurrent `SetIsMap` + `Get`. |
| `LocaleData` file read & JSON decode | `sync.Once` (`l.once`) | At-most-once execution per `(resource, locale)`; happens-before memory barrier for all callers. |
| `LocaleData.data`, `dataMap`, `weighted`, `loadErr` | `sync.RWMutex` (`l.mu`) | Race-free reads and writes even if callers invoke `Get`/`Err` concurrently with `Load`. |
| `LocaleData.loaded`, `isMap`, `loadCount` | `atomic.Bool` / `atomic.Int64` | Lock-free race-free status inspection (`IsLoaded`, `IsMap`, `LoadCount`). |
| `Fakery.locale`, `registry`, `rng` | `sync.RWMutex` (`f.mu`) & `sync.Mutex` (`f.rngMu`) | Safe locale/registry switching and safe RNG draws even when a `*Fakery` instance is shared across goroutines. |

Locks follow a strict acyclic hierarchy (`Fakery` -> `LoaderRegistry` -> `DataLoader` -> `LocaleData`) and are never held across nested subsystem calls, preventing deadlocks.

---

## 4. Locale Resolution & Fallback Semantics

Generators distinguish between two access patterns:

1. **Generic Locale Data (`LoadGenericResource` / `LoadGenericLocale`)**:
   Resources shared across all locales (`beer.json`, `book.json`, `car.json`, `cars/*.json`, `color.json`, `currency.json`, `internet.json`, `jobs.json`, `wine.json`, `words.json`, and generic keys of `address.json`) resolve directly against `GenericLocale` (`"generic"`).

2. **Locale-Specific Data with Fallback (`LoadResource` / `LoadLocale` / `ResolveLocale`)**:
   Locale-dependent resources (`names.json` and regional keys of `address.json`) evaluate candidates in deterministic order:
   1. Exact requested locale (e.g., `"en_GB"`, `"fr_CA"`).
   2. Base language prefix when a regional subtag is present (e.g., `"fr"` for `"fr_CA"`).
   3. `DefaultLocale` (`"en_US"`).
   4. `GenericLocale` (`"generic"`).

   **Error Semantics During Fallback**:
   - Fallback advances to the next candidate **only** when the candidate file is absent (`errors.Is(err, os.ErrNotExist)`). For example, `NewFromLocale("en_GB").Address()` loads `en_GB` names from `en_GB/names.json`, falls back to `en_US/address.json` for state names because `en_GB/address.json` does not exist, and formats British postal codes and `"United Kingdom"` country metadata.
   - If a locale file **exists** on disk but contains **malformed JSON** or an incompatible JSON type, `ResolveLocale` stops immediately and surfaces that `LocaleData` entry and its decode error rather than silently masking data corruption behind `DefaultLocale`.

---

## 5. Compatibility Considerations

All public APIs and observable generator behaviors are preserved:

- **Generator Signatures & Determinism**: Every public method and struct across `Fakery` (`Person`, `Address`, `Beer`, `Book`, `Car`, `Color`, `CreditCard`, `Currency`, `Emoji`, `Internet`, `Job`, `OS`, `UserAgent`, `Wine`, `Word`) retains its exact signature and output format. Deterministic seeding via `NewFromSeed(seed)` produces repeatable sequences.
- **Exported Loader & Data Types**:
  - `DataLoader` (`Init`, `SetIsMap`, `Preload`, `Get`, `EnsureLoaded`) and `LocaleData` (`Get`, `RandomWeightedItem`, `Load`, `GetWeightedArray`) remain exported and backward-compatible with external callers and tests (such as `tests/racecondition_test.go`), while gaining full concurrency safety and inspection accessors (`ID`, `Locale`, `FileName`, `IsMap`, `IsLoaded`, `LoadCount`, `Err`, `Entries`).
  - Exported schema structs (`BeerData`, `BooksData`, `CarData`, `ColorData`, `NamesData`, `WineData`, `WordsData`) and `ConvertMapToStruct` remain available for external callers.
  - `Fakery.LoadLocale(*DataLoader)` and `Fakery.LoadGenericLocale(*DataLoader)` remain supported alongside the new registry-backed `Fakery.LoadResource(string)` and `Fakery.LoadGenericResource(string)` methods.
- **`go vet` Compliance**: Unexported fields on `LocaleData` no longer carry invalid `json` struct tags, ensuring clean `go vet ./...` execution.

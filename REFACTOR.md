# Locale-Data Subsystem Refactoring Summary (`REFACTOR.md`)

## 1. Motivation & Prior Architectural Limitations

Prior to this refactoring, locale and generic resource loading in `fakery` suffered from four structural issues:

1. **Duplicated Package-Level Loader Initialization**:
   Eleven generator files (`address.go`, `beer.go`, `book.go`, `car.go`, `color.go`, `currency.go`, `internet.go`, `job.go`, `person.go`, `wine.go`, and `word.go`) each declared their own package-level `DataLoader` variable and `init()` function.
2. **Inconsistent Loading Paradigms & Uncached Per-Call I/O**:
   - Five generators (`beer.go`, `book.go`, `car.go`, `color.go`, `wine.go`) eagerly read JSON files during package `init()`, copied slices into `map[string]interface{}` via `Preload`, re-marshaled those maps back into JSON bytes via `ConvertMapToStruct`, and unmarshaled them a second time into package-level structs—incurring startup disk I/O and double serialization even when those generators were never called.
   - `car.go` instantiated a brand-new local `DataLoader` inside `makeFromModel` on every call to `CarModel`, `CarMakeAndModel`, or `Car`, re-reading and re-parsing `data/locales/generic/cars/<maker>.json` from disk on every invocation without caching.
3. **Concurrency Hazards & Data Races**:
   - `DataLoader.Get` and `DataLoader.SetIsMap` read and mutated `localeDataMap` and `configIsMap` without synchronization, causing concurrent map read/write panics, data races, and duplicate `LocaleData` allocations under concurrent load.
   - `DataLoader.EnsureLoaded` inspected `LocaleData.loaded` outside `sync.Once` while `LocaleData.Load` wrote `loaded = true` inside `sync.Once`, creating a data race under `go test -race`.
   - `LocaleData.Load` stored its load error in a stack-local variable inside `Load()`, causing concurrent or subsequent callers of a failed load to falsely receive `nil` errors.
   - `LocaleData.Get` exposed raw internal slices and `RandomStringExcl` mutated slice backing arrays in place, risking cache corruption across callers.
4. **Static Analysis & Edge-Case Fragility**:
   - Unexported fields in `LocaleData` carried `json` struct tags, causing `go vet ./...` to fail.
   - Missing resources, empty slices, or malformed weighted-array strings could trigger index-out-of-range or `rand.Intn(0)` panics instead of degrading cleanly.

---

## 2. Architectural & Design Changes

### 2.1 Introduction of Central `LoaderRegistry` (`registry.go`)
- Created `LoaderRegistry` as a centralized, concurrency-safe registry mapping resource identifiers (such as `address.json`, `names.json`, `currency.json`, and `cars/<maker>.json`) to singleton `DataLoader` instances.
- Consolidated all built-in resource registrations into `newDefaultRegistry`, eliminating scattered loader setup across generator files.
- Configured map-slice JSON decoding centrally both for static map resources (`currency.json`) and for dynamic car model catalogs via a `ResourceRule` matching `cars/*.json`.
- Added constructor support for isolated registries (`NewLoaderRegistry`, `NewDefaultLoaderRegistry`) to enable deterministic filesystem-isolated testing without mutating repository locale files.

### 2.2 Separation & Hardening of `DataLoader` and `LocaleData` (`loader.go`)
- Extracted `LocaleData`, `DataLoader`, `WeightedItem`, `WeightedArray`, and module locale path resolution out of `helpers.go` into a dedicated `loader.go` module, leaving `helpers.go` focused on string, slice, and conversion utilities.
- Synchronized `DataLoader` with `sync.RWMutex` and double-checked locking in `Get(locale)` so concurrent requests for the same locale share a single `LocaleData` instance while concurrent requests for distinct locales populate the cache safely.
- Synchronized `SetIsMap` and `SetDefaultIsMap` with `Get` using `loader.mu` and atomic state updates on unloaded `LocaleData` entries.
- Removed the unsynchronized outer `loaded` check from `EnsureLoaded`, routing all loads through `LocaleData.Load()` and its `sync.Once` barrier so concurrent callers wait for completion and observe consistent cached data and cached `loadErr` state.
- Unmarshaled JSON into local temporary structures before publishing under `LocaleData.mu`, ensuring malformed JSON never leaves partial state in memory.
- Added per-`(key, sep)` caching and strict validation for `GetWeightedArray`, defensive cloning on `Get`, `Entries`, `RandomWeightedItem`, and `GetWeightedArray`, path-traversal protection, and atomic observability accessors (`ID`, `Locale`, `FileName`, `IsMap`, `IsLoaded`, `LoadCount`, `Err`).
- Implemented `DataLoader.ResolveLocale` to formalize fallback resolution (`requested locale -> base language -> DefaultLocale -> GenericLocale`) strictly on `os.ErrNotExist` while preserving malformed JSON errors.
- Removed invalid `json` struct tags from unexported `LocaleData` fields to satisfy `go vet`.

### 2.3 `Fakery` Core Integration (`fakery.go`)
- Added `LoadResource` and `LoadGenericResource` methods on `Fakery` backed by the instance's `LoaderRegistry` (defaulting to `DefaultRegistry()`), while retaining `LoadLocale(*DataLoader)` and `LoadGenericLocale(*DataLoader)` for backward compatibility.
- Added `Locale`, `Registry`, `SetRegistry`, and `NewWithRegistry` to support concurrency-safe locale inspection and custom registry injection.
- Synchronized `Fakery`'s RNG and locale state (`rngMu` and `mu`), guarded empty-slice/zero-bound edge cases in `IntRange`, `OneOf`, `RandomString`, and `RandomWeightedItem`, and updated `RandomStringExcl` to filter into a fresh slice rather than mutating caller backing arrays.

### 2.4 Consistent Generator Refactoring
Refactored all eleven data-backed generator modules to follow a single lazy-access pattern through `f.LoadResource(...)` and `f.LoadGenericResource(...)`, removing all package-level `DataLoader` variables, eager `Preload`/`ConvertMapToStruct` globals, per-call `DataLoader` allocations, and `init()` functions:
- `address.go`: Uses `ResourceAddress` via `LoadResource` (regional state data with fallback) and `LoadGenericResource` (shared street, building, city, and country data).
- `beer.go`: Uses `ResourceBeer` lazily via `LoadGenericResource`, removing eager `beerLoader`, `beerData`, and `init()`.
- `book.go`: Uses `ResourceBook` lazily via `LoadGenericResource`, removing eager `bookLoader`, `booksData`, and `init()`.
- `car.go`: Uses `ResourceCar` and centrally registered `cars/<maker>.json` map resources via `LoadGenericResource`, eliminating eager `carLoader`/`carData`, `init()`, and the uncached local `makerLoader` in `makeFromModel`.
- `color.go`: Uses `ResourceColor` lazily via `LoadGenericResource`, removing eager `colorLoader`, `colorData`, and `init()`.
- `currency.go`: Uses centrally map-configured `ResourceCurrency` via `LoadGenericResource`, removing `currencyLoader` and `init()`.
- `internet.go`: Uses `ResourceInternet` via `LoadGenericResource` (including cached weighted TLD arrays), removing `netLoader` and `init()`.
- `job.go`: Uses `ResourceJobs` via `LoadGenericResource`, removing `jobLoader` and `init()`.
- `person.go`: Uses `ResourceNames` via `LoadResource`, removing `personLoader`, unused `namesData`, and `init()`.
- `wine.go`: Uses `ResourceWine` lazily via `LoadGenericResource`, removing eager `wineLoader`, `wineData`, and `init()`.
- `word.go`: Uses `ResourceWords` via `LoadGenericResource`, removing `wordsLoader` and `init()`.

### 2.5 Test Suite Expansion (`tests/loader_test.go` & `tests/racecondition_test.go`)
- Preserved all existing generator tests and concurrency regression tests in `tests/racecondition_test.go` (all of which now pass under `go test -race ./...`).
- Added comprehensive unit and concurrency tests in `tests/loader_test.go` covering:
  - Central registry registration for static and dynamic map-decoded resources.
  - Lazy loading verification (`LoadCount() == 0` prior to access) and cached reuse across repeated accesses even after removing backing files from disk, including dynamic car-maker model loaders.
  - Strict isolation between `en_US`, `en_GB`, and custom locales, plus dynamic `SetLocale` switching.
  - Multi-step locale fallback (`regional -> base language -> DefaultLocale -> GenericLocale`) and non-US address generation.
  - High-concurrency single-flight loading, multi-locale cache isolation, dynamic map-loader access, and end-to-end concurrent generator execution under `-race`.
  - Missing resource caching, path-traversal rejection, malformed JSON syntax and type-mismatch error propagation (verifying malformed JSON is never masked by fallback), weighted-array validation, defensive copy immutability, word generators, and deterministic `NewFromSeed` behaviour.

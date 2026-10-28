// Locale data loading subsystem.
//
// Data files live under <dataRoot>/<locale>/<fileName>, where dataRoot
// defaults to the "data/locales" directory shipped with this module.
//
// The subsystem has three layers:
//
//  1. LocaleData - one (fileName, locale) resource. It is decoded lazily,
//     exactly once, guarded by a sync.Once. After loading it is immutable,
//     so any number of goroutines may read it without further locking.
//  2. DataLoader - every locale of one data file. It hands out a single
//     shared *LocaleData per locale, guarded by a sync.RWMutex.
//  3. Registry (registry.go) - every data file, keyed by file name. It
//     hands out a single shared *DataLoader per file.
//
// Since each layer returns one shared instance per key, every
// (fileName, locale) pair is read from disk at most once per Registry.
package fakery

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

// Relative location of the bundled locale data inside the module.
const dataLocalesDir = "data/locales"

// ErrLoaderNotInitialized is returned when loading a LocaleData that has
// no file name or locale. This happens with a zero-value LocaleData or a
// DataLoader that was never given a file name via Init.
var ErrLoaderNotInitialized = errors.New("error - loader not initialized")

// moduleDataRoot is the absolute path of the bundled data/locales
// directory. It is resolved once from this source file's location, so it
// works no matter what the caller's working directory is.
var moduleDataRoot = sync.OnceValue(func() string {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return dataLocalesDir
	}
	return filepath.Join(filepath.Dir(currentFile), dataLocalesDir)
})

// resolveDataPath maps a (root, locale, fileName) tuple to a file path.
// An empty root means the bundled module data.
func resolveDataPath(root, locale, fileName string) string {
	if root == "" {
		root = moduleDataRoot()
	}
	return filepath.Join(root, locale, filepath.FromSlash(fileName))
}

// LocaleData holds the decoded contents of one (fileName, locale) data
// resource.
//
// Concurrency: the resource is loaded at most once (sync.Once). The
// payload is written only inside that Once and never changed afterwards,
// so readers that reach it through Load/EnsureLoaded need no locking.
type LocaleData struct {
	locale   string
	fileName string
	root     string
	// owner is the loader that created this value. It is consulted at load
	// time for the map/array decoding mode.
	owner *DataLoader

	once   sync.Once
	loaded atomic.Bool
	err    error
	isMap  bool

	// Most data is a JSON object of string arrays.
	data map[string][]string
	// Some data is an array of flat objects (e.g. currency, car models).
	dataMap []map[string]string
}

// Locale returns the locale this data belongs to.
func (l *LocaleData) Locale() string { return l.locale }

// FileName returns the data file name (relative to the locale directory).
func (l *LocaleData) FileName() string { return l.fileName }

// Loaded reports whether a load attempt has finished, successfully or not.
func (l *LocaleData) Loaded() bool { return l.loaded.Load() }

// Err returns the error from the load attempt, or nil if the load
// succeeded or has not run yet.
func (l *LocaleData) Err() error {
	if !l.loaded.Load() {
		return nil
	}
	return l.err
}

// IsMap reports whether the resource was decoded as an array of maps.
// It is only meaningful once Loaded returns true.
func (l *LocaleData) IsMap() bool {
	if !l.loaded.Load() {
		return false
	}
	return l.isMap
}

// Load reads and decodes the resource. Only the first call does any work.
// Every call, including concurrent ones, waits for that work to finish
// and returns the same error.
func (l *LocaleData) Load() error {
	_, err := l.loadOnce()
	return err
}

// loadOnce runs the load at most once. first is true only for the call
// that actually did the load, so callers can report a failure once instead
// of on every access. err is the (shared) load error.
func (l *LocaleData) loadOnce() (first bool, err error) {
	l.once.Do(func() {
		first = true
		l.err = l.load()
		l.loaded.Store(true)
	})
	return first, l.err
}

// load does the file I/O and decoding. It must only be called from
// within l.once.
func (l *LocaleData) load() error {
	if l.fileName == "" || l.locale == "" {
		return ErrLoaderNotInitialized
	}

	if l.owner != nil {
		l.isMap = l.owner.IsMap(l.locale)
	}

	raw, err := os.ReadFile(resolveDataPath(l.root, l.locale, l.fileName))
	if err != nil {
		return err
	}

	if l.isMap {
		var dataMap []map[string]string
		if err := json.Unmarshal(raw, &dataMap); err != nil {
			return fmt.Errorf("decoding %s/%s: %w", l.locale, l.fileName, err)
		}
		l.dataMap = dataMap
		return nil
	}

	var data map[string][]string
	if err := json.Unmarshal(raw, &data); err != nil {
		return fmt.Errorf("decoding %s/%s: %w", l.locale, l.fileName, err)
	}
	l.data = data
	return nil
}

// arrayData returns the decoded object-of-arrays payload, or nil if the
// resource has not finished loading. Checking the atomic loaded flag
// first makes every accessor race-free, even on a *LocaleData obtained
// from DataLoader.Get while another goroutine is loading it.
func (l *LocaleData) arrayData() map[string][]string {
	if !l.loaded.Load() {
		return nil
	}
	return l.data
}

// mapData returns the decoded array-of-maps payload, or nil if the
// resource has not finished loading.
func (l *LocaleData) mapData() []map[string]string {
	if !l.loaded.Load() {
		return nil
	}
	return l.dataMap
}

// Get returns the string array stored under key. It returns nil and logs
// if the key is missing.
func (l *LocaleData) Get(key string) []string {
	val, exists := l.arrayData()[key]
	if !exists {
		log.Printf("error - key doesnt exist %s\n", key)
		return nil
	}
	return val
}

// Has reports whether key exists in the (array-of-strings) data.
func (l *LocaleData) Has(key string) bool {
	_, ok := l.arrayData()[key]
	return ok
}

// Keys returns the top-level keys of the (array-of-strings) data.
func (l *LocaleData) Keys() []string {
	data := l.arrayData()
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	return keys
}

// Len returns the number of top-level keys for array data, or the number
// of records for map data.
func (l *LocaleData) Len() int {
	if l.IsMap() {
		return len(l.mapData())
	}
	return len(l.arrayData())
}

// RandomWeightedItem returns one random record from map-based data, or
// nil if there are no records.
func (l *LocaleData) RandomWeightedItem(f *Fakery) map[string]string {
	dataMap := l.mapData()
	if len(dataMap) == 0 {
		return nil
	}
	return dataMap[f.IntRange(len(dataMap))]
}

// GetWeightedArray parses the entries under key, each in the form
// "<item><sep><weight>", into a validated WeightedArray.
func (l *LocaleData) GetWeightedArray(key, sep string) (*WeightedArray, error) {
	var dataArray WeightedArray

	for _, dataItem := range l.Get(key) {
		items := strings.SplitN(dataItem, sep, 2)
		if len(items) != 2 {
			return nil, fmt.Errorf("weighted array[key: %s] malformed item %q", key, dataItem)
		}
		weight, err := strconv.ParseFloat(items[1], 64)
		if err != nil {
			return nil, err
		}
		dataArray.Items = append(dataArray.Items,
			WeightedItem{Item: items[0], Weight: weight})
	}

	if ok, val := dataArray.Validate(); !ok {
		return nil, fmt.Errorf("weighted array[key: %s] didn't validate - weight %.2f", key, val)
	}

	return &dataArray, nil
}

// Decode decodes the loaded data into result (a pointer) using its JSON
// tags. Array data goes into a struct or map; map data goes into a slice.
func (l *LocaleData) Decode(result interface{}) error {
	if err := l.Err(); err != nil {
		return err
	}
	var payload interface{} = l.arrayData()
	if l.IsMap() {
		payload = l.mapData()
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, result)
}

// DataLoader manages every locale of a single data file.
//
// The zero value is ready to use once Init has been called. All methods
// are safe for concurrent use. A DataLoader must not be copied after first
// use.
type DataLoader struct {
	mu sync.RWMutex
	// Data root directory; empty means the bundled module data.
	root string
	// Common file name for a specific type of data.
	fileName      string
	localeDataMap map[string]*LocaleData
	configIsMap   map[string]bool
}

// NewDataLoader returns a loader for fileName, initialised and ready to use.
func NewDataLoader(fileName string) *DataLoader {
	l := &DataLoader{}
	l.Init(fileName)
	return l
}

// Init (re)initialises the loader for fileName and drops any cached
// locale data. Every piece of data is identified by a unique
// {locale, fileName} tuple.
func (loader *DataLoader) Init(fileName string) {
	loader.mu.Lock()
	defer loader.mu.Unlock()
	loader.fileName = fileName
	loader.localeDataMap = make(map[string]*LocaleData)
	loader.configIsMap = make(map[string]bool)
}

// FileName returns the data file name this loader manages.
func (loader *DataLoader) FileName() string {
	loader.mu.RLock()
	defer loader.mu.RUnlock()
	return loader.fileName
}

// setRoot sets the data root directory. It is used by Registry.
func (loader *DataLoader) setRoot(root string) {
	loader.mu.Lock()
	defer loader.mu.Unlock()
	loader.root = root
}

// ensureMapsLocked lazily allocates the internal maps so a zero-value
// loader never panics. The write lock must be held.
func (loader *DataLoader) ensureMapsLocked() {
	if loader.localeDataMap == nil {
		loader.localeDataMap = make(map[string]*LocaleData)
	}
	if loader.configIsMap == nil {
		loader.configIsMap = make(map[string]bool)
	}
}

// SetIsMap marks the locale's data as an array of maps instead of an
// object of string arrays. Call it before the locale is first loaded;
// calling it afterwards does not affect data that is already decoded.
func (loader *DataLoader) SetIsMap(locale string) {
	loader.mu.Lock()
	defer loader.mu.Unlock()
	loader.ensureMapsLocked()
	loader.configIsMap[locale] = true
}

// IsMap reports whether the locale is configured for map-based decoding.
func (loader *DataLoader) IsMap(locale string) bool {
	loader.mu.RLock()
	defer loader.mu.RUnlock()
	return loader.configIsMap[locale]
}

// Get returns the shared, possibly not yet loaded, LocaleData for locale,
// creating it if needed. Concurrent callers always get the same pointer.
func (loader *DataLoader) Get(locale string) *LocaleData {
	// Fast path: shared lock only.
	loader.mu.RLock()
	val, ok := loader.localeDataMap[locale]
	loader.mu.RUnlock()
	if ok {
		return val
	}

	// Slow path: re-check under the write lock, in case another goroutine
	// created it in the meantime.
	loader.mu.Lock()
	defer loader.mu.Unlock()
	loader.ensureMapsLocked()
	if val, ok = loader.localeDataMap[locale]; ok {
		return val
	}

	val = &LocaleData{
		locale:   locale,
		fileName: loader.fileName,
		root:     loader.root,
		owner:    loader,
	}
	loader.localeDataMap[locale] = val
	return val
}

// EnsureLoaded returns the LocaleData for locale and loads it if it has
// not been loaded yet. Data is read from disk at most once per locale and
// then stays in memory. A load failure is logged once and the returned
// value is empty; its Err method reports the cause.
func (loader *DataLoader) EnsureLoaded(locale string) *LocaleData {
	localeData := loader.Get(locale)
	if first, err := localeData.loadOnce(); err != nil && first {
		log.Printf("error - loading locale data for locale: %s - %v\n", locale, err)
	}
	return localeData
}

// Preload loads data for locale and returns a shallow copy of its keys as
// a generic map. It returns nil if the locale data cannot be obtained.
func (loader *DataLoader) Preload(locale string) map[string]interface{} {
	localeData := loader.EnsureLoaded(locale)
	if localeData == nil {
		return nil
	}

	source := localeData.arrayData()
	data := make(map[string]interface{}, len(source))
	for key, val := range source {
		data[key] = val
	}
	return data
}

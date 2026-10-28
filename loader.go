package fakery

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
)

const prefix = "data/locales"

var (
	moduleDirOnce sync.Once
	moduleDirPath string
	moduleDirErr  error
)

func resolveModuleDir() (string, error) {
	moduleDirOnce.Do(func() {
		_, currentFile, _, ok := runtime.Caller(0)
		if !ok {
			moduleDirErr = os.ErrNotExist
			return
		}
		moduleDirPath = filepath.Dir(currentFile)
	})
	return moduleDirPath, moduleDirErr
}

func defaultLocalesDir() (string, error) {
	dir, err := resolveModuleDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, filepath.FromSlash(prefix)), nil
}

// getModuleFile returns the full path to a file relative to the module directory.
func getModuleFile(relativePath string) (string, error) {
	dir, err := resolveModuleDir()
	if err != nil {
		return "", err
	}

	fullPath := filepath.Join(dir, relativePath)
	if _, err := os.Stat(fullPath); err != nil {
		return "", err
	}

	return fullPath, nil
}

// WeightedItem represents a single string option and its selection probability weight.
type WeightedItem struct {
	Item   string  `json:"item"`
	Weight float64 `json:"weight"`
}

// WeightedArray holds a collection of weighted items whose weights sum to 1.0.
type WeightedArray struct {
	Items []WeightedItem
}

// Validate reports whether the weights in the array sum to 1.0 (rounded to 2 decimals).
func (w WeightedArray) Validate() (bool, float64) {
	var cumWeight float64
	for _, item := range w.Items {
		cumWeight += item.Weight
	}

	cumWeight = math.Round(cumWeight*100) / 100
	if cumWeight == 1.0 {
		return true, cumWeight
	}
	return false, cumWeight
}

func cloneWeightedArray(src *WeightedArray) *WeightedArray {
	if src == nil {
		return nil
	}
	items := make([]WeightedItem, len(src.Items))
	copy(items, src.Items)
	return &WeightedArray{Items: items}
}

// LocaleData holds lazily decoded data for a single (resource, locale) pair.
// Loading runs at most once per LocaleData instance via sync.Once and all
// read operations are safe for concurrent access across goroutines.
type LocaleData struct {
	id        uuid.UUID
	once      sync.Once
	logOnce   sync.Once
	mu        sync.RWMutex
	locale    string
	fileName  string
	baseDir   string
	data      map[string][]string
	dataMap   []map[string]string
	weighted  map[string]*WeightedArray
	loadErr   error
	loaded    atomic.Bool
	isMap     atomic.Bool
	loadCount atomic.Int64
}

// ID returns the unique identifier assigned to this cached LocaleData entry.
func (l *LocaleData) ID() uuid.UUID {
	if l == nil {
		return uuid.UUID{}
	}
	return l.id
}

// Locale returns the locale key associated with this LocaleData entry.
func (l *LocaleData) Locale() string {
	if l == nil {
		return ""
	}
	return l.locale
}

// FileName returns the resource file path associated with this LocaleData entry.
func (l *LocaleData) FileName() string {
	if l == nil {
		return ""
	}
	return l.fileName
}

// IsMap reports whether this LocaleData decodes JSON arrays of string maps.
func (l *LocaleData) IsMap() bool {
	if l == nil {
		return false
	}
	return l.isMap.Load()
}

// IsLoaded reports whether Load has completed for this LocaleData entry.
func (l *LocaleData) IsLoaded() bool {
	if l == nil {
		return false
	}
	return l.loaded.Load()
}

// LoadCount returns how many times the underlying file load routine was executed.
func (l *LocaleData) LoadCount() int64 {
	if l == nil {
		return 0
	}
	return l.loadCount.Load()
}

// Err returns any error encountered when loading this LocaleData entry.
func (l *LocaleData) Err() error {
	if l == nil {
		return nil
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.loadErr
}

func (l *LocaleData) load() error {
	l.loadCount.Add(1)

	baseDir := l.baseDir
	if baseDir == "" {
		dir, err := defaultLocalesDir()
		if err != nil {
			return err
		}
		baseDir = dir
	}

	relPath := filepath.Clean(filepath.FromSlash(path.Join(l.locale, l.fileName)))
	if relPath == "." || relPath == ".." || strings.HasPrefix(relPath, ".."+string(filepath.Separator)) {
		return os.ErrNotExist
	}

	filePath := filepath.Join(baseDir, relPath)
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	if l.isMap.Load() {
		var parsed []map[string]string
		if err := json.Unmarshal(raw, &parsed); err != nil {
			return fmt.Errorf("decoding map resource %s (%s): %w", l.fileName, l.locale, err)
		}
		l.mu.Lock()
		l.dataMap = parsed
		l.mu.Unlock()
		return nil
	}

	var parsed map[string][]string
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return fmt.Errorf("decoding resource %s (%s): %w", l.fileName, l.locale, err)
	}
	l.mu.Lock()
	l.data = parsed
	l.mu.Unlock()
	return nil
}

// Load decodes the underlying JSON resource once and caches the result or error.
func (l *LocaleData) Load() error {
	if l == nil || l.fileName == "" || l.locale == "" {
		return fmt.Errorf("error - loader not initialized")
	}

	l.once.Do(func() {
		err := l.load()
		l.mu.Lock()
		l.loadErr = err
		l.mu.Unlock()
		l.loaded.Store(true)
	})

	l.mu.RLock()
	err := l.loadErr
	l.mu.RUnlock()
	return err
}

// Get returns a copy of the string slice stored under key, or nil if absent.
func (l *LocaleData) Get(key string) []string {
	if l == nil {
		return nil
	}

	l.mu.RLock()
	val, exists := l.data[key]
	l.mu.RUnlock()

	if !exists {
		log.Printf("error - key doesnt exist %s\n", key)
		return nil
	}

	out := make([]string, len(val))
	copy(out, val)
	return out
}

// Entries returns a defensive copy of all map-decoded records for map-based resources.
func (l *LocaleData) Entries() []map[string]string {
	if l == nil {
		return nil
	}

	l.mu.RLock()
	src := l.dataMap
	l.mu.RUnlock()

	if len(src) == 0 {
		return nil
	}

	out := make([]map[string]string, len(src))
	for i, item := range src {
		cloned := make(map[string]string, len(item))
		for k, v := range item {
			cloned[k] = v
		}
		out[i] = cloned
	}
	return out
}

// RandomWeightedItem selects a random map record from map-based locale data.
func (l *LocaleData) RandomWeightedItem(f *Fakery) map[string]string {
	if l == nil || f == nil {
		return nil
	}

	l.mu.RLock()
	dataMap := l.dataMap
	l.mu.RUnlock()

	if len(dataMap) == 0 {
		return nil
	}

	src := dataMap[f.IntRange(len(dataMap))]
	cloned := make(map[string]string, len(src))
	for k, v := range src {
		cloned[k] = v
	}
	return cloned
}

// GetWeightedArray fetches key and parses "item<sep>weight" strings into a validated WeightedArray.
func (l *LocaleData) GetWeightedArray(key, sep string) (*WeightedArray, error) {
	if l == nil {
		return nil, fmt.Errorf("locale data is nil")
	}

	cacheKey := key + "\x00" + sep

	l.mu.RLock()
	if l.weighted != nil {
		if cached, ok := l.weighted[cacheKey]; ok {
			l.mu.RUnlock()
			return cloneWeightedArray(cached), nil
		}
	}
	dataItems, exists := l.data[key]
	l.mu.RUnlock()

	if !exists {
		log.Printf("error - key doesnt exist %s\n", key)
		return nil, fmt.Errorf("weighted array key %q does not exist", key)
	}

	var dataArray WeightedArray
	for _, dataItem := range dataItems {
		items := strings.Split(dataItem, sep)
		if len(items) != 2 {
			return nil, fmt.Errorf("invalid weighted item %q for key %q", dataItem, key)
		}
		weight, err := strconv.ParseFloat(items[1], 64)
		if err != nil {
			return nil, err
		}
		dataArray.Items = append(dataArray.Items, WeightedItem{
			Item:   items[0],
			Weight: weight,
		})
	}

	if ok, val := dataArray.Validate(); !ok {
		return nil, fmt.Errorf("weighted array[key: %s] didn't validate - weight %.2f", key, val)
	}

	l.mu.Lock()
	if l.weighted == nil {
		l.weighted = make(map[string]*WeightedArray)
	}
	l.weighted[cacheKey] = &dataArray
	l.mu.Unlock()

	return cloneWeightedArray(&dataArray), nil
}

// DataLoader manages cached LocaleData instances per locale for a single resource file.
// All DataLoader methods are concurrency-safe.
type DataLoader struct {
	mu            sync.RWMutex
	localeDataMap map[string]*LocaleData
	configIsMap   map[string]bool
	defaultIsMap  bool
	fileName      string
	baseDir       string
}

// NewDataLoader creates a concurrency-safe DataLoader for the given resource filename.
func NewDataLoader(fileName string) *DataLoader {
	return NewDataLoaderWithBaseDir(fileName, "")
}

// NewMapDataLoader creates a concurrency-safe DataLoader configured for map-slice JSON decoding.
func NewMapDataLoader(fileName string) *DataLoader {
	loader := NewDataLoader(fileName)
	loader.SetDefaultIsMap(true)
	return loader
}

// NewDataLoaderWithBaseDir creates a DataLoader rooted at baseDir (or default data/locales if empty).
func NewDataLoaderWithBaseDir(fileName, baseDir string) *DataLoader {
	return &DataLoader{
		localeDataMap: make(map[string]*LocaleData),
		configIsMap:   make(map[string]bool),
		fileName:      fileName,
		baseDir:       baseDir,
	}
}

func (loader *DataLoader) ensureMapsLocked() {
	if loader.localeDataMap == nil {
		loader.localeDataMap = make(map[string]*LocaleData)
	}
	if loader.configIsMap == nil {
		loader.configIsMap = make(map[string]bool)
	}
}

// Init initializes or resets the DataLoader for filePath.
func (loader *DataLoader) Init(filePath string) {
	loader.mu.Lock()
	defer loader.mu.Unlock()

	loader.fileName = filePath
	loader.localeDataMap = make(map[string]*LocaleData)
	loader.configIsMap = make(map[string]bool)
}

// FileName returns the resource filename configured on this DataLoader.
func (loader *DataLoader) FileName() string {
	loader.mu.RLock()
	defer loader.mu.RUnlock()
	return loader.fileName
}

// SetIsMap marks a specific locale's JSON resource as an array of string maps.
func (loader *DataLoader) SetIsMap(locale string) {
	loader.mu.Lock()
	defer loader.mu.Unlock()

	loader.ensureMapsLocked()
	loader.configIsMap[locale] = true
	if existing, ok := loader.localeDataMap[locale]; ok && !existing.loaded.Load() {
		existing.isMap.Store(true)
	}
}

// SetDefaultIsMap configures whether all locales for this resource decode as map slices.
func (loader *DataLoader) SetDefaultIsMap(isMap bool) {
	loader.mu.Lock()
	defer loader.mu.Unlock()

	loader.defaultIsMap = isMap
	if isMap {
		for _, existing := range loader.localeDataMap {
			if !existing.loaded.Load() {
				existing.isMap.Store(true)
			}
		}
	}
}

// Preload loads locale data and returns a shallow map[string]interface{} view of its keys.
func (loader *DataLoader) Preload(locale string) map[string]interface{} {
	localeData := loader.EnsureLoaded(locale)
	if localeData == nil {
		return nil
	}

	localeData.mu.RLock()
	defer localeData.mu.RUnlock()

	data := make(map[string]interface{}, len(localeData.data))
	for key, val := range localeData.data {
		copied := make([]string, len(val))
		copy(copied, val)
		data[key] = copied
	}

	return data
}

// Get returns or initializes the cached LocaleData instance for locale without loading disk data.
func (loader *DataLoader) Get(locale string) *LocaleData {
	loader.mu.RLock()
	if loader.localeDataMap != nil {
		if val, ok := loader.localeDataMap[locale]; ok {
			loader.mu.RUnlock()
			return val
		}
	}
	loader.mu.RUnlock()

	loader.mu.Lock()
	defer loader.mu.Unlock()

	loader.ensureMapsLocked()
	if val, ok := loader.localeDataMap[locale]; ok {
		return val
	}

	newVal := &LocaleData{
		id:       uuid.New(),
		locale:   locale,
		fileName: loader.fileName,
		baseDir:  loader.baseDir,
		weighted: make(map[string]*WeightedArray),
	}
	newVal.isMap.Store(loader.defaultIsMap || loader.configIsMap[locale])
	loader.localeDataMap[locale] = newVal

	return newVal
}

// EnsureLoaded lazily loads and caches locale data for the exact requested locale.
func (loader *DataLoader) EnsureLoaded(locale string) *LocaleData {
	localeData := loader.Get(locale)
	if localeData == nil {
		log.Fatalf("error - locale data not initialized for locale %s", locale)
		return nil
	}

	if err := localeData.Load(); err != nil {
		localeData.logOnce.Do(func() {
			log.Printf("error - loading locale data for locale: %s - %v\n", locale, err)
		})
	}

	return localeData
}

func localeFallbackChain(locale string) []string {
	candidates := make([]string, 0, 4)
	add := func(loc string) {
		if loc == "" {
			return
		}
		for _, existing := range candidates {
			if existing == loc {
				return
			}
		}
		candidates = append(candidates, loc)
	}

	add(locale)
	if idx := strings.IndexByte(locale, '_'); idx > 0 {
		add(locale[:idx])
	}
	add(DefaultLocale)
	add(GenericLocale)
	return candidates
}

// ResolveLocale loads locale data for locale, falling back through its base language,
// DefaultLocale, and GenericLocale only when a locale-specific file does not exist.
// Malformed JSON errors are preserved and never masked by fallback.
func (loader *DataLoader) ResolveLocale(locale string) *LocaleData {
	candidates := localeFallbackChain(locale)
	var primary *LocaleData

	for i, candidate := range candidates {
		data := loader.Get(candidate)
		err := data.Load()
		if i == 0 {
			primary = data
		}
		if err == nil {
			return data
		}
		if !errors.Is(err, os.ErrNotExist) {
			data.logOnce.Do(func() {
				log.Printf("error - loading locale data for locale: %s - %v\n", candidate, err)
			})
			return data
		}
	}

	if primary != nil {
		primary.logOnce.Do(func() {
			log.Printf("error - loading locale data for locale: %s - %v\n", primary.locale, primary.Err())
		})
		return primary
	}

	return loader.EnsureLoaded(DefaultLocale)
}

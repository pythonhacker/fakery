package fakery

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Built-in locale data resource identifiers.
const (
	ResourceAddress  = "address.json"
	ResourceBeer     = "beer.json"
	ResourceBook     = "book.json"
	ResourceCar      = "car.json"
	ResourceColor    = "color.json"
	ResourceCurrency = "currency.json"
	ResourceInternet = "internet.json"
	ResourceJobs     = "jobs.json"
	ResourceNames    = "names.json"
	ResourceWine     = "wine.json"
	ResourceWords    = "words.json"

	carModelsPrefix = "cars/"
)

// ResourceConfig defines the registration metadata for a locale data resource.
type ResourceConfig struct {
	FileName   string
	IsMap      bool
	MapLocales []string
}

// ResourceRule matches dynamic resource names (such as "cars/<maker>.json") to a ResourceConfig.
type ResourceRule func(resource string) (ResourceConfig, bool)

// LoaderRegistry provides central, concurrency-safe registration and lookup of DataLoaders
// keyed by resource filename.
type LoaderRegistry struct {
	mu      sync.RWMutex
	baseDir string
	loaders map[string]*DataLoader
	configs map[string]ResourceConfig
	rules   []ResourceRule
}

var defaultRegistry = newDefaultRegistry("")

// NewLoaderRegistry creates a new empty LoaderRegistry rooted at baseDir
// (or the package's default data/locales directory when baseDir is empty).
func NewLoaderRegistry(baseDir string) *LoaderRegistry {
	return &LoaderRegistry{
		baseDir: baseDir,
		loaders: make(map[string]*DataLoader),
		configs: make(map[string]ResourceConfig),
	}
}

// NewDefaultLoaderRegistry creates a LoaderRegistry pre-populated with all built-in
// fakery resource registrations and dynamic map rules, rooted at baseDir.
func NewDefaultLoaderRegistry(baseDir string) *LoaderRegistry {
	return newDefaultRegistry(baseDir)
}

func newDefaultRegistry(baseDir string) *LoaderRegistry {
	reg := NewLoaderRegistry(baseDir)

	builtInResources := []ResourceConfig{
		{FileName: ResourceAddress},
		{FileName: ResourceBeer},
		{FileName: ResourceBook},
		{FileName: ResourceCar},
		{FileName: ResourceColor},
		{FileName: ResourceCurrency, IsMap: true, MapLocales: []string{GenericLocale}},
		{FileName: ResourceInternet},
		{FileName: ResourceJobs},
		{FileName: ResourceNames},
		{FileName: ResourceWine},
		{FileName: ResourceWords},
	}

	for _, cfg := range builtInResources {
		reg.Register(cfg)
	}

	// Register car maker model catalogs ("cars/<maker>.json") as map-based resources.
	reg.RegisterRule(func(resource string) (ResourceConfig, bool) {
		if strings.HasPrefix(resource, carModelsPrefix) && strings.HasSuffix(resource, ".json") {
			return ResourceConfig{
				FileName:   resource,
				IsMap:      true,
				MapLocales: []string{GenericLocale},
			}, true
		}
		return ResourceConfig{}, false
	})

	return reg
}

// DefaultRegistry returns the package-wide central LoaderRegistry.
func DefaultRegistry() *LoaderRegistry {
	return defaultRegistry
}

func (r *LoaderRegistry) buildLoaderLocked(cfg ResourceConfig) *DataLoader {
	loader := NewDataLoaderWithBaseDir(cfg.FileName, r.baseDir)
	if cfg.IsMap {
		loader.SetDefaultIsMap(true)
	}
	for _, loc := range cfg.MapLocales {
		loader.SetIsMap(loc)
	}
	return loader
}

// Register registers a resource configuration in the registry once and returns its DataLoader.
// If the resource is already registered, its existing DataLoader is updated with any map
// settings and returned without discarding cached locale data.
func (r *LoaderRegistry) Register(cfg ResourceConfig) *DataLoader {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.configs[cfg.FileName] = cfg
	if existing, ok := r.loaders[cfg.FileName]; ok {
		if cfg.IsMap {
			existing.SetDefaultIsMap(true)
		}
		for _, loc := range cfg.MapLocales {
			existing.SetIsMap(loc)
		}
		return existing
	}

	loader := r.buildLoaderLocked(cfg)
	r.loaders[cfg.FileName] = loader
	return loader
}

// RegisterRule adds a dynamic matching rule used when resolving previously unseen resources.
func (r *LoaderRegistry) RegisterRule(rule ResourceRule) {
	if rule == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rules = append(r.rules, rule)
}

// Has reports whether a DataLoader has been registered or instantiated for resource.
func (r *LoaderRegistry) Has(resource string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.loaders[resource]
	return ok
}

// RegisteredResources returns a sorted snapshot of registered resource filenames.
func (r *LoaderRegistry) RegisteredResources() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.loaders))
	for name := range r.loaders {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Loader returns the central DataLoader for resource, creating and caching it once if needed.
func (r *LoaderRegistry) Loader(resource string) *DataLoader {
	r.mu.RLock()
	if loader, ok := r.loaders[resource]; ok {
		r.mu.RUnlock()
		return loader
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if loader, ok := r.loaders[resource]; ok {
		return loader
	}

	cfg, ok := r.configs[resource]
	if !ok {
		for _, rule := range r.rules {
			if matchedCfg, matched := rule(resource); matched {
				cfg = matchedCfg
				ok = true
				break
			}
		}
	}
	if !ok {
		cfg = ResourceConfig{FileName: resource}
	}

	loader := r.buildLoaderLocked(cfg)
	r.configs[resource] = cfg
	r.loaders[resource] = loader
	return loader
}

// LoadLocale resolves and lazily loads locale-specific data (with locale fallback) for resource.
func (r *LoaderRegistry) LoadLocale(resource, locale string) *LocaleData {
	return r.Loader(resource).ResolveLocale(locale)
}

// LoadGeneric lazily loads generic locale data for resource.
func (r *LoaderRegistry) LoadGeneric(resource string) *LocaleData {
	return r.Loader(resource).EnsureLoaded(GenericLocale)
}

func carModelResource(makeName string) string {
	return fmt.Sprintf("%s%s.json", carModelsPrefix, NormalizeString(makeName))
}

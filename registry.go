// Central registry of locale data sources.
package fakery

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
)

// ErrSourceNotRegistered is returned (through LocaleData.Err) when a file
// name that matches no registered source or prefix family is requested.
var ErrSourceNotRegistered = errors.New("error - data source not registered")

// SourceOption configures a data source at registration time.
type SourceOption func(*DataLoader)

// WithMapLocales marks the source as map based (a JSON array of flat
// objects) for the given locales.
func WithMapLocales(locales ...string) SourceOption {
	return func(l *DataLoader) {
		for _, locale := range locales {
			l.SetIsMap(locale)
		}
	}
}

// sourceFamily registers every file name under a common prefix with the
// same options, e.g. "cars/" for the per-maker model files.
type sourceFamily struct {
	prefix string
	opts   []SourceOption
}

// Registry maps data file names to their DataLoader. It is the single
// place where data sources are declared and the single owner of their
// cached data. All methods are safe for concurrent use.
type Registry struct {
	mu       sync.RWMutex
	root     string
	loaders  map[string]*DataLoader
	families []sourceFamily
}

// NewRegistry returns an empty registry that reads the bundled data.
func NewRegistry() *Registry {
	return NewRegistryWithRoot("")
}

// NewRegistryWithRoot returns an empty registry that reads files from
// <root>/<locale>/<fileName>. An empty root means the bundled data. This
// is useful for tests and for applications that ship their own data.
func NewRegistryWithRoot(root string) *Registry {
	return &Registry{
		root:    root,
		loaders: make(map[string]*DataLoader),
	}
}

// Root returns the data root directory ("" means the bundled data).
func (r *Registry) Root() string { return r.root }

// Register declares the data source fileName and returns its loader.
// Registering is idempotent: a second call for the same name returns the
// same loader (and so the same cache), after applying any extra options.
//
// A new loader is fully configured before it is published in the
// registry, so no concurrent Load can see it with only some of its options
// applied (e.g. decode map data before WithMapLocales is set).
func (r *Registry) Register(fileName string, opts ...SourceOption) *DataLoader {
	r.mu.Lock()
	defer r.mu.Unlock()

	loader, ok := r.loaders[fileName]
	if !ok {
		loader = NewDataLoader(fileName)
		loader.SetRoot(r.root)
	}
	// Options only lock the loader (never the registry), so applying them
	// while holding r.mu cannot deadlock.
	for _, opt := range opts {
		opt(loader)
	}
	if !ok {
		r.loaders[fileName] = loader
	}
	return loader
}

// RegisterPrefix declares a family of data sources: any file name that
// starts with prefix is registered on first use with opts. Use it for
// resources whose names are only known at runtime.
func (r *Registry) RegisterPrefix(prefix string, opts ...SourceOption) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.families = append(r.families, sourceFamily{prefix: prefix, opts: opts})
}

// Loader returns the loader for fileName. Names that match a registered
// prefix family are registered on demand. ok is false if fileName is
// unknown.
func (r *Registry) Loader(fileName string) (loader *DataLoader, ok bool) {
	r.mu.RLock()
	loader, ok = r.loaders[fileName]
	var family *sourceFamily
	if !ok {
		// Prefer the longest matching prefix.
		for i := range r.families {
			f := &r.families[i]
			if strings.HasPrefix(fileName, f.prefix) &&
				(family == nil || len(f.prefix) > len(family.prefix)) {
				family = f
			}
		}
	}
	r.mu.RUnlock()

	if ok {
		return loader, true
	}
	if family == nil {
		return nil, false
	}
	// Register handles a concurrent registration of the same name.
	return r.Register(fileName, family.opts...), true
}

// Sources returns the sorted names of all registered (or on-demand
// registered) data sources.
func (r *Registry) Sources() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.loaders))
	for name := range r.loaders {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Load returns the loaded data of fileName for locale. Each
// (fileName, locale) pair is read from disk at most once; later and
// concurrent calls share the cached *LocaleData.
//
// Load never returns nil. If the source is unknown or fails to load, the
// returned value is empty and its Err method reports why.
func (r *Registry) Load(fileName, locale string) *LocaleData {
	loader, ok := r.Loader(fileName)
	if !ok {
		log.Printf("error - %v: %s\n", ErrSourceNotRegistered, fileName)
		return failedLocaleData(fileName, locale,
			fmt.Errorf("%w: %s", ErrSourceNotRegistered, fileName))
	}
	return loader.EnsureLoaded(locale)
}

// failedLocaleData builds an empty, already "loaded" LocaleData that
// carries err.
func failedLocaleData(fileName, locale string, err error) *LocaleData {
	l := &LocaleData{fileName: fileName, locale: locale}
	l.once.Do(func() {
		l.err = err
		l.loaded.Store(true)
	})
	return l
}

// DataView is a lazily decoded, typed, read-only view of one
// (source, locale) resource, e.g. *BeerData for beer.json. The first Get
// loads and decodes the resource. Later calls return the same pointer,
// and the value must be treated as read-only.
type DataView[T any] struct {
	registry *Registry
	source   string
	locale   string
	once     sync.Once
	value    *T
}

// NewDataView returns a view of source/locale in registry. Nothing is
// loaded until the first Get.
func NewDataView[T any](registry *Registry, source, locale string) *DataView[T] {
	return &DataView[T]{registry: registry, source: source, locale: locale}
}

// Get returns the decoded value. If the data cannot be loaded or decoded,
// the error is logged once and a zero value is returned.
func (v *DataView[T]) Get() *T {
	v.once.Do(func() {
		v.value = new(T)
		if err := v.registry.Load(v.source, v.locale).Decode(v.value); err != nil {
			log.Printf("error - decoding %s/%s: %v\n", v.locale, v.source, err)
		}
	})
	return v.value
}

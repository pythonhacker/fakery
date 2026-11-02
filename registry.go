// Central registry for locale data resources.
//
// Every data file shipped with fakery (or supplied by a caller) is described
// by a ResourceSpec and registered exactly once in a Registry. The registry
// owns the cache of decoded data, keyed by (resource name, locale), and
// guarantees that each combination is read and decoded at most once, no
// matter how many goroutines request it concurrently.
package fakery

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
)

// Root of the bundled data, relative to the module directory.
const prefix = "data/locales"

// ResourceKind describes how the JSON payload of a resource is decoded.
type ResourceKind int

const (
	// ListResource is a JSON object mapping keys to string arrays,
	// e.g. {"states": ["Alabama", ...]}. This is the default kind.
	ListResource ResourceKind = iota
	// MapResource is a JSON array of string maps,
	// e.g. [{"code": "INR", "currency": "Rupee"}, ...].
	MapResource
)

func (k ResourceKind) String() string {
	switch k {
	case ListResource:
		return "list"
	case MapResource:
		return "map"
	}
	return fmt.Sprintf("ResourceKind(%d)", int(k))
}

// ResourceSpec describes a data resource. Name is a slash separated file
// name relative to a locale directory (e.g. "address.json"). A Name that
// contains glob meta characters ("*", "?", "[") is a pattern and applies
// to every file it matches (e.g. "cars/*.json").
type ResourceSpec struct {
	Name string
	Kind ResourceKind
}

func (s ResourceSpec) isPattern() bool {
	return strings.ContainsAny(s.Name, "*?[")
}

var (
	// ErrInvalidResource is returned for empty or malformed resource names
	// and locales.
	ErrInvalidResource = errors.New("fakery: invalid resource")
	// ErrResourceConflict is returned when a resource is registered again
	// with a different kind.
	ErrResourceConflict = errors.New("fakery: conflicting resource registration")
	// ErrResourceLoaded is returned when trying to change the kind of a
	// resource which has already been loaded.
	ErrResourceLoaded = errors.New("fakery: resource already loaded")
)

type cacheKey struct {
	name   string
	locale string
}

// A cache slot. The once guards the single load of the slot; data and err
// are written inside the once and are read-only afterwards.
type cacheEntry struct {
	once sync.Once
	data *LocaleData
	err  error
}

// Registry maps resource names to their specs and caches decoded data per
// (resource, locale). A Registry is safe for concurrent use.
type Registry struct {
	fsys fs.FS

	mu       sync.RWMutex
	specs    map[string]ResourceSpec
	patterns []ResourceSpec
	cache    map[cacheKey]*cacheEntry
}

// NewRegistry returns a registry reading data from fsys. The root of fsys
// must contain one directory per locale (e.g. "en_US/names.json",
// "generic/beer.json"). All builtin resources are registered.
func NewRegistry(fsys fs.FS) *Registry {
	r := &Registry{
		fsys:  fsys,
		specs: make(map[string]ResourceSpec),
		cache: make(map[cacheKey]*cacheEntry),
	}
	for _, spec := range BuiltinResources() {
		if err := r.Register(spec); err != nil {
			panic(err)
		}
	}
	return r
}

var defaultRegistry = NewRegistry(defaultDataFS())

// DefaultRegistry returns the process wide registry backed by the data
// bundled with the module. It is used by every Fakery which has not been
// given a registry explicitly.
func DefaultRegistry() *Registry {
	return defaultRegistry
}

// defaultDataFS resolves the bundled data directory relative to this
// source file, so data is found regardless of the working directory.
func defaultDataFS() fs.FS {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return os.DirFS(filepath.FromSlash(prefix))
	}
	return os.DirFS(filepath.Join(filepath.Dir(currentFile), filepath.FromSlash(prefix)))
}

// Register adds a resource spec. Registering the same spec twice is a
// no-op. Registering a name again with a different kind fails with
// ErrResourceConflict, and changing the kind of a resource which already
// has cached data fails with ErrResourceLoaded.
func (r *Registry) Register(spec ResourceSpec) error {
	if spec.Name == "" || !fs.ValidPath(spec.Name) {
		return fmt.Errorf("%w: name %q", ErrInvalidResource, spec.Name)
	}
	if spec.Kind != ListResource && spec.Kind != MapResource {
		return fmt.Errorf("%w: kind %v for %q", ErrInvalidResource, spec.Kind, spec.Name)
	}
	if spec.isPattern() {
		if _, err := path.Match(spec.Name, ""); err != nil {
			return fmt.Errorf("%w: pattern %q: %v", ErrInvalidResource, spec.Name, err)
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if spec.isPattern() {
		for _, p := range r.patterns {
			if p.Name == spec.Name {
				if p.Kind == spec.Kind {
					return nil
				}
				return fmt.Errorf("%w: %q is %v, not %v", ErrResourceConflict, spec.Name, p.Kind, spec.Kind)
			}
		}
		r.patterns = append(r.patterns, spec)
		return nil
	}

	if existing, ok := r.specs[spec.Name]; ok {
		if existing.Kind == spec.Kind {
			return nil
		}
		return fmt.Errorf("%w: %q is %v, not %v", ErrResourceConflict, spec.Name, existing.Kind, spec.Kind)
	}
	if r.specLocked(spec.Name).Kind != spec.Kind {
		for key := range r.cache {
			if key.name == spec.Name {
				return fmt.Errorf("%w: %q", ErrResourceLoaded, spec.Name)
			}
		}
	}
	r.specs[spec.Name] = spec
	return nil
}

// Spec returns the spec which applies to name: an exact registration,
// else the first matching pattern, else an implicit ListResource spec.
// The boolean reports whether the spec came from a registration.
func (r *Registry) Spec(name string) (ResourceSpec, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.lookupSpecLocked(name)
}

func (r *Registry) specLocked(name string) ResourceSpec {
	spec, _ := r.lookupSpecLocked(name)
	return spec
}

func (r *Registry) lookupSpecLocked(name string) (ResourceSpec, bool) {
	if spec, ok := r.specs[name]; ok {
		return spec, true
	}
	for _, p := range r.patterns {
		if p.Name == name {
			return p, true
		}
		if ok, _ := path.Match(p.Name, name); ok {
			return p, true
		}
	}
	return ResourceSpec{Name: name, Kind: ListResource}, false
}

// Resources returns all registered specs (exact names and patterns),
// sorted by name.
func (r *Registry) Resources() []ResourceSpec {
	r.mu.RLock()
	defer r.mu.RUnlock()
	specs := make([]ResourceSpec, 0, len(r.specs)+len(r.patterns))
	for _, spec := range r.specs {
		specs = append(specs, spec)
	}
	specs = append(specs, r.patterns...)
	sort.Slice(specs, func(i, j int) bool { return specs[i].Name < specs[j].Name })
	return specs
}

// Loaded reports whether data for (name, locale) has been requested and
// is cached.
func (r *Registry) Loaded(name, locale string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.cache[cacheKey{name, locale}]
	return ok
}

// Load returns the data of resource name for locale. It never returns nil:
// on failure the error is logged once and an empty LocaleData carrying the
// error (see LocaleData.Err) is returned, mirroring the historical
// DataLoader behaviour.
func (r *Registry) Load(name, locale string) *LocaleData {
	data, _ := r.Lookup(name, locale)
	return data
}

// Lookup returns the data of resource name for locale, loading and caching
// it on first use. Concurrent callers for the same (name, locale) block
// until the single load completes and then share its result; failures are
// cached as well.
//
// If the file does not exist for locale, Lookup falls back to
// DefaultLocale and then to GenericLocale. The generic locale never falls
// back. The returned data reports the locale it was actually read from.
func (r *Registry) Lookup(name, locale string) (*LocaleData, error) {
	if err := validateKey(name, locale); err != nil {
		log.Printf("error - %v\n", err)
		return &LocaleData{locale: locale, fileName: name, err: err}, err
	}

	entry := r.entry(cacheKey{name, locale})
	entry.once.Do(func() {
		entry.data, entry.err = r.resolve(name, locale)
		if entry.err != nil {
			log.Printf("error - loading locale data for locale: %s - %v\n", locale, entry.err)
		}
	})
	return entry.data, entry.err
}

func validateKey(name, locale string) error {
	if name == "" || locale == "" {
		return fmt.Errorf("%w: loader not initialized (name %q, locale %q)", ErrInvalidResource, name, locale)
	}
	if !fs.ValidPath(name) {
		return fmt.Errorf("%w: name %q", ErrInvalidResource, name)
	}
	if strings.Contains(locale, "/") || !fs.ValidPath(locale) {
		return fmt.Errorf("%w: locale %q", ErrInvalidResource, locale)
	}
	return nil
}

// entry returns the cache slot for key, creating it if needed.
func (r *Registry) entry(key cacheKey) *cacheEntry {
	r.mu.RLock()
	entry, ok := r.cache[key]
	r.mu.RUnlock()
	if ok {
		return entry
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if entry, ok = r.cache[key]; !ok {
		entry = &cacheEntry{}
		r.cache[key] = entry
	}
	return entry
}

// resolve reads (name, locale) and walks the fallback chain when the file
// is missing. Fallback locales are loaded through Lookup so that they are
// cached and shared too.
func (r *Registry) resolve(name, locale string) (*LocaleData, error) {
	data, err := r.read(name, locale)
	if err == nil || !errors.Is(err, fs.ErrNotExist) {
		return data, err
	}

	for _, fallback := range fallbackLocales(locale) {
		fbData, fbErr := r.Lookup(name, fallback)
		if fbErr == nil {
			return fbData, nil
		}
		if !errors.Is(fbErr, fs.ErrNotExist) {
			return fbData, fbErr
		}
	}
	return data, err
}

// fallbackLocales returns the next locale to try. Each fallback resolves
// its own chain, so one step is enough: locale -> default -> generic.
func fallbackLocales(locale string) []string {
	switch locale {
	case GenericLocale:
		return nil
	case DefaultLocale:
		return []string{GenericLocale}
	}
	return []string{DefaultLocale}
}

// read loads and decodes a single file without any fallback.
func (r *Registry) read(name, locale string) (*LocaleData, error) {
	r.mu.RLock()
	spec := r.specLocked(name)
	r.mu.RUnlock()

	data := &LocaleData{locale: locale, fileName: name, isMap: spec.Kind == MapResource}

	if r.fsys == nil {
		data.err = fmt.Errorf("fakery: no data source for %s/%s: %w", locale, name, fs.ErrNotExist)
		return data, data.err
	}

	raw, err := fs.ReadFile(r.fsys, path.Join(locale, name))
	if err != nil {
		data.err = fmt.Errorf("fakery: reading %s/%s: %w", locale, name, err)
		return data, data.err
	}

	if err = data.decode(raw); err != nil {
		data.err = fmt.Errorf("fakery: decoding %s/%s as %v resource: %w", locale, name, spec.Kind, err)
		return data, data.err
	}
	return data, nil
}

// decode fills the data from a raw JSON payload according to its kind.
// On failure the data is left empty rather than partially filled.
func (l *LocaleData) decode(raw []byte) error {
	if l.isMap {
		var items []map[string]string
		if err := json.Unmarshal(raw, &items); err != nil {
			return err
		}
		l.dataMap = items
		return nil
	}

	var keyed map[string][]string
	if err := json.Unmarshal(raw, &keyed); err != nil {
		return err
	}
	l.data = keyed
	return nil
}

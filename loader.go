// DataLoader - a named handle to a resource in a Registry
package fakery

import (
	"log"
)

// DataLoader is a lightweight handle to one resource of a Registry. It is
// kept for compatibility with code written against the previous per-file
// loader API; it holds no data itself; all caching happens in the
// registry. The zero value, initialised with Init, uses the registry of
// the Fakery it is passed to (or the default registry).
type DataLoader struct {
	// Common file path for a specific type of data
	fileName string
	registry *Registry
}

// Loader returns a DataLoader bound to this registry for resource name.
func (r *Registry) Loader(name string) *DataLoader {
	return &DataLoader{fileName: name, registry: r}
}

// Initialize the data loader with a filePath. Every facet of data is
// associated with a unique {locale, filePath} tuple.
func (loader *DataLoader) Init(filePath string) {
	loader.fileName = filePath
}

// FileName returns the resource name of the loader.
func (loader *DataLoader) FileName() string {
	return loader.fileName
}

func (loader *DataLoader) registryOr(fallback *Registry) *Registry {
	if loader.registry != nil {
		return loader.registry
	}
	if fallback != nil {
		return fallback
	}
	return defaultRegistry
}

// SetIsMap registers the resource as map based data. The kind of a
// resource applies to all locales; the locale argument is kept for
// compatibility.
func (loader *DataLoader) SetIsMap(locale string) {
	spec := ResourceSpec{Name: loader.fileName, Kind: MapResource}
	if err := loader.registryOr(nil).Register(spec); err != nil {
		log.Printf("error - marking %s as map data for locale %s: %v\n", loader.fileName, locale, err)
	}
}

// Preload data for a given locale, returning a copy of its keys and values
func (loader *DataLoader) Preload(locale string) map[string]interface{} {
	localeData := loader.EnsureLoaded(locale)

	data := make(map[string]interface{}, len(localeData.data))
	for key, val := range localeData.data {
		data[key] = val
	}
	return data
}

// Return the data for the given locale, loading it if needed
func (loader *DataLoader) Get(locale string) *LocaleData {
	return loader.EnsureLoaded(locale)
}

// Lazy loader wrapper - lazily loads locale data for given locale
// once whenever requested. Once loaded, data remains in memory.
func (loader *DataLoader) EnsureLoaded(locale string) *LocaleData {
	return loader.registryOr(nil).Load(loader.fileName, locale)
}

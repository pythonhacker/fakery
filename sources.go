// Declaration of every built-in data source.
package fakery

import (
	"path"
	"strings"
)

// Built-in data source identifiers. Each is the file name relative to the
// locale directory (data/locales/<locale>/).
const (
	SourceAddress  = "address.json"
	SourceBeer     = "beer.json"
	SourceBook     = "book.json"
	SourceCar      = "car.json"
	SourceColor    = "color.json"
	SourceCurrency = "currency.json"
	SourceInternet = "internet.json"
	SourceJobs     = "jobs.json"
	SourceNames    = "names.json"
	SourceWine     = "wine.json"
	SourceWords    = "words.json"

	// SourceCarModelsPrefix is the family of per-maker car model files,
	// e.g. "cars/toyota.json".
	SourceCarModelsPrefix = "cars/"
)

// builtinSource is one entry of the source declaration table.
type builtinSource struct {
	name string
	opts []SourceOption
}

// builtinSources is the single declaration point for the bundled data.
// Adding a data file means adding one line here.
var builtinSources = []builtinSource{
	{name: SourceAddress},
	{name: SourceBeer},
	{name: SourceBook},
	{name: SourceCar},
	{name: SourceColor},
	{name: SourceCurrency, opts: []SourceOption{WithMapLocales(GenericLocale)}},
	{name: SourceInternet},
	{name: SourceJobs},
	{name: SourceNames},
	{name: SourceWine},
	{name: SourceWords},
}

// builtinFamilies declares the prefix families of the bundled data.
var builtinFamilies = []builtinSource{
	{name: SourceCarModelsPrefix, opts: []SourceOption{WithMapLocales(GenericLocale)}},
}

// RegisterBuiltinSources registers all bundled data sources in r. It lets
// a custom registry (e.g. one with a different data root) serve the same
// sources as the default registry.
func RegisterBuiltinSources(r *Registry) *Registry {
	for _, s := range builtinSources {
		r.Register(s.name, s.opts...)
	}
	for _, s := range builtinFamilies {
		r.RegisterPrefix(s.name, s.opts...)
	}
	return r
}

// defaultRegistry serves all generators. Being a package-level variable,
// it is ready before any init function or dependent variable runs.
var defaultRegistry = RegisterBuiltinSources(NewRegistry())

// DefaultRegistry returns the registry used by all Fakery generators.
func DefaultRegistry() *Registry {
	return defaultRegistry
}

// CarModelsSource returns the data source name holding the models of a
// car maker, e.g. "Mercedes-Benz" -> "cars/mercedesbenz.json".
func CarModelsSource(maker string) string {
	return path.Join(strings.TrimSuffix(SourceCarModelsPrefix, "/"), NormalizeString(maker)+".json")
}

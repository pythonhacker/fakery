// The catalogue of data resources bundled with fakery
package fakery

import (
	"fmt"
)

// Names of the bundled resources, relative to a locale directory.
const (
	ResourceAddress   = "address.json"
	ResourceBeer      = "beer.json"
	ResourceBook      = "book.json"
	ResourceCar       = "car.json"
	ResourceCarModels = "cars/*.json"
	ResourceColor     = "color.json"
	ResourceCurrency  = "currency.json"
	ResourceInternet  = "internet.json"
	ResourceJobs      = "jobs.json"
	ResourceNames     = "names.json"
	ResourceWine      = "wine.json"
	ResourceWords     = "words.json"
)

// BuiltinResources returns the specs of all bundled resources. Every
// Registry registers them on creation.
func BuiltinResources() []ResourceSpec {
	return []ResourceSpec{
		{Name: ResourceAddress, Kind: ListResource},
		{Name: ResourceBeer, Kind: ListResource},
		{Name: ResourceBook, Kind: ListResource},
		{Name: ResourceCar, Kind: ListResource},
		{Name: ResourceCarModels, Kind: MapResource},
		{Name: ResourceColor, Kind: ListResource},
		{Name: ResourceCurrency, Kind: MapResource},
		{Name: ResourceInternet, Kind: ListResource},
		{Name: ResourceJobs, Kind: ListResource},
		{Name: ResourceNames, Kind: ListResource},
		{Name: ResourceWine, Kind: ListResource},
		{Name: ResourceWords, Kind: ListResource},
	}
}

// carModelsResource returns the resource holding the models of a car maker
func carModelsResource(maker string) string {
	return fmt.Sprintf("cars/%s.json", NormalizeString(maker))
}

// Data of resource name for the Fakery's locale (with fallback)
func (f *Fakery) localeData(name string) *LocaleData {
	return f.Registry().Load(name, f.locale)
}

// Data of resource name for the generic locale
func (f *Fakery) genericData(name string) *LocaleData {
	return f.Registry().Load(name, GenericLocale)
}

// Generic data of resource name decoded into a T, decoded once and shared
func genericView[T any](f *Fakery, name string) *T {
	return viewOf[T](f.genericData(name))
}

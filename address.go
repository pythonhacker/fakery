// Functions related to a fake address
package gofakelib

import (
	"strings"
)

var (
	adata DataLoader
)

func init() {
	adata.Init("address.json")
}

type Address struct {
	Number     string `json:"unit,omitempty"`        // e.g: 112
	Building   string `json:"building,omitempty"`    // e.g: Tower B, Glennfidich Terraces
	Street     string `json:"street"`                // e.g: Beef Turkey Road
	Locality   string `json:"locality,omitempty"`    // Glennspring Premises
	City       string `json:"city"`                  // Edinburgh
	State      string `json:"state"`                 // Edinburgh
	PostalCode string `json:"postal_code,omitempty"` // Everywhere else on Earth
	ZipCode    string `json:"zip,omitempty"`         // The U.S only
	Country    string `json:"country"`               // Scotland
}

var cityFormats = WeightedArray{
	Items: []WeightedItem{
		{Item: "{{cityPrefix}} {{firstName}}{{citySuffix}}", Weight: 0.30},
		{Item: "{{cityPrefix}} {{firstName}}", Weight: 0.40},
		{Item: "{{firstName}}{{citySuffix}}", Weight: 0.20},
		{Item: "{{lastName}}{{citySuffix}}", Weight: 0.10}},
}

// Return a random fake city
func (f *Faker) City() string {
	localeData := adata.EnsureLoaded(GenericLocale)

	err, cityFormat := f.RandomItem(&cityFormats)
	if err != nil {
		return ""
	}
	// Get random prefix
	if strings.Contains(cityFormat, "{{cityPrefix}}") {
		cityFormat = strings.Replace(cityFormat, "{{cityPrefix}}", f.RandomString(localeData.Get("city_prefixes")), 1)
	}
	if strings.Contains(cityFormat, "{{firstName}}") {
		cityFormat = strings.Replace(cityFormat, "{{firstName}}", f.FirstName(), 1)
	} else if strings.Contains(cityFormat, "{{lastName}}") {
		cityFormat = strings.Replace(cityFormat, "{{lastName}}", f.LastName(), 1)
	}
	if strings.Contains(cityFormat, "{{citySuffix}}") {
		cityFormat = strings.Replace(cityFormat, "{{citySuffix}}", f.RandomString(localeData.Get("city_suffixes")), 1)
	}

	return cityFormat
}

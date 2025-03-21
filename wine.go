// Random wines! Cheers...\U0001f377
package gofakelib

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	wdata DataLoader
)

func init() {
	wdata.Init("wine.json")
}

type Wine struct {
	Name      string `json:"name"`      // Example: "Chateau Soleil Rouge"
	Varietal  string `json:"varietal"`  // Example: "Cabernet Sauvignon"
	Region    string `json:"region"`    // Example: "Napa Valley"
	Country   string `json:"country"`   // Example: "France"
	Vintage   string `json:"vintage"`   // Example: "2018"
	Alcohol   string `json:"alcohol"`   // Example: "13.5%"
	Body      string `json:"body"`      // Example: "Full-bodied"
	Acidity   string `json:"acidity"`   // Example: "Medium"
	Tannins   string `json:"tannins"`   // Example: "High"
	Sweetness string `json:"sweetness"` // Example: "Dry"
}

func (w Wine) String() string {
	val, _ := json.MarshalIndent(w, "", "\t")
	return string(val)
}

func (f *Faker) WineName() string {
	localeData := wdata.EnsureLoaded(GenericLocale)
	return f.RandomString(localeData.Get("wine_names"))
}

func (f *Faker) WineVarietal() string {
	localeData := wdata.EnsureLoaded(GenericLocale)
	return f.RandomString(localeData.Get("wine_varietals"))
}

func (f *Faker) WineRegion() string {
	localeData := wdata.EnsureLoaded(GenericLocale)
	return f.RandomString(localeData.Get("wine_regions"))
}

func (f *Faker) WineBody() string {
	localeData := wdata.EnsureLoaded(GenericLocale)
	return f.RandomString(localeData.Get("wine_bodies"))
}

func (f *Faker) WineAcidity() string {
	localeData := wdata.EnsureLoaded(GenericLocale)
	return f.RandomString(localeData.Get("wine_acidities"))
}

func (f *Faker) WineTannins() string {
	localeData := wdata.EnsureLoaded(GenericLocale)
	return f.RandomString(localeData.Get("wine_tannins"))
}

func (f *Faker) WineSweetness() string {
	localeData := wdata.EnsureLoaded(GenericLocale)
	return f.RandomString(localeData.Get("wine_sweetness"))
}

func (f *Faker) WineVintage() string {
	// This can be a year in the last 50 years, let us always go 1 year back
	year := time.Now().UTC().Year()
	return fmt.Sprintf("%d", f.RandIntBetween(year-50, year-1))
}

func (f *Faker) WineAlcohol() string {
	// in the 5 - 23% range
	return strconv.FormatFloat(f.RandFloat(2, 5.0, 24.0), 'f', 1, 64) + "%"
}

func (f *Faker) Wine() *Wine {
	var w Wine

	w.Name = f.WineName()
	w.Varietal = f.WineVarietal()
	region := f.WineRegion()
	// split this
	items := strings.SplitN(region, ",", 3)
	// first 2 for the region (area, city) and the last
	// the country
	w.Region = strings.Join(items[:2], ",")
	w.Country = items[2]
	w.Vintage = f.WineVintage()
	w.Alcohol = f.WineAlcohol()
	w.Body = f.WineBody()
	w.Acidity = f.WineAcidity()
	w.Tannins = f.WineTannins()
	w.Sweetness = f.WineSweetness()

	return &w
}

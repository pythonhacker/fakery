// Random wines! Cheers...\U0001f377
package fakery

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Wine structure and wine data - courtesy ChatGPT.
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
	Base
}

// Structure representing loaded wine data
type WineData struct {
	WineNames     []string `json:"wine_names"`
	WineVarietals []string `json:"wine_varietals"`
	WineRegions   []string `json:"wine_regions"`
	WineBodies    []string `json:"wine_bodies"`
	WineAcidities []string `json:"wine_acidities"`
	WineTannins   []string `json:"wine_tannins"`
	WineSweetness []string `json:"wine_sweetness"`
}

func (w Wine) String() string {
	return w.Base.String(w)
}

// Generic wine data, decoded once and shared
func (f *Fakery) wineData() *WineData {
	return genericView[WineData](f, ResourceWine)
}

func (f *Fakery) WineName() string {
	return f.RandomString(f.wineData().WineNames)
}

func (f *Fakery) WineVarietal() string {
	return f.RandomString(f.wineData().WineVarietals)
}

func (f *Fakery) WineRegion() string {
	return f.RandomString(f.wineData().WineRegions)
}

func (f *Fakery) WineBody() string {
	return f.RandomString(f.wineData().WineBodies)
}

func (f *Fakery) WineAcidity() string {
	return f.RandomString(f.wineData().WineAcidities)
}

func (f *Fakery) WineTannins() string {
	return f.RandomString(f.wineData().WineTannins)
}

func (f *Fakery) WineSweetness() string {
	return f.RandomString(f.wineData().WineSweetness)
}

func (f *Fakery) WineVintage() string {
	// This can be a year in the last 50 years, let us always go 1 year back
	year := time.Now().UTC().Year()
	return fmt.Sprintf("%d", f.RandIntBetween(year-50, year-1))
}

func (f *Fakery) WineAlcohol() string {
	// in the 5 - 23% range
	return strconv.FormatFloat(f.RandFloat(2, 5.0, 24.0), 'f', 1, 64) + "%"
}

func (f *Fakery) Wine() *Wine {
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

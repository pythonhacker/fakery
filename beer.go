// Random beers! Cheers...\U0001f37a
package gofakelib

import (
	"fmt"
	"strconv"
)

var (
	bdata DataLoader
)

func init() {
	bdata.Init("beer.json")
}

type Beer struct {
	Name    string `json:"name"`  // Crane Blowout IPA
	Style   string `json:"style"` // IPA
	Hops    string `json:"hops"`  // Wolfshade Hops
	Malt    string `json:"malt"`  // Belgian Malt
	Ibu     string `json:"ibu"`
	Blg     string `json:"blg"`
	Alcohol string `json:"alcohol"`
	Base
}

func (b Beer) String() string {
	return b.Base.String(b)
}

func (f *Faker) BeerName() string {
	localeData := bdata.EnsureLoaded(GenericLocale)
	return f.RandomString(localeData.Get("beer_names"))
}

func (f *Faker) BeerStyle() string {
	localeData := bdata.EnsureLoaded(GenericLocale)
	return f.RandomString(localeData.Get("beer_styles"))
}

func (f *Faker) BeerHops() string {
	localeData := bdata.EnsureLoaded(GenericLocale)
	return fmt.Sprintf("%s Hops", f.RandomString(localeData.Get("beer_hops")))
}

func (f *Faker) BeerMalt() string {
	localeData := bdata.EnsureLoaded(GenericLocale)
	return fmt.Sprintf("%s Malt", f.RandomString(localeData.Get("beer_malts")))
}

func (f *Faker) BeerAlcohol() string {
	return strconv.FormatFloat(f.RandFloat(2, 2.0, 10.0), 'f', 1, 64) + "%"
}

// Ibu will return a random beer ibu value between 10 and 100
func (f *Faker) BeerIbu() string {
	return strconv.Itoa(f.RandIntBetween(10, 100)) + " IBU"
}

// Blg will return a random beer blg between 5.0 and 20.0
func (f *Faker) BeerBlg() string {
	return strconv.FormatFloat(f.RandFloat(2, 5.0, 20.0), 'f', 1, 64) + "°Blg"
}

func (f *Faker) Beer() *Beer {
	var b Beer

	b.Name = f.BeerName()
	b.Style = f.BeerStyle()
	b.Hops = f.BeerHops()
	b.Malt = f.BeerMalt()
	b.Ibu = f.BeerIbu()
	b.Blg = f.BeerBlg()
	b.Alcohol = f.BeerAlcohol()

	return &b
}

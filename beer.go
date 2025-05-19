// Random beers! Cheers...\U0001f37a
package fakery

import (
	"fmt"
	"strconv"
)

var (
	beerLoader DataLoader
	beerData   BeerData
)

func init() {
	// Preload since this is all generic data
	beerLoader.Init("beer.json")
	dataMap := beerLoader.Preload(GenericLocale)
	// Convert to structure
	ConvertMapToStruct(dataMap, &beerData)
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

// Structure representing loaded beer data
type BeerData struct {
	BeerNames  []string `json:"beer_names"`
	BeerStyles []string `json:"beer_styles"`
	BeerHops   []string `json:"beer_hops"`
	BeerMalts  []string `json:"beer_malts"`
}

func (b Beer) String() string {
	return b.Base.String(b)
}

func (f *Faker) BeerName() string {
	return f.RandomString(beerData.BeerNames)
}

func (f *Faker) BeerStyle() string {
	return f.RandomString(beerData.BeerStyles)
}

func (f *Faker) BeerHops() string {
	return fmt.Sprintf("%s Hops", f.RandomString(beerData.BeerHops))
}

func (f *Faker) BeerMalt() string {
	return fmt.Sprintf("%s Malt", f.RandomString(beerData.BeerMalts))
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

// Functions related to a fake address
package gofakelib

import (
	"fmt"
	"strings"
)

var (
	adata DataLoader
)

func init() {
	adata.Init("address.json")
}

type Address struct {
	Number     string `json:"number"`                // A-1142
	Building   string `json:"building,omitempty"`    // Harvey Towers
	Street     string `json:"street"`                // Rue de Einstein
	City       string `json:"city"`                  // Edinburgh
	State      string `json:"state"`                 // Edinburgh
	PostalCode string `json:"postal_code,omitempty"` // Everywhere else on Earth
	ZipCode    string `json:"zip,omitempty"`         // Zipcode - The U.S only
	Country    string `json:"country"`               // Scotland

	FullAddress string `json:"full_address"` // Full address
	Base
}

func (a Address) String() string {
	return a.Base.String(a)
}

var cityFormats = WeightedArray{
	Items: []WeightedItem{
		{Item: "{{cityPrefix}} {{firstName}}{{citySuffix}}", Weight: 0.50},
		//		{Item: "{{cityPrefix}} {{firstName}}", Weight: 0.0},
		{Item: "{{firstName}}{{citySuffix}}", Weight: 0.40},
		{Item: "{{lastName}}{{citySuffix}}", Weight: 0.10}},
}

var streetAddressFormats = []string{
	"{{buildingNumber}} {{buildingName}} {{streetName}}",
}

var buildingNumberFormats = []string{"##", "@-##", "###", "##@", "#####", "###@", "@-###",
	"##-###", "## - Tower @", "### - Tower @"}
var buildingAddressFormats = []string{"Apt.", "Suite"}
var postCode = []string{"#####", "######", "#####-####"}
var zipCode = []string{"#####", "#####-####"}

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

// Return a random building number
func (f *Faker) BuildingNumber() string {

	var secFormat string

	format := f.RandomString(buildingNumberFormats)
	if f.Choice() == 1 {
		// Add a secondary address format to it
		secFormat = f.RandomString(buildingAddressFormats)
	}
	val := f.AlphifySpecific(f.Numerify(format), "G")

	if secFormat != "" {
		return fmt.Sprintf("%s %s", secFormat, val)
	}

	return fmt.Sprintf("#%s", val)
}

// Return random building name
func (f *Faker) BuildingName() string {

	var namePieces []string
	localeData := adata.EnsureLoaded(GenericLocale)

	if f.Choice() == 1 {
		namePieces = append(namePieces, f.FirstName())
	} else {
		namePieces = append(namePieces, f.LastName())
	}
	suffix := f.RandomString(localeData.Get("building_suffixes"))
	namePieces = append(namePieces, suffix)

	return strings.Join(namePieces, " ")
}

// Return random street name
func (f *Faker) StreetName() string {

	var name string
	localeData := adata.EnsureLoaded(GenericLocale)

	suffix := f.RandomString(localeData.Get("street_suffixes"))

	if f.Choice() == 1 {
		name = f.FirstName()
	} else {
		name = f.LastName()
	}
	// Add full name only if we get 6
	if f.RollDice() == 6 {
		name = f.Name()
	}

	// Only for "Rue"
	if suffix == "Rue" {
		return fmt.Sprintf("%s de %s", suffix, name)
	}

	return fmt.Sprintf("%s %s", name, suffix)
}

// Return a random street address
func (f *Faker) StreetAddress() string {

	streetAddress := f.RandomString(streetAddressFormats)

	if strings.Contains(streetAddress, "{{buildingNumber}}") {
		streetAddress = strings.Replace(streetAddress, "{{buildingNumber}}", f.BuildingNumber(), 1)
	}
	if strings.Contains(streetAddress, "{{buildingName}}") {
		streetAddress = strings.Replace(streetAddress, "{{buildingName}}", f.BuildingName(), 1)
	}

	streetAddress = strings.Replace(streetAddress, "{{streetName}}", f.StreetName(), 1)

	return streetAddress
}

// Random two letter state abbreviation
func (f *Faker) StateAbbr() string {
	localeData := adata.EnsureLoaded(f.locale)
	return f.RandomString(localeData.Get("state_abbrevs"))
}

// Random state
func (f *Faker) State() string {
	localeData := adata.EnsureLoaded(f.locale)
	states := localeData.Get("states")
	return f.RandomString(states)
}

func (f *Faker) PostCode() string {
	format := f.RandomString(postCode)
	return f.Numerify(format)
}

// For US
func (f *Faker) ZipCode() string {
	format := f.RandomString(zipCode)
	return f.Numerify(format)
}

func (f *Faker) Country() string {
	localeData := adata.EnsureLoaded(GenericLocale)
	return f.RandomString(localeData.Get("countries"))
}

func (f *Faker) Address() *Address {

	var a Address
	var code string

	streetAddress := f.RandomString(streetAddressFormats)

	a.Number = f.BuildingNumber()
	if strings.Contains(streetAddress, "{{buildingName}}") {
		a.Building = f.BuildingName()
	}
	a.Street = f.StreetName()
	a.City = f.City()
	a.State = f.State()
	if f.locale == "en_US" {
		a.ZipCode = f.ZipCode()
		code = a.ZipCode
	} else {
		a.PostalCode = f.PostCode()
		code = a.PostalCode
	}

	// get matching country of locale
	a.Country = getCountry(f.locale)
	if a.Building != "" {
		a.FullAddress = fmt.Sprintf("%s %s, %s, %s - %s, %s, %s", a.Number, a.Building, a.Street, a.City, code, a.State, a.Country)
	} else {
		a.FullAddress = fmt.Sprintf("%s, %s %s - %s, %s, %s", a.Number, a.Street, a.City, code, a.State, a.Country)
	}
	return &a
}

// Random state - not used
func (f *Faker) FakeState() string {
	localeData := adata.EnsureLoaded(f.locale)
	states := localeData.Get("states")

	// Remove any state <= 4 in length
	states = filterByLength(states, 5)

	// Load a state
	state1 := f.RandomString(states)
	// Remove this one

	state2 := f.RandomStringExcl(states, state1)
	// Split the first one at a vowel for example Kansas -> Kans
	// Split the second one at a vowel for example Hawaii -> Haw
	pieces1 := splitVowel(state1)
	pieces2 := splitVowel(state2)
	fmt.Println(state1, pieces1)
	fmt.Println(state2, pieces2)

	// convert string to lowercase

	chunks := []string{}

	if f.Choice() == 0 {
		if f.Choice() == 1 {
			chunks = append(chunks, strings.ToLower(pieces1[0]))
			chunks = append(chunks, strings.ToLower(pieces2[0]))
		} else {
			chunks = append(chunks, strings.ToLower(pieces1[0]))
			chunks = append(chunks, strings.ToLower(pieces2[1]))
		}

	} else {
		if f.Choice() == 0 {
			chunks = append(chunks, strings.ToLower(pieces1[1]))
			chunks = append(chunks, strings.ToLower(pieces2[0]))
		} else {
			chunks = append(chunks, strings.ToLower(pieces1[1]))
			chunks = append(chunks, strings.ToLower(pieces2[1]))
		}
	}

	return strings.Title(strings.Join(chunks, ""))

}

// given locale get the country
func getCountry(locale string) string {

	// Split the locale by underscore
	parts := strings.Split(locale, "_")

	// Validate format
	if len(parts) != 2 {
		return ""
	}

	localeData := adata.EnsureLoaded(GenericLocale)
	countryCodes := localeData.Get("country_codes")
	countries := localeData.Get("countries")

	// Extract the country code and convert to uppercase
	countryCode := strings.ToUpper(parts[1])

	// We are maintaining an index to index mapping
	// from country code -> countries arrays
	for idx, cCode := range countryCodes {
		if cCode == countryCode {
			return countries[idx]
		}
	}
	return ""
}

// Functions related to a fake person
package gofakelib

import (
	"encoding/json"
	"strings"
)

type Gender string

const (
	GenderMale   Gender = "Male"
	GenderFemale Gender = "Female"
)

var genders = []Gender{GenderMale, GenderFemale}

var (
	pdata DataLoader
)

func init() {
	pdata.Init(DefaultLocale, "names.json")
}

type Person struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Gender    string `json:"gender"`
	Prefix    string `json:"prefix,omitempty"`
	Suffix    string `json:"prefix,omitempty"`
	Username  string `json:"user_name"`
	Email     string `json:"email"`
}

func (p Person) String() string {
	val, _ := json.MarshalIndent(p, "", "\t")
	return string(val)
}

var nameFormats = WeightedArray{
	Items: []WeightedItem{
		// firstName lastName format - most common
		{Item: "{{firstName}} {{lastName}}", Weight: 0.70},
		{Item: "{{firstName}} {{lastName}} {{suffix}}", Weight: 0.10},
		{Item: "{{prefix}} {{firstName}} {{lastName}} {{suffix}}", Weight: 0.05},
		{Item: "{{prefix}} {{firstName}} {{lastName}}", Weight: 0.15},
	},
}

// Return a random gender
func (f Faker) Gender() Gender {
	idx := f.Choice()
	return genders[idx]
}

// Return a random name
func (f Faker) Name() string {

	var firstName string
	var lastName string

	localeData, _ := pdata.EnsureLoaded(f.locale)

	if f.Choice() == 0 {
		firstName = f.RandomString(localeData.Get("first_name_male"))
	} else {
		firstName = f.RandomString(localeData.Get("first_name_female"))
	}
	lastName = f.RandomString(localeData.Get("last_name"))

	return strings.Join([]string{firstName, lastName}, " ")
}

// returns a fake Person object
func (f Faker) Person() *Person {
	var person *Person

	gender := f.Gender()

	switch gender {
	case GenderMale:
		person = f.PersonMale()
	case GenderFemale:
		person = f.PersonFemale()
	}

	// Fill in rest
	person.Email = f.EmailWithName(person.FirstName, person.LastName)
	return person
}

// Returns a fake Person object with Male gender
func (f Faker) PersonMale() *Person {

	var person Person
	var nameFormat string

	err, nameFormat := f.RandomItem(&nameFormats)
	if err != nil {
		return nil
	}

	localeData, _ := pdata.EnsureLoaded(f.locale)

	person.FirstName = f.RandomString(localeData.Get("first_name_male"))
	person.LastName = f.RandomString(localeData.Get("last_name"))

	if strings.Contains(nameFormat, "{{prefix}}") {
		person.Prefix = f.RandomString(localeData.Get("prefix_male"))
	}

	if strings.Contains(nameFormat, "{{suffix}}") {
		person.Suffix = f.RandomString(localeData.Get("suffix_male"))
	}

	person.Gender = string(GenderMale)
	return &person
}

// Returns a fake Person object with Female gender
func (f Faker) PersonFemale() *Person {

	var person Person
	var nameFormat string

	err, nameFormat := f.RandomItem(&nameFormats)
	if err != nil {
		return nil
	}

	localeData, _ := pdata.EnsureLoaded(f.locale)

	person.FirstName = f.RandomString(localeData.Get("first_name_female"))
	person.LastName = f.RandomString(localeData.Get("last_name"))

	if strings.Contains(nameFormat, "{{prefix}}") {
		person.Prefix = f.RandomString(localeData.Get("prefix_female"))
	}

	if strings.Contains(nameFormat, "{{suffix}}") {
		person.Suffix = f.RandomString(localeData.Get("suffix_female"))
	}

	person.Gender = string(GenderFemale)
	return &person
}

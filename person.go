// Functions related to a fake person
package gofakelib

import (
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
	pdata.Init("names.json")
}

type Person struct {
	Name string `json:"name"`
	// full name with suffix prefix and all
	FullName  string `json:"full_name"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Gender    string `json:"gender"`
	Prefix    string `json:"prefix,omitempty"`
	Suffix    string `json:"suffix,omitempty"`
	Username  string `json:"user_name,omitempty"`
	Email     string `json:"email"`
	Job       string `json:"job"`
	Base
}

func (p Person) String() string {
	return p.Base.String(p)
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
func (f *Faker) Gender() Gender {
	idx := f.Choice()
	return genders[idx]
}

// Return a random name
func (f *Faker) Name() string {

	var firstName string
	var lastName string

	localeData := pdata.EnsureLoaded(f.locale)

	if f.Choice() == 0 {
		firstName = f.RandomString(localeData.Get("first_name_male"))
	} else {
		firstName = f.RandomString(localeData.Get("first_name_female"))
	}
	lastName = f.RandomString(localeData.Get("last_name"))

	return strings.Join([]string{firstName, lastName}, " ")
}

// Return random first name
func (f *Faker) FirstName() string {

	var firstName string

	localeData := pdata.EnsureLoaded(f.locale)

	if f.Choice() == 0 {
		firstName = f.RandomString(localeData.Get("first_name_male"))
	} else {
		firstName = f.RandomString(localeData.Get("first_name_female"))
	}

	return firstName
}

// Return random last name
func (f *Faker) LastName() string {

	var lastName string

	localeData := pdata.EnsureLoaded(f.locale)

	lastName = f.RandomString(localeData.Get("last_name"))
	return lastName
}

// returns a fake Person object
func (f *Faker) Person() *Person {
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
	person.Job = f.Job().Title
	return person
}

// Returns a fake Person object with Male gender
func (f *Faker) PersonMale() *Person {

	var person Person
	var nameFormat string
	var namePieces []string

	err, nameFormat := f.RandomItem(&nameFormats)
	if err != nil {
		return nil
	}

	localeData := pdata.EnsureLoaded(f.locale)

	if strings.Contains(nameFormat, "{{prefix}}") {
		person.Prefix = f.RandomString(localeData.Get("prefix_male"))
		namePieces = append(namePieces, person.Prefix)
	}

	person.FirstName = f.RandomString(localeData.Get("first_name_male"))
	person.LastName = f.RandomString(localeData.Get("last_name"))
	person.Name = strings.Join([]string{person.FirstName, person.LastName}, " ")

	namePieces = append(namePieces, person.FirstName)
	namePieces = append(namePieces, person.LastName)

	if strings.Contains(nameFormat, "{{suffix}}") {
		person.Suffix = f.RandomString(localeData.Get("suffix_male"))
		namePieces = append(namePieces, person.Suffix)
	}

	person.FullName = strings.Join(namePieces, " ")
	person.Gender = string(GenderMale)
	return &person
}

// Returns a fake Person object with Female gender
func (f *Faker) PersonFemale() *Person {

	var person Person
	var nameFormat string
	var namePieces []string

	err, nameFormat := f.RandomItem(&nameFormats)
	if err != nil {
		return nil
	}

	localeData := pdata.EnsureLoaded(f.locale)

	if strings.Contains(nameFormat, "{{prefix}}") {
		person.Prefix = f.RandomString(localeData.Get("prefix_female"))
		namePieces = append(namePieces, person.Prefix)
	}

	person.FirstName = f.RandomString(localeData.Get("first_name_female"))
	person.LastName = f.RandomString(localeData.Get("last_name"))
	person.Name = strings.Join([]string{person.FirstName, person.LastName}, " ")

	namePieces = append(namePieces, person.FirstName)
	namePieces = append(namePieces, person.LastName)

	if strings.Contains(nameFormat, "{{suffix}}") {
		person.Suffix = f.RandomString(localeData.Get("suffix_female"))
		namePieces = append(namePieces, person.Suffix)
	}

	person.FullName = strings.Join(namePieces, " ")
	person.Gender = string(GenderFemale)
	return &person
}

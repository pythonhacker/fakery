// Functions related to a fake person
package fakery

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
	personLoader DataLoader
	namesData    NamesData
)

func init() {
	personLoader.Init("names.json")
}

// Struct describing a person
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

// All data representing names
type NamesData struct {
	FirstNameMale   []string
	FirstNameFemale []string
	LastName        []string
	PrefixFemale    []string
	PrefixMale      []string
	SuffixMale      []string
	SuffixFemale    []string
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

	data := f.LoadLocale(&personLoader)

	if f.Choice() == 0 {
		firstName = f.RandomString(data.Get("first_name_male"))
	} else {
		firstName = f.RandomString(data.Get("first_name_female"))
	}
	lastName = f.RandomString(data.Get("last_name"))

	return strings.Join([]string{firstName, lastName}, " ")
}

// Return random first name
func (f *Faker) FirstName() string {

	var firstName string

	data := f.LoadLocale(&personLoader)

	if f.Choice() == 0 {
		firstName = f.RandomString(data.Get("first_name_male"))
	} else {
		firstName = f.RandomString(data.Get("first_name_female"))
	}

	return firstName
}

// Return random last name
func (f *Faker) LastName() string {

	var lastName string

	data := f.LoadLocale(&personLoader)

	lastName = f.RandomString(data.Get("last_name"))
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

	nameFormat, err := f.RandomWeightedItem(&nameFormats)
	if err != nil {
		return nil
	}

	data := f.LoadLocale(&personLoader)

	if strings.Contains(nameFormat, "{{prefix}}") {
		person.Prefix = f.RandomString(data.Get("prefix_male"))
		namePieces = append(namePieces, person.Prefix)
	}

	person.FirstName = f.RandomString(data.Get("first_name_male"))
	person.LastName = f.RandomString(data.Get("last_name"))
	person.Name = strings.Join([]string{person.FirstName, person.LastName}, " ")

	namePieces = append(namePieces, person.FirstName)
	namePieces = append(namePieces, person.LastName)

	if strings.Contains(nameFormat, "{{suffix}}") {
		person.Suffix = f.RandomString(data.Get("suffix_male"))
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

	nameFormat, err := f.RandomWeightedItem(&nameFormats)
	if err != nil {
		return nil
	}

	data := f.LoadLocale(&personLoader)

	if strings.Contains(nameFormat, "{{prefix}}") {
		person.Prefix = f.RandomString(data.Get("prefix_female"))
		namePieces = append(namePieces, person.Prefix)
	}

	person.FirstName = f.RandomString(data.Get("first_name_female"))
	person.LastName = f.RandomString(data.Get("last_name"))
	person.Name = strings.Join([]string{person.FirstName, person.LastName}, " ")

	namePieces = append(namePieces, person.FirstName)
	namePieces = append(namePieces, person.LastName)

	if strings.Contains(nameFormat, "{{suffix}}") {
		person.Suffix = f.RandomString(data.Get("suffix_female"))
		namePieces = append(namePieces, person.Suffix)
	}

	person.FullName = strings.Join(namePieces, " ")
	person.Gender = string(GenderFemale)
	return &person
}

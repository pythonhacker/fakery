// Functions related to a fake person
package fakelib

import (
	"fakelib/data"
	"strings"
)

type Gender string

const (
	GenderMale   Gender = "Male"
	GenderFemale Gender = "Female"
)

var genders = []Gender{GenderMale, GenderFemale}

type Person struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Gender    string `json:"gender"`
	Prefix    string `json:"prefix,omitempty"`
	Suffix    string `json:"prefix,omitempty"`
	Username  string `json:"user_name"`
	Email     string `json:"email"`
}

var nameFormats = data.WeightedArray{
	Items: []data.WeightedItem{
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

	if f.Choice() == 0 {
		firstName = f.RandomString(data.FirstNameMale)
	} else {
		firstName = f.RandomString(data.FirstNameFemale)
	}
	lastName = f.RandomString(data.LastName)

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

	person.FirstName = f.RandomString(data.FirstNameMale)
	person.LastName = f.RandomString(data.LastName)

	if strings.Contains(nameFormat, "{{prefix}}") {
		person.Prefix = f.RandomString(data.PrefixMale)
	}

	if strings.Contains(nameFormat, "{{suffix}}") {
		person.Suffix = f.RandomString(data.SuffixMale)
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

	person.FirstName = f.RandomString(data.FirstNameFemale)
	person.LastName = f.RandomString(data.LastName)

	if strings.Contains(nameFormat, "{{prefix}}") {
		person.Prefix = f.RandomString(data.PrefixFemale)
	}

	if strings.Contains(nameFormat, "{{suffix}}") {
		person.Suffix = f.RandomString(data.SuffixFemale)
	}

	person.Gender = string(GenderFemale)
	return &person
}

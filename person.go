package fakelib

import (
	"fakelib/data"
	"fmt"
	"strings"
	// "sync"
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

// returns a fake Person object
func (f Faker) Person() (*Person, error) {
	gender := f.Gender()

	switch gender {
	case GenderMale:
		return f.PersonMale()
	case GenderFemale:
		return f.PersonFemale()
	}

	return nil, fmt.Errorf("invalid gender - %s", gender)
}

// Returns a fake Person object with Male gender
func (f Faker) PersonMale() (*Person, error) {

	var person Person
	var nameFormat string

	err, nameFormat := f.RandomItem(nameFormats)
	if err != nil {
		return nil, err
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
	return &person, nil
}

// Returns a fake Person object with Female gender
func (f Faker) PersonFemale() (*Person, error) {

	var person Person
	var nameFormat string

	err, nameFormat := f.RandomItem(nameFormats)
	if err != nil {
		return nil, err
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
	return &person, nil
}

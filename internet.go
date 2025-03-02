// Functions related to a fake Internet
package gofakelib

import (
	"gofakelib/data"
	"strings"
)

func (f Faker) GetRandomTLD() string {
	_, tld := f.RandomItem(data.WeightedTLD)
	return tld
}

func (f Faker) GetRandomEmailDomain() string {
	return f.RandomString(data.FakeEmailDomains)
}

// return random email
func (f Faker) Email() string {
	var name string
	var pieces []string
	var prefix string
	var domain string

	name = f.Name()
	pieces = strings.Split(name, " ")
	domain = f.GetRandomEmailDomain()

	if f.Choice() == 0 {
		// first name first
		prefix = strings.Join([]string{pieces[0], pieces[1]}, ".")
	} else {
		// last name first
		prefix = strings.Join([]string{pieces[1], pieces[0]}, ".")
	}

	return strings.Join([]string{strings.ToLower(prefix), domain}, "@")
}

// Return random email but with given first name and last name
func (f Faker) EmailWithName(firstName, lastName string) string {
	var prefix string
	var domain string

	domain = f.GetRandomEmailDomain()

	if f.Choice() == 0 {
		// first name first
		prefix = strings.Join([]string{firstName, lastName}, ".")
	} else {
		// last name first
		prefix = strings.Join([]string{lastName, firstName}, ".")
	}

	return strings.Join([]string{strings.ToLower(prefix), domain}, "@")
}

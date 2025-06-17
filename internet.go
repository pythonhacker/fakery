// Functions related to a fake Internet
package fakery

import (
	"fmt"
	"strings"
)

var (
	netLoader DataLoader
)

func init() {
	netLoader.Init("internet.json")
}

// Internet data belongs to generic locale
// return random TLD
func (f *Fakery) TLD() string {

	tldArray, _ := f.LoadGenericLocale(&netLoader).GetWeightedArray("common_tlds_weighted", ":")
	tld, _ := f.RandomWeightedItem(tldArray)

	return tld
}

func (f *Fakery) EmailDomain() string {
	return f.RandomString(f.LoadGenericLocale(&netLoader).Get("fake_email_domains"))
}

func (f *Fakery) FreeEmailDomain() string {
	return f.RandomString(f.LoadGenericLocale(&netLoader).Get("free_email_domains"))
}

// return random email
func (f *Fakery) Email() string {
	var name string
	var pieces []string
	var prefix string
	var domain string

	name = f.Name()
	pieces = strings.Split(name, " ")

	domain = f.EmailDomain()

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
func (f *Fakery) EmailWithName(firstName, lastName string) string {
	var prefix string
	var domain string

	domain = f.EmailDomain()

	if f.Choice() == 0 {
		// first name first
		prefix = strings.Join([]string{firstName, lastName}, ".")
	} else {
		// last name first
		prefix = strings.Join([]string{lastName, firstName}, ".")
	}

	return strings.Join([]string{strings.ToLower(prefix), domain}, "@")
}

// Return a random username
func (f *Fakery) UserName() string {

	var separators = []string{".", "_", "-", ""}
	var adj, sep2 string

	firstName := strings.ToLower(f.FirstName())
	lastName := strings.ToLower(f.LastName())
	sep := f.RandomString(separators)

	choice := f.IntRange(12)

	switch choice {
	case 0:
		// firstname + seperator + lastname
		return firstName + sep + lastName
	case 1:
		// lastName + separator + firstName
		return lastName + sep + firstName
	case 2:
		// firstName + separator + random number
		return fmt.Sprintf("%s%s%d", firstName, sep, f.IntRange(1000))
	case 3:
		// firstName + separator + lastName + random number
		return fmt.Sprintf("%s%s%s%d", firstName, sep, lastName, f.IntRange(1000))
	case 4:
		// firstName + lastName + number
		return fmt.Sprintf("%s%s%d", firstName, lastName, f.IntRange(1000))
	case 5:
		// firstName + separator + adjective
		adj = f.AdjectivePositive()
		return fmt.Sprintf("%s%s%s", firstName, sep, adj)
	case 6:
		// lastName + separator + adjective
		adj = f.AdjectivePositive()
		return fmt.Sprintf("%s%s%s", lastName, sep, adj)
	case 7:
		// adjective + separator + firstName
		adj = f.AdjectivePositive()
		return fmt.Sprintf("%s%s%s", adj, sep, firstName)
	case 8:
		// adjective + separator + lastName
		adj = f.AdjectivePositive()
		return fmt.Sprintf("%s%s%s", adj, sep, lastName)
	case 9:
		adj = f.AdjectivePositive()
		// adjective + separator + lastName or firstName + number
		if f.Choice() == 1 {
			return fmt.Sprintf("%s%s%s%d", adj, sep, firstName, f.IntRange(1000))
		} else {
			return fmt.Sprintf("%s%s%s%d", adj, sep, lastName, f.IntRange(1000))
		}
	case 10:
		sep2 = f.RandomString(separators)
		adj = f.AdjectivePositive()
		if f.Choice() == 1 {
			// adjective + separator + firstName + separator + lastName
			return fmt.Sprintf("%s%s%s%s%s", adj, sep, firstName, sep2, lastName)
		} else {
			// firstName + separator + lastName + separator + adjective
			return fmt.Sprintf("%s%s%s%s%s", firstName, sep2, lastName, sep, adj)
		}
	case 11:
		// firstName + separator + lastName + separator+ adjective + number
		sep2 = f.RandomString(separators)
		adj = f.AdjectivePositive()
		return fmt.Sprintf("%s%s%s%s%s%d", firstName, sep2, lastName, sep, adj, f.IntRange(1000))
	}

	return ""
}

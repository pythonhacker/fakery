package data

import (
	"fmt"
)

var (
	personDataLoader JSONDataLoader
	FirstNameMale    []string
	FirstNameFemale  []string
	LastName         []string
	PrefixMale       []string
	PrefixFemale     []string
	SuffixMale       []string
	SuffixFemale     []string
)

func init() {
	// Load default locale data
	if err := personDataLoader.Load(defaultLocale, "names.json"); err != nil {
		fmt.Println("error loading names data - ", err)
	}
	FirstNameMale = personDataLoader.Get("first_name_male")
	FirstNameFemale = personDataLoader.Get("first_name_female")
	LastName = personDataLoader.Get("last_name")
	PrefixMale = personDataLoader.Get("prefix_male")
	PrefixFemale = personDataLoader.Get("prefix_female")
	SuffixMale = personDataLoader.Get("suffix_male")
	SuffixFemale = personDataLoader.Get("suffix_female")
}

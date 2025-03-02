package data

import (
	"log"
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
		log.Fatalf("error loading names data - %v", err)
	}
	// female and male names have been generated from the data
	// at https://www.ssa.gov/OACT/babynames/decades/names[decade]s.html
	// whee [decade]: range(1880, 2010)
	FirstNameMale = personDataLoader.Get("first_name_male")
	FirstNameFemale = personDataLoader.Get("first_name_female")
	// This has been generated from
	// https://babynames.com/blogs/names/1000-most-popular-last-names-in-the-u-s/
	LastName = personDataLoader.Get("last_name")
	PrefixMale = personDataLoader.Get("prefix_male")
	PrefixFemale = personDataLoader.Get("prefix_female")
	SuffixMale = personDataLoader.Get("suffix_male")
	SuffixFemale = personDataLoader.Get("suffix_female")
}

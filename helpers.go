// Data helpers module
package fakery

import (
	"encoding/json"
	"math"
	"regexp"
	"strings"
)

var MinInt = func(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}
var MaxInt = func(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

// structure which allows weighted data to allow
// probability based randomness
type WeightedItem struct {
	Item   string  `json:"item"`
	Weight float64 `json:"weight"`
}

// Structure holding array of weighted items
type WeightedArray struct {
	Items []WeightedItem
}

// validating a weighted array
func (w WeightedArray) Validate() (bool, float64) {
	// weights should add to 1.0
	var cumWeight float64 = 0.0

	for _, item := range w.Items {
		cumWeight += item.Weight
	}

	cumWeight = math.Round(cumWeight*100) / 100

	if cumWeight == 1.0 {
		return true, cumWeight
	}
	return false, cumWeight
}

// Filter an array of strings by length. Only strings
// with length >= minLength is returned in the return array
func FilterByLength(in []string, minLength int) []string {
	var filteredStrings []string

	for _, elem := range in {
		if len(elem) < minLength {
			continue
		}
		filteredStrings = append(filteredStrings, elem)
	}

	return filteredStrings

}

// Split a string at a vowel returning
// the two pieces, none smaller than 2 chars
func SplitVowel(in string) []string {

	vowelMap := map[string]bool{
		"a": true,
		"e": true,
		"i": true,
		"o": true,
		"u": true,
	}

	// replace any spaces
	if len(strings.Split(in, " ")) > 1 {
		// join em
		in = strings.Join(strings.Split(in, " "), "")
	}

	var pieces []string

	for idx, subs := range strings.Split(in, "") {
		if _, ok := vowelMap[subs]; ok && idx > 1 && idx < len(in) {
			subs1 := in[:idx]
			subs2 := in[idx:]
			pieces = append(pieces, subs1)
			pieces = append(pieces, subs2)
		}
	}

	return pieces
}

func startsWithVowel(in string) bool {
	return strings.Contains("AEIOU", string(in[0]))
}

// Normalize a string
// remove punctuations, remove spaces and lowercase
func NormalizeString(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Use a regex to remove all non-alphanumeric characters
	reg := regexp.MustCompile(`[^a-z0-9]`)
	s = reg.ReplaceAllString(s, "")

	return s
}

// determine article depending on following word
func DetermineArticle(following string, f *Fakery) string {
	var article = "a"
	if startsWithVowel(following) {
		if f.Choice() == 1 {
			article = "an"
		} else {
			article = "the"
		}
	}

	return article
}

func ConvertMapToStruct(data map[string]interface{}, result interface{}) error {
	// Marshal map into JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Unmarshal JSON into the struct
	err = json.Unmarshal(jsonData, result)
	if err != nil {
		return err
	}

	return nil
}

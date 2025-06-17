// The main fakery module. We also absorb most utility functions in this module.
package fakery

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const upperAlpha = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const lowerAlpha = "abcdefghijklmnopqrstuvwxy"
const hexChars = "0123456789ABCDEF"

// Base is the base class for data types, not Fakery
type Base struct{}

func (b Base) String(v interface{}) string {
	var sb strings.Builder

	encoder := json.NewEncoder(&sb)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(v)
	if err == nil {
		return sb.String()
	}
	return ""
}

// We will use the format of having a few top level
// structures and associating field data to those
// structures. E.g: Person->Name, Person->Address,
// Person->Creditcard etc.
type Fakery struct {
	rng    *rand.Rand
	locale string
	// Cached locale data
	data *LocaleData
}

// Wrapper function to load locale data
// at any given time a faker instance is associated with
// only one state mapping to its current request
// so keep this state locally is safe.
func (f *Fakery) LoadLocale(l *DataLoader) *LocaleData {
	return l.EnsureLoaded(f.locale)
}

// Wrapper function to load generic locale data
func (f *Fakery) LoadGenericLocale(l *DataLoader) *LocaleData {
	return l.EnsureLoaded(GenericLocale)
}

// Return an integer in the interval [0, n)
func (f *Fakery) IntRange(n int) int {
	return f.rng.Intn(n)
}

// Return a random digit in range 0..9
func (f *Fakery) RandDigit() int {
	return f.IntRange(9999) % 10
}

// Random integer between x and y, y>x
func (f *Fakery) RandIntBetween(x, y int) int {
	if y < x {
		return 0
	}
	return x + f.IntRange(y-x)
}

// Return fake random float in range of (x, y)
func (f *Fakery) RandFloat(maxDecimals, x, y int) float64 {
	value := float64(f.RandIntBetween(x, y-1))
	if maxDecimals < 1 {
		return value
	}

	p := int(math.Pow10(maxDecimals))
	decimals := float64(f.RandIntBetween(0, p)) / float64(p)

	return value + decimals
}

// Return a random digit in range 1..9
func (f *Fakery) RandDigitNonZero() int {
	return f.IntRange(9) + 1
}

// Return either 0 or 1 for a two choice array index
func (f *Fakery) Choice() int {
	randVal := f.rng.Float64()
	if randVal <= 0.5 {
		return 0
	}
	return 1
}

// Roll dice i.e return number in range [1-6]
func (f *Fakery) RollDice() int {
	return f.IntRange(6) + 1
}

// Random integer of given length
func (f *Fakery) RandInteger(length int) int {

	if length == 1 {
		return f.RandDigit()
	}

	var min int = int(math.Pow10(length - 1))
	var max int = int(math.Pow10(length)) - 1

	return f.RandIntBetween(min, max)
}

// Return a random item according to weights
func (f *Fakery) RandomWeightedItem(array *WeightedArray) (string, error) {
	randVal := f.rng.Float64()

	if ok, val := array.Validate(); !ok {
		return "", fmt.Errorf("weighted array validation failed, weight: %.2f", val)
	}

	// Track cumulative weight
	var cumulativeWeight float64 = 0.0

	// Find the item whose cumulative weight range contains randVal
	for _, item := range array.Items {
		cumulativeWeight += item.Weight
		if randVal < cumulativeWeight {
			return item.Item, nil
		}
	}

	// Fallback in case of rounding errors
	return array.Items[len(array.Items)-1].Item, nil
}

// Return one of any two strings in a string array
// Use only when the array has two items
func (f *Fakery) OneOf(choices []string) string {
	return choices[f.Choice()]
}

// Return a random string
func (f *Fakery) RandomString(stringItems []string) string {

	idx := f.IntRange(len(stringItems))
	return stringItems[idx]
}

// Return a random string excluding given one
func (f *Fakery) RandomStringExcl(stringItems []string, excl string) string {

	var idxElem int

	for i := 0; i < len(stringItems); i++ {
		if stringItems[i] == excl {
			idxElem = i
			break
		}
	}

	stringItems = append(stringItems[:idxElem], stringItems[idxElem+1:]...)
	idx := f.IntRange(len(stringItems))
	return stringItems[idx]
}

// Return a random letter [A-Z]
func (f *Fakery) RandomAZ() string {

	return fmt.Sprintf("%c", upperAlpha[f.IntRange(26)])
}

// Return a random hex char as string
func (f *Fakery) RandomHex() string {
	return fmt.Sprintf("%c", hexChars[f.IntRange(len(hexChars))])
}

// Return a random letter [A-Z] in a specific range
func (f *Fakery) RandomAZSpecific(upto string) string {

	idx := strings.Index(upperAlpha, upto)
	if idx == -1 {
		return ""
	}
	return fmt.Sprintf("%c", upperAlpha[f.IntRange(idx+1)])
}

// Return a string which replaces all '#' chars in a string
// with numbers
func (f *Fakery) Numerify(inputString string) string {
	var digit int

	count := strings.Count(inputString, "#")
	for i := 1; i <= count; i++ {
		if i == 1 {
			// we dont want the string to begin with 0
			digit = f.RandDigitNonZero()
		} else {
			digit = f.RandDigit()
		}
		inputString = strings.Replace(inputString, "#", fmt.Sprintf("%d", digit), 1)
	}

	return inputString
}

// Return a string which replaces all '@' chars in a string
// with alphabets
func (f *Fakery) Alphify(inputString string) string {

	count := strings.Count(inputString, "@")
	for i := 1; i <= count; i++ {
		inputString = strings.Replace(inputString, "@", f.RandomAZ(), 1)
	}

	return inputString
}

// Return a string which replaces all '@' chars in a string
// with alphabets till a specific string
func (f *Fakery) AlphifySpecific(inputString, upto string) string {

	count := strings.Count(inputString, "@")
	for i := 1; i <= count; i++ {
		inputString = strings.Replace(inputString, "@", f.RandomAZSpecific(upto), 1)
	}

	return inputString
}

// Return a string which replaces all '@' chars in a string
// with alphabets upto a specific range
func (f *Fakery) SpecificAlphify(inputString string) string {

	count := strings.Count(inputString, "@")
	for i := 1; i <= count; i++ {
		inputString = strings.Replace(inputString, "@", f.RandomAZ(), 1)
	}

	return inputString
}

// Capitalize all words in a  sentence
func (f *Fakery) Capitalize(sentence string) string {

	return cases.Title(language.Und).String(sentence)
}

// Set locale
func (f *Fakery) SetLocale(locale string) {
	f.locale = locale
}

// Fakery constructors
func New() *Fakery {
	seed := time.Now().Nanosecond()
	return &Fakery{
		rng:    rand.New(rand.NewSource(int64(seed))),
		locale: DefaultLocale,
	}
}

func NewFromLocale(locale string) *Fakery {
	seed := time.Now().Nanosecond()
	return &Fakery{
		rng:    rand.New(rand.NewSource(int64(seed))),
		locale: locale,
	}
}

func NewFromSeed(seed int64) *Fakery {
	return &Fakery{
		rng:    rand.New(rand.NewSource(int64(seed))),
		locale: DefaultLocale,
	}
}

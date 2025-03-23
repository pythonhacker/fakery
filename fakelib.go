package gofakelib

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"
)

const upperAlpha = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const lowerAlpha = "abcdefghijklmnopqrstuvwxy"

// Base is the base class for data types, not Faker
type Base struct{}

func (b Base) String(v interface{}) string {
	val, _ := json.MarshalIndent(v, "", "\t")
	return string(val)
}

// We will use the format of having a few top level
// structures and associating field data to those
// structures. E.g: Person->Name, Person->Address,
// Person->Creditcard etc.
type Faker struct {
	rng    *rand.Rand
	locale string
	// Cached locale data
	data *LocaleData
}

// Wrapper function to load locale data
// at any given time a faker instance is associated with
// only one state mapping to its current request
// so keep this state locally is safe.
func (f *Faker) LoadLocale(l *DataLoader) *LocaleData {
	return l.EnsureLoaded(f.locale)
}

// Wrapper function to load generic locale data
func (f *Faker) LoadGenericLocale(l *DataLoader) *LocaleData {
	return l.EnsureLoaded(GenericLocale)
}

// Return an integer in the interval [0, n)
func (f *Faker) IntRange(n int) int {
	return f.rng.Intn(n)
}

// Return a random digit in range 0..9
func (f *Faker) RandDigit() int {
	return f.IntRange(9999) % 10
}

// Random integer between x and y, y>x
func (f *Faker) RandIntBetween(x, y int) int {
	if y < x {
		return 0
	}
	return x + f.IntRange(y-x)
}

// Return fake random float in range of (x, y)
func (f *Faker) RandFloat(maxDecimals, x, y int) float64 {
	value := float64(f.RandIntBetween(x, y-1))
	if maxDecimals < 1 {
		return value
	}

	p := int(math.Pow10(maxDecimals))
	decimals := float64(f.RandIntBetween(0, p)) / float64(p)

	return value + decimals
}

// Return a random digit in range 1..9
func (f *Faker) RandDigitNonZero() int {
	return f.IntRange(9) + 1
}

// Return either 0 or 1 for a two choice array index
func (f *Faker) Choice() int {
	randVal := f.rng.Float64()
	if randVal <= 0.5 {
		return 0
	}
	return 1
}

// Roll dice i.e return number in range [1-6]
func (f *Faker) RollDice() int {
	return f.IntRange(6) + 1
}

// Return a random item according to weights
func (f *Faker) RandomItem(array *WeightedArray) (string, error) {
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
func (f *Faker) OneOf(choices []string) string {
	return choices[f.Choice()]
}

// Return a random string
func (f *Faker) RandomString(stringItems []string) string {

	idx := f.IntRange(len(stringItems))
	return stringItems[idx]
}

// Return a random string excluding given one
func (f *Faker) RandomStringExcl(stringItems []string, excl string) string {

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
func (f *Faker) RandomAZ() string {

	return fmt.Sprintf("%c", upperAlpha[f.IntRange(26)])
}

// Return a random letter [A-Z] in a specific range
func (f *Faker) RandomAZSpecific(upto string) string {

	idx := strings.Index(upperAlpha, upto)
	if idx == -1 {
		return ""
	}
	return fmt.Sprintf("%c", upperAlpha[f.IntRange(idx+1)])
}

// Return a string which replaces all '#' chars in a string
// with numbers
func (f *Faker) Numerify(inputString string) string {
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
func (f *Faker) Alphify(inputString string) string {

	count := strings.Count(inputString, "@")
	for i := 1; i <= count; i++ {
		inputString = strings.Replace(inputString, "@", f.RandomAZ(), 1)
	}

	return inputString
}

// Return a string which replaces all '@' chars in a string
// with alphabets till a specific string
func (f *Faker) AlphifySpecific(inputString, upto string) string {

	count := strings.Count(inputString, "@")
	for i := 1; i <= count; i++ {
		inputString = strings.Replace(inputString, "@", f.RandomAZSpecific(upto), 1)
	}

	return inputString
}

// Return a string which replaces all '@' chars in a string
// with alphabets upto a specific range
func (f *Faker) SpecificAlphify(inputString string) string {

	count := strings.Count(inputString, "@")
	for i := 1; i <= count; i++ {
		inputString = strings.Replace(inputString, "@", f.RandomAZ(), 1)
	}

	return inputString
}

// Set locale
func (f *Faker) SetLocale(locale string) {
	f.locale = locale
}

// Faker constructors
func New() *Faker {
	seed := time.Now().Nanosecond()
	return &Faker{
		rng:    rand.New(rand.NewSource(int64(seed))),
		locale: DefaultLocale,
	}
}

func NewFromLocale(locale string) *Faker {
	seed := time.Now().Nanosecond()
	return &Faker{
		rng:    rand.New(rand.NewSource(int64(seed))),
		locale: locale,
	}
}

func NewFromSeed(seed int64) *Faker {
	return &Faker{
		rng:    rand.New(rand.NewSource(int64(seed))),
		locale: DefaultLocale,
	}
}

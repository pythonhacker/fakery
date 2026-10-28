// The main fakery module. We also absorb most utility functions in this module.
package fakery

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
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
	mu       sync.RWMutex
	rngMu    sync.Mutex
	rng      *rand.Rand
	locale   string
	registry *LoaderRegistry
}

func (f *Fakery) loaderRegistry() *LoaderRegistry {
	f.mu.RLock()
	reg := f.registry
	f.mu.RUnlock()
	if reg != nil {
		return reg
	}
	return defaultRegistry
}

// Locale returns the locale currently configured on this Fakery instance.
func (f *Fakery) Locale() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.locale
}

// Registry returns the LoaderRegistry bound to this Fakery instance.
func (f *Fakery) Registry() *LoaderRegistry {
	return f.loaderRegistry()
}

// SetRegistry binds a custom LoaderRegistry to this Fakery instance (nil restores DefaultRegistry).
func (f *Fakery) SetRegistry(reg *LoaderRegistry) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.registry = reg
}

// Wrapper function to load locale data from an explicit DataLoader.
func (f *Fakery) LoadLocale(l *DataLoader) *LocaleData {
	if l == nil {
		return nil
	}
	return l.ResolveLocale(f.Locale())
}

// Wrapper function to load generic locale data from an explicit DataLoader.
func (f *Fakery) LoadGenericLocale(l *DataLoader) *LocaleData {
	if l == nil {
		return nil
	}
	return l.EnsureLoaded(GenericLocale)
}

// LoadResource loads locale-specific data (with fallback) from the central LoaderRegistry.
func (f *Fakery) LoadResource(resource string) *LocaleData {
	return f.loaderRegistry().LoadLocale(resource, f.Locale())
}

// LoadGenericResource loads generic locale data from the central LoaderRegistry.
func (f *Fakery) LoadGenericResource(resource string) *LocaleData {
	return f.loaderRegistry().LoadGeneric(resource)
}

// Return an integer in the interval [0, n)
func (f *Fakery) IntRange(n int) int {
	if n <= 0 {
		return 0
	}
	f.rngMu.Lock()
	defer f.rngMu.Unlock()
	return f.rng.Intn(n)
}

// Return a random digit in range 0..9
func (f *Fakery) RandDigit() int {
	return f.IntRange(9999) % 10
}

// Random integer between x and y, y>x
func (f *Fakery) RandIntBetween(x, y int) int {
	if y <= x {
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
	f.rngMu.Lock()
	randVal := f.rng.Float64()
	f.rngMu.Unlock()
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
	if array == nil || len(array.Items) == 0 {
		return "", fmt.Errorf("weighted array is empty")
	}

	f.rngMu.Lock()
	randVal := f.rng.Float64()
	f.rngMu.Unlock()

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
	if len(choices) == 0 {
		return ""
	}
	if len(choices) == 1 {
		return choices[0]
	}
	return choices[f.Choice()]
}

// Return a random string
func (f *Fakery) RandomString(stringItems []string) string {
	if len(stringItems) == 0 {
		return ""
	}
	idx := f.IntRange(len(stringItems))
	return stringItems[idx]
}

// Return a random string excluding given one
func (f *Fakery) RandomStringExcl(stringItems []string, excl string) string {
	filtered := make([]string, 0, len(stringItems))
	for _, item := range stringItems {
		if item != excl {
			filtered = append(filtered, item)
		}
	}
	return f.RandomString(filtered)
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

// Capitalize all words in a sentence
func (f *Fakery) Capitalize(sentence string) string {
	return cases.Title(language.Und).String(sentence)
}

// Set locale
func (f *Fakery) SetLocale(locale string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.locale = locale
}

// Fakery constructors
func New() *Fakery {
	seed := time.Now().UnixNano()
	return &Fakery{
		rng:    rand.New(rand.NewSource(seed)),
		locale: DefaultLocale,
	}
}

func NewFromLocale(locale string) *Fakery {
	seed := time.Now().UnixNano()
	return &Fakery{
		rng:    rand.New(rand.NewSource(seed)),
		locale: locale,
	}
}

func NewFromSeed(seed int64) *Fakery {
	return &Fakery{
		rng:    rand.New(rand.NewSource(seed)),
		locale: DefaultLocale,
	}
}

// NewWithRegistry creates a Fakery instance configured with a specific locale and LoaderRegistry.
func NewWithRegistry(locale string, reg *LoaderRegistry) *Fakery {
	seed := time.Now().UnixNano()
	return &Fakery{
		rng:      rand.New(rand.NewSource(seed)),
		locale:   locale,
		registry: reg,
	}
}

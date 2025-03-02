package fakelib

import (
	"errors"
	"math/rand"
	"time"
)

// We will use the format of having a few top level
// structures and associating field data to those
// structures. E.g: Person->Name, Person->Address,
// Person->Creditcard etc.
type Faker struct {
	rng *rand.Rand
}

type WeightedItem struct {
	Item   string
	Weight float64
}

type WeightedArray struct {
	Items []WeightedItem
}

func (w WeightedArray) validate() bool {
	// weights should add to 1.0
	var cumWeight float64 = 0.0

	for _, item := range w.Items {
		cumWeight += item.Weight
	}

	if cumWeight == 1.0 {
		return true
	}
	return false
}

// Return an integer in the interval [0, n)
func (f Faker) IntRange(n int) int {
	return f.rng.Intn(n)
}

// Return either 0 or 1 for a two choice array index
func (f Faker) Choice() int {
	randVal := f.rng.Float64()
	if randVal <= 0.5 {
		return 0
	}
	return 1
}

// Return a random item according to weights
func (f Faker) RandomItem(array WeightedArray) (error, string) {
	randVal := f.rng.Float64()

	if val := array.validate(); !val {
		return errors.New("weighted array validation failed"), ""
	}

	// Track cumulative weight
	var cumulativeWeight float64 = 0.0

	// Find the item whose cumulative weight range contains randVal
	for _, item := range array.Items {
		cumulativeWeight += item.Weight
		if randVal < cumulativeWeight {
			return nil, item.Item
		}
	}

	// Fallback in case of rounding errors
	return nil, array.Items[len(array.Items)-1].Item
}

// Return a random string
func (f Faker) RandomString(stringItems []string) string {

	idx := f.IntRange(len(stringItems))
	return stringItems[idx]
}

func New() *Faker {
	seed := time.Now().Nanosecond()
	return &Faker{
		rng: rand.New(rand.NewSource(int64(seed))),
	}
}

package fakelib

import (
	"fakelib/data"
	"fmt"
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
func (f Faker) RandomItem(array *data.WeightedArray) (error, string) {
	randVal := f.rng.Float64()

	if ok, val := array.Validate(); !ok {
		return fmt.Errorf("weighted array validation failed, weight: %.2f", val), ""
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

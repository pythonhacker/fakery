package tests

import (
	"fmt"
	"sync"
	"fakery"
	"testing"
)

// Run these tests with: go test -race ./...

const raceWorkers = 64

func runTogether(workers int, fn func(int)) {
	start := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()
			<-start
			fn(i)
		}(i)
	}

	close(start)
	wg.Wait()
}

// Concurrent first access to one locale.
//
// DataLoader.Get reads and writes localeDataMap without synchronization.
// Multiple goroutines can all observe that the locale is absent and
// concurrently insert a LocaleData.
func TestDataLoaderGetConcurrentSameLocale(t *testing.T) {
	var loader fakery.DataLoader
	loader.Init("names.json")

	runTogether(raceWorkers, func(_ int) {
		loader.Get("en")
	})
}

// Concurrent first access to different locales.
//
// This exercises concurrent writes to the same map even when each
// goroutine requests a different locale.
func TestDataLoaderGetConcurrentDifferentLocales(t *testing.T) {
	var loader fakery.DataLoader
	loader.Init("names.json")

	runTogether(raceWorkers, func(i int) {
		loader.Get(fmt.Sprintf("locale-%d", i))
	})
}

// SetIsMap writes configIsMap while Get reads it to initialize LocaleData.
//
// The two operations are not synchronized in the current implementation.
func TestDataLoaderSetIsMapConcurrentWithGet(t *testing.T) {
	var loader fakery.DataLoader
	loader.Init("currency.json")

	runTogether(raceWorkers, func(i int) {
		if i%2 == 0 {
			loader.SetIsMap("en")
		} else {
			loader.Get("en")
		}
	})
}

// EnsureLoaded checks LocaleData.loaded outside sync.Once.
//
// Pre-create the LocaleData so this test focuses on concurrent access to
// loaded rather than concurrent map initialization.
func TestDataLoaderEnsureLoadedConcurrent(t *testing.T) {
	var loader fakery.DataLoader
	loader.Init("race-test-file-that-does-not-exist.json")

	localeData := loader.Get("en")
	if localeData == nil {
		t.Fatal("Get returned nil")
	}

	runTogether(raceWorkers, func(_ int) {
		loader.EnsureLoaded("en")
	})
}

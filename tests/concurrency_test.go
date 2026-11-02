package tests

import (
	"fakery"
	"fmt"
	"sync"
	"testing"
)

// Run fn from n goroutines released at the same time
func runConcurrently(n int, fn func(i int)) {
	var start, done sync.WaitGroup
	start.Add(1)
	done.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer done.Done()
			start.Wait()
			fn(i)
		}(i)
	}
	start.Done()
	done.Wait()
}

func TestConcurrentLoadSameResource(t *testing.T) {
	r, cfs := newTestRegistry()

	const workers = 64
	results := make([]*fakery.LocaleData, workers)
	runConcurrently(workers, func(i int) {
		results[i] = r.Load(fakery.ResourceNames, "en_US")
	})

	for _, data := range results {
		Expect(t, true, data == results[0])
		Expect(t, nil, data.Err())
	}
	Expect(t, 1, cfs.count("en_US/names.json"))
}

func TestConcurrentLoadDifferentLocales(t *testing.T) {
	r, cfs := newTestRegistry()

	locales := []string{"en_US", "en_GB", "fr_FR", "de_DE", fakery.GenericLocale}
	resources := []string{
		fakery.ResourceNames, fakery.ResourceAddress, fakery.ResourceJobs,
		fakery.ResourceCurrency, "malformed.json", "missing.json",
	}

	const rounds = 16
	type result struct {
		key  string
		data *fakery.LocaleData
	}
	results := make([]result, rounds*len(locales)*len(resources))
	runConcurrently(len(results), func(i int) {
		locale := locales[i%len(locales)]
		name := resources[(i/len(locales))%len(resources)]
		results[i] = result{key: name + "@" + locale, data: r.Load(name, locale)}
	})

	// Every (resource, locale) pair resolved to a single shared value
	seen := make(map[string]*fakery.LocaleData)
	for _, res := range results {
		if prev, ok := seen[res.key]; ok {
			Expect(t, true, prev == res.data, res.key)
		}
		seen[res.key] = res.data
	}

	// Each file was read at most once, despite fallbacks racing each other
	for _, locale := range locales {
		for _, name := range resources {
			p := locale + "/" + name
			Expect(t, true, cfs.count(p) <= 1, p)
		}
	}

	// Locale specific data stayed isolated
	Expect(t, "Doe", seen[fakery.ResourceNames+"@en_US"].Get("last_name")[0])
	Expect(t, "Smith", seen[fakery.ResourceNames+"@en_GB"].Get("last_name")[0])
	Expect(t, "en_US", seen[fakery.ResourceNames+"@fr_FR"].Locale())
	Expect(t, fakery.GenericLocale, seen[fakery.ResourceJobs+"@de_DE"].Locale())
}

func TestConcurrentGeneratorsPerInstance(t *testing.T) {
	locales := []string{"en_US", "en_GB"}

	runConcurrently(32, func(i int) {
		f := fakery.NewFromLocale(locales[i%len(locales)])
		for j := 0; j < 10; j++ {
			if f.Person() == nil || f.Address() == nil || f.Car() == nil ||
				f.Currency() == nil || f.Book() == nil || f.Beer() == nil ||
				f.Wine() == nil || f.Color() == nil || f.CreditCard() == nil {
				panic(fmt.Sprintf("worker %d: nil value", i))
			}
			if len(f.TLD()) == 0 || len(f.UserName()) == 0 || len(f.ColorName()) == 0 {
				panic(fmt.Sprintf("worker %d: empty value", i))
			}
		}
	})
}

func TestConcurrentGeneratorsSharedInstance(t *testing.T) {
	f := fakery.New()

	var mu sync.Mutex
	var empty []string
	runConcurrently(32, func(i int) {
		for j := 0; j < 10; j++ {
			values := map[string]string{
				"name":     f.Name(),
				"city":     f.City(),
				"state":    f.State(),
				"car":      f.CarModel(),
				"currency": f.CurrencyCode(),
				"book":     f.BookTitle(),
				"beer":     f.BeerName(),
				"wine":     f.WineName(),
				"job":      f.Job().Title,
				"adverb":   f.Adverb(),
			}
			for k, v := range values {
				if v == "" {
					mu.Lock()
					empty = append(empty, k)
					mu.Unlock()
				}
			}
		}
	})
	Expect(t, 0, len(empty), empty)
}

func TestConcurrentRegisterAndLoad(t *testing.T) {
	r, _ := newTestRegistry()

	runConcurrently(32, func(i int) {
		name := fmt.Sprintf("dyn%d.json", i%4)
		_ = r.Register(fakery.ResourceSpec{Name: name, Kind: fakery.ListResource})
		r.Load(name, fakery.GenericLocale)
		r.Spec(name)
		r.Resources()
		r.Loaded(name, fakery.GenericLocale)
	})
}

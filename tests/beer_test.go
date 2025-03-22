package tests

import (
	"gofakelib"
	"testing"
)

func TestBeerName(t *testing.T) {
	Expect(t, true, len(gofakelib.New().BeerName()) > 0)
}

func TestBeerStyle(t *testing.T) {
	Expect(t, true, len(gofakelib.New().BeerStyle()) > 0)
}

func TestBeerHops(t *testing.T) {
	Expect(t, true, len(gofakelib.New().BeerHops()) > 0)
}

func TestBeerMalt(t *testing.T) {
	Expect(t, true, len(gofakelib.New().BeerMalt()) > 0)
}

func TestBeerAlcohol(t *testing.T) {
	Expect(t, true, len(gofakelib.New().BeerAlcohol()) > 0)
}

func TestBeerIbu(t *testing.T) {
	Expect(t, true, len(gofakelib.New().BeerIbu()) > 0)
}

func TestBeerBlg(t *testing.T) {
	Expect(t, true, len(gofakelib.New().BeerBlg()) > 0)
}

func TestBeer(t *testing.T) {
	b := gofakelib.New().Beer()
	Expect(t, true, b != nil)
	// Test fields which are never empty
	Expect(t, true, len(b.Name) > 0)
	Expect(t, true, len(b.Style) > 0)
	Expect(t, true, len(b.Hops) > 0)
	Expect(t, true, len(b.Malt) > 0)
	Expect(t, true, len(b.Alcohol) > 0)
	Expect(t, true, len(b.Ibu) > 0)
	Expect(t, true, len(b.Blg) > 0)
}

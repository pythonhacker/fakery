package tests

import (
	"gofakelib"
	"testing"
)

func TestCurrencyName(t *testing.T) {
	Expect(t, true, len(gofakelib.New().CurrencyName()) > 0)
}

func TestCurrencyCode(t *testing.T) {
	Expect(t, true, len(gofakelib.New().CurrencyCode()) > 0)
}

func TestCurrencyCountry(t *testing.T) {
	Expect(t, true, len(gofakelib.New().CurrencyCountry()) > 0)
}

func TestCurrency(t *testing.T) {
	c := gofakelib.New().Currency()
	Expect(t, true, c != nil)
	// Test fields which are never empty
	Expect(t, true, len(c.Name) > 0)
	Expect(t, true, len(c.Code) > 0)
	Expect(t, true, len(c.Country) > 0)
	Expect(t, true, c.Amount > 0)
}

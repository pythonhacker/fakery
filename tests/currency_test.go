package tests

import (
	"fakery"
	"testing"
)

func TestCurrencyName(t *testing.T) {
	Expect(t, true, len(fakery.New().CurrencyName()) > 0)
}

func TestCurrencyCode(t *testing.T) {
	Expect(t, true, len(fakery.New().CurrencyCode()) > 0)
}

func TestCurrencyCountry(t *testing.T) {
	Expect(t, true, len(fakery.New().CurrencyCountry()) > 0)
}

func TestCurrency(t *testing.T) {
	c := fakery.New().Currency()
	Expect(t, true, c != nil)
	// Test fields which are never empty
	Expect(t, true, len(c.Name) > 0)
	Expect(t, true, len(c.Code) > 0)
	Expect(t, true, len(c.Country) > 0)
	Expect(t, true, c.Amount > 0)
}

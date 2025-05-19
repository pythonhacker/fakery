package tests

import (
	"fakery"
	"testing"
)

func TestCity(t *testing.T) {
	Expect(t, true, len(fakery.New().City()) > 0)
}

func TestBuildingNumber(t *testing.T) {
	Expect(t, true, len(fakery.New().BuildingNumber()) > 0)
}

func TestBuildingName(t *testing.T) {
	Expect(t, true, len(fakery.New().BuildingName()) > 0)
}

func TestStreetName(t *testing.T) {
	Expect(t, true, len(fakery.New().StreetName()) > 0)
}

func TestStreetAddress(t *testing.T) {
	Expect(t, true, len(fakery.New().StreetAddress()) > 0)
}

func TestState(t *testing.T) {
	Expect(t, true, len(fakery.New().State()) > 0)
}

func TestStateAbbr(t *testing.T) {
	Expect(t, true, len(fakery.New().StateAbbr()) > 0)
}

func TestPostCode(t *testing.T) {
	Expect(t, true, len(fakery.New().PostCode()) > 0)
}

func TestZipCode(t *testing.T) {
	Expect(t, true, len(fakery.New().ZipCode()) > 0)
}

func TestCountry(t *testing.T) {
	Expect(t, true, len(fakery.New().Country()) > 0)
}

func TestAddress(t *testing.T) {
	a := fakery.New().Address()
	Expect(t, true, a != nil)
	Expect(t, true, len(a.Number) > 0)
	Expect(t, true, len(a.Street) > 0)
	Expect(t, true, len(a.City) > 0)
	Expect(t, true, len(a.ZipCode) > 0)
	Expect(t, true, len(a.Country) > 0)
	Expect(t, true, len(a.FullAddress) > 0)
}

package tests

import (
	"gofakelib"
	"testing"
)

func TestCity(t *testing.T) {
	Expect(t, true, len(gofakelib.New().City()) > 0)
}

func TestBuildingNumber(t *testing.T) {
	Expect(t, true, len(gofakelib.New().BuildingNumber()) > 0)
}

func TestBuildingName(t *testing.T) {
	Expect(t, true, len(gofakelib.New().BuildingName()) > 0)
}

func TestStreetName(t *testing.T) {
	Expect(t, true, len(gofakelib.New().StreetName()) > 0)
}

func TestStreetAddress(t *testing.T) {
	Expect(t, true, len(gofakelib.New().StreetAddress()) > 0)
}

func TestState(t *testing.T) {
	Expect(t, true, len(gofakelib.New().State()) > 0)
}

func TestStateAbbr(t *testing.T) {
	Expect(t, true, len(gofakelib.New().StateAbbr()) > 0)
}

func TestPostCode(t *testing.T) {
	Expect(t, true, len(gofakelib.New().PostCode()) > 0)
}

func TestZipCode(t *testing.T) {
	Expect(t, true, len(gofakelib.New().ZipCode()) > 0)
}

func TestCountry(t *testing.T) {
	Expect(t, true, len(gofakelib.New().Country()) > 0)
}

func TestAddress(t *testing.T) {
	a := gofakelib.New().Address()
	Expect(t, true, a != nil)
	Expect(t, true, len(a.Number) > 0)
	Expect(t, true, len(a.Street) > 0)
	Expect(t, true, len(a.City) > 0)
	Expect(t, true, len(a.ZipCode) > 0)
	Expect(t, true, len(a.Country) > 0)
	Expect(t, true, len(a.FullAddress) > 0)
}

package tests

import (
	"fakery"
	"testing"
)

func TestCarMake(t *testing.T) {
	Expect(t, true, len(fakery.New().CarMake()) > 0)
}

func TestCarModel(t *testing.T) {
	Expect(t, true, len(fakery.New().CarModel()) > 0)
}

func TestCarCategory(t *testing.T) {
	Expect(t, true, len(fakery.New().CarCategory()) > 0)
}

func TestCarSeries(t *testing.T) {
	Expect(t, true, len(fakery.New().CarSeries()) > 0)
}

func TestCarTransmission(t *testing.T) {
	Expect(t, true, len(fakery.New().CarTransmission()) > 0)
}

func TestCarPlate(t *testing.T) {
	Expect(t, true, len(fakery.New().CarPlate()) > 0)
}

func TestCar(t *testing.T) {
	c := fakery.New().Car()
	Expect(t, true, c != nil)
	// Test fields which are never empty
	Expect(t, true, len(c.Make) > 0)
	Expect(t, true, len(c.Model) > 0)
	Expect(t, true, len(c.Category) > 0)
	Expect(t, true, len(c.Series) > 0)
	Expect(t, true, len(c.Type) > 0)
	Expect(t, true, len(c.Transmission) > 0)
	Expect(t, true, len(c.Plate) > 0)
	Expect(t, true, c.Year >= 1990)

}

package tests

import (
	"fakery"
	"testing"
)

func TestWineName(t *testing.T) {
	Expect(t, true, len(fakery.New().WineName()) > 0)
}

func TestWineVarietal(t *testing.T) {
	Expect(t, true, len(fakery.New().WineVarietal()) > 0)
}

func TestWineRegion(t *testing.T) {
	Expect(t, true, len(fakery.New().WineRegion()) > 0)
}

func TestWineBody(t *testing.T) {
	Expect(t, true, len(fakery.New().WineBody()) > 0)
}

func TestWineAcidity(t *testing.T) {
	Expect(t, true, len(fakery.New().WineAcidity()) > 0)
}

func TestWineTannins(t *testing.T) {
	Expect(t, true, len(fakery.New().WineTannins()) > 0)
}

func TestWineSweetness(t *testing.T) {
	Expect(t, true, len(fakery.New().WineSweetness()) > 0)
}

func TestWineAlcohol(t *testing.T) {
	Expect(t, true, len(fakery.New().WineAlcohol()) > 0)
}

func TestWine(t *testing.T) {
	w := fakery.New().Wine()
	Expect(t, true, w != nil)
	// Test fields which are never empty
	Expect(t, true, len(w.Name) > 0)
	Expect(t, true, len(w.Varietal) > 0)
	Expect(t, true, len(w.Region) > 0)
	Expect(t, true, len(w.Country) > 0)
	Expect(t, true, len(w.Region) > 0)
	Expect(t, true, len(w.Alcohol) > 0)
	Expect(t, true, len(w.Acidity) > 0)
	Expect(t, true, len(w.Tannins) > 0)
	Expect(t, true, len(w.Sweetness) > 0)
}

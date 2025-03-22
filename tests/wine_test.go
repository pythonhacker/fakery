package tests

import (
	"gofakelib"
	"testing"
)

func TestWineName(t *testing.T) {
	Expect(t, true, len(gofakelib.New().WineName()) > 0)
}

func TestWineVarietal(t *testing.T) {
	Expect(t, true, len(gofakelib.New().WineVarietal()) > 0)
}

func TestWineRegion(t *testing.T) {
	Expect(t, true, len(gofakelib.New().WineRegion()) > 0)
}

func TestWineBody(t *testing.T) {
	Expect(t, true, len(gofakelib.New().WineBody()) > 0)
}

func TestWineAcidity(t *testing.T) {
	Expect(t, true, len(gofakelib.New().WineAcidity()) > 0)
}

func TestWineTannins(t *testing.T) {
	Expect(t, true, len(gofakelib.New().WineTannins()) > 0)
}

func TestWineSweetness(t *testing.T) {
	Expect(t, true, len(gofakelib.New().WineSweetness()) > 0)
}

func TestWineAlcohol(t *testing.T) {
	Expect(t, true, len(gofakelib.New().WineAlcohol()) > 0)
}

func TestWine(t *testing.T) {
	w := gofakelib.New().Wine()
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

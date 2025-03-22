package tests

import (
	"gofakelib"
	"testing"
)

func TestBlood(t *testing.T) {
	f := gofakelib.New()
	b := f.Blood()
	Expect(t, true, b != nil)
	Expect(t, true, len(b.String()) > 0)
}

package tests

import (
	"fakery"
	"testing"
)

func TestBlood(t *testing.T) {
	f := fakery.New()
	b := f.Blood()
	Expect(t, true, b != nil)
	Expect(t, true, len(b.String()) > 0)
}

package tests

import (
	"fakery"
	"testing"
)

func TestBinaryString(t *testing.T) {
	f := fakery.New()
	length := f.RandDigitNonZero()
	Expect(t, true, len(f.BinaryString(length)) == length)
}

package tests

import (
	"gofakelib"
	"testing"
)

func TestBinaryString(t *testing.T) {
	f := gofakelib.New()
	length := f.RandDigitNonZero()
	Expect(t, true, len(f.BinaryString(length)) == length)
}

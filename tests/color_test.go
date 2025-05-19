package tests

import (
	"fmt"
	"fakery"
	"strings"
	"testing"
)

func TestColorName(t *testing.T) {
	colorName := fakery.New().ColorName()
	pieces := strings.Split(colorName, " ")
	Expect(t, true, len(colorName) > 0)
	Expect(t, true, len(pieces) > 1)
}

func TestSafeColorName(t *testing.T) {
	colorName := fakery.New().SafeColorName()
	pieces := strings.Split(colorName, " ")
	Expect(t, true, len(colorName) > 0)
	Expect(t, true, len(pieces) == 1)
}

func TestHexColor(t *testing.T) {
	hexColor := fakery.New().HexColor()
	Expect(t, true, len(hexColor) > 0)
	// Scan into int
	var hexInt int
	fmt.Sscanf(hexColor, "#%06X", &hexInt)
	Expect(t, true, hexInt < 0xFFFFFF)
}

func TestRGBColor(t *testing.T) {
	rgbColor := fakery.New().RGBColor()
	Expect(t, true, len(rgbColor) > 0)
	Expect(t, true, len(strings.Split(rgbColor, ",")) == 3)
}

func TestHSLColor(t *testing.T) {
	hslColor := fakery.New().HSLColor()
	Expect(t, true, len(hslColor) > 0)
	Expect(t, true, len(strings.Split(hslColor, ",")) == 3)
}

func TestColor(t *testing.T) {
	color := fakery.New().Color()
	Expect(t, true, color != nil)
	Expect(t, true, len(color.Hex) > 0)
	Expect(t, true, len(color.HSL) > 0)
	Expect(t, true, len(color.RGB) > 0)
}

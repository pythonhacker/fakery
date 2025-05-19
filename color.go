// Fake colors
package fakery

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

var (
	colorLoader DataLoader
	colorData   ColorData
)

func init() {
	colorLoader.Init("color.json")
	ConvertMapToStruct(colorLoader.Preload(GenericLocale), &colorData)
}

type Color struct {
	Hex  string `json:"hex"`                    // hex color
	RGB  string `json:"rgb"`                    // matching rgb color
	HSL  string `json:"hsl"`                    // matching hsl color
	Name string `json:"closest_name,omitempty"` // closest matching color name
	Base
}

// Named colors with their RGB values
var namedColors = map[string][3]int{
	"Red":       {255, 0, 0},
	"Green":     {0, 255, 0},
	"Blue":      {0, 0, 255},
	"Yellow":    {255, 255, 0},
	"Orange":    {255, 165, 0},
	"Purple":    {128, 0, 128},
	"Pink":      {255, 192, 203},
	"Brown":     {165, 42, 42},
	"Cyan":      {0, 255, 255},
	"Magenta":   {255, 0, 255},
	"Gray":      {128, 128, 128},
	"Black":     {0, 0, 0},
	"White":     {255, 255, 255},
	"Maroon":    {128, 0, 0},
	"Olive":     {128, 128, 0},
	"Navy":      {0, 0, 128},
	"Teal":      {0, 128, 128},
	"Silver":    {192, 192, 192},
	"Gold":      {255, 215, 0},
	"Beige":     {245, 245, 220},
	"Turquoise": {64, 224, 208},
	"Salmon":    {250, 128, 114},
	"Crimson":   {220, 20, 60},
	"Indigo":    {75, 0, 130},
	"Plum":      {221, 160, 221},
	"Lavender":  {230, 230, 250},
	"Azure":     {240, 255, 255},
}

func (c Color) String() string {
	return c.Base.String(c)
}

type ColorData struct {
	ColorAdjectives []string `json:"color_adjectives"`
	ColorNames      []string `json:"color_names"`
}

// Return a random color name
func (f *Faker) ColorName() string {
	colorAdjective := f.RandomString(colorData.ColorAdjectives)
	colorName := f.RandomString(colorData.ColorNames)

	return strings.Join([]string{colorAdjective, colorName}, " ")
}

// Return a more realistic "safe" color name
func (f *Faker) SafeColorName() string {
	return f.RandomString(colorData.ColorNames)
}

// Random hex color
func (f *Faker) HexColor() string {
	return fmt.Sprintf("#%06X", f.IntRange(0xFFFFFF))
}

// Random RGB Color
func (f *Faker) RGBColor() string {
	var colors []string

	for _, item := range f.rgb() {
		colors = append(colors, strconv.Itoa(item))
	}

	return strings.Join(colors, ",")
}

// Return Random color in HSL
func (f *Faker) HSLColor() string {

	var hsl []string

	for _, item := range f.hsl() {
		hsl = append(hsl, strconv.Itoa(item))
	}
	return strings.Join(hsl, ",")
}

// Return a random color
func (f *Faker) Color() *Color {

	var c Color
	var hslColor []string
	var rgbColor []string

	// First make a random HSL color
	hsl := f.hsl()

	for _, item := range hsl {
		hslColor = append(hslColor, strconv.Itoa(item))
	}

	c.HSL = strings.Join(hslColor, ",")
	// Now convert this to RGB and Hex
	rgb := f.hslToRGB(hsl)

	for _, item := range rgb {
		rgbColor = append(rgbColor, strconv.Itoa(item))
	}

	c.RGB = strings.Join(rgbColor, ",")
	c.Hex = f.rgbToHex(rgb)
	// Try to find closest match in name
	c.Name = f.closestColor(rgb)

	return &c

}

// rgb as 3 tuple int
func (f *Faker) rgb() []int {
	r := f.IntRange(256)
	g := f.IntRange(256)
	b := f.IntRange(256)

	return []int{r, g, b}
}

// hsl as 3 tuple int
func (f *Faker) hsl() []int {
	// Hue is in range 0-359,
	// Sat is in range 0-100
	// Light is in range 0-100
	var h, s, l int

	h = f.IntRange(360)
	s = f.IntRange(101)
	l = f.IntRange(101)

	return []int{h, s, l}
}

// Convert Hex color to RGB int array
func (f *Faker) hexToRGB(hex string) []int {
	var r, g, b int

	fmt.Sscanf(hex, "#%02X%02X%02X", &r, &g, &b) // Parse hex values
	return []int{r, g, b}
}

// HSL to RGB conversion
func (f *Faker) hslToRGB(hsl []int) []int {
	var h, s, l int

	h = hsl[0]
	s = hsl[1]
	l = hsl[2]

	H := float64(h) / 360.0
	S := float64(s) / 100.0
	L := float64(l) / 100.0

	var r, g, b float64

	if S == 0 {
		r, g, b = L, L, L // Achromatic (gray)
	} else {
		var hueToRgb = func(p, q, t float64) float64 {
			if t < 0 {
				t += 1
			}
			if t > 1 {
				t -= 1
			}
			if t < 1.0/6.0 {
				return p + (q-p)*6*t
			}
			if t < 1.0/2.0 {
				return q
			}
			if t < 2.0/3.0 {
				return p + (q-p)*(2.0/3.0-t)*6
			}
			return p
		}

		q := L * (1 + S)
		if L >= 0.5 {
			q = L + S - L*S
		}
		p := 2*L - q

		r = hueToRgb(p, q, H+1.0/3.0)
		g = hueToRgb(p, q, H)
		b = hueToRgb(p, q, H-1.0/3.0)
	}

	return []int{int(r * 255), int(g * 255), int(b * 255)}
}

// Convert RGB to HEX
func (f *Faker) rgbToHex(rgb []int) string {
	return fmt.Sprintf("#%02X%02X%02X", rgb[0], rgb[1], rgb[2])
}

// Calculate Euclidean distance between two RGB colors
func (f *Faker) colorDistance(r1, g1, b1, r2, g2, b2 int) float64 {
	return math.Sqrt(math.Pow(float64(r1-r2), 2) + math.Pow(float64(g1-g2), 2) + math.Pow(float64(b1-b2), 2))
}

// Find the closest named color
func (f *Faker) closestColor(rgb []int) string {

	var closestName string

	minDist := math.MaxFloat64
	for name, namedRgb := range namedColors {
		dist := f.colorDistance(rgb[0], rgb[1], rgb[2], namedRgb[0], namedRgb[1], namedRgb[2])
		if dist < minDist {
			minDist = dist
			closestName = name
		}
	}

	shade := f.getShade(f.calculateLuminance(rgb[0], rgb[1], rgb[2]))
	// Dont add shades for white and black
	if closestName != "White" && closestName != "Black" {
		return fmt.Sprintf("%s %s", shade, closestName)
	}

	return closestName
}

// Calculate brightness (luminance)
func (f *Faker) calculateLuminance(r, g, b int) float64 {
	return 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
}

// Determine shade category
func (f *Faker) getShade(luminance float64) string {
	if luminance < 85 {
		return "Dark"
	} else if luminance < 170 {
		return "Medium"
	} else {
		return "Light"
	}
}

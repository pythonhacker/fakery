// Build and deliver fake cars.
package fakery

import (
	"fmt"
	"time"
)

var (
	carLoader DataLoader
	carData   CarData
)

func init() {
	carLoader.Init("car.json")
	ConvertMapToStruct(carLoader.Preload(GenericLocale), &carData)
}

type Car struct {
	Make         string `json:"make"`            // Company - Toyota, Tata, Suzuki, Subaru
	Model        string `json:"model"`           // Glanza, Harrier, Swift
	Category     string `json:"category"`        // Hatchback, SUV, Coupe
	Series       string `json:"series"`          // E/S/G/V, XE/XT/XA, VXI/VZI etc
	Type         string `json:"type"`            // Fuel type - petrol, diesel, electric, hybrid, alien
	Transmission string `json:"transmission"`    // Manual, Automatic, CVT, DCT etc
	Year         int    `json:"year"`            // Manufacturing year
	Plate        string `json:"plate,omitempty"` // Registration plate
	Base
}

func (c Car) String() string {
	return c.Base.String(c)
}

type CarData struct {
	CarMakers             []string `json:"car_makers"`
	CarCategories         []string `json:"car_categories"`
	CarSeries             []string `json:"car_series"`
	CarCategoryAdjectives []string `json:"car_category_adjectives"`
	CarFuelTypes          []string `json:"car_fuel_types"`
}

var transmissionTypes = WeightedArray{
	Items: []WeightedItem{
		{Item: "Automatic", Weight: 0.35},
		{Item: "Manual", Weight: 0.22},
		{Item: "CVT", Weight: 0.15},
		{Item: "DCT", Weight: 0.12},
		{Item: "AMT", Weight: 0.10},
		{Item: "IMT", Weight: 0.05},
		{Item: "Triptronic", Weight: 0.01},
	},
}

var carPlateFormats = []string{
	"@@ ####",
	"@@-##-@@-####",
	"@@## @@@",
	"@@@ ####",
	"@@-###-@@",
	"@@-####",
	"####-@@-@@",
	"@@ ## ####",
	"@@ ## @@@",
	"@-###-@@",
	"@@@-##-####",
	"@@-@@-####",
	"@@-##-@@-####",
	"@@@-###-@@",
	"###-@@-####",
	"@@-###-@@@",
	"@@-####-@@",
	"@ ### @@",
	"@@@ ## ####",
	"@@-@@-##-####",
}

// The car company/make
func (f *Faker) CarMake() string {
	return f.RandomString(carData.CarMakers)
}

func (f *Faker) CarSeries() string {
	return f.RandomString(carData.CarSeries)
}

func (f *Faker) makeFromModel(make string) string {
	var makerLoader DataLoader

	makerNorm := NormalizeString(make)
	makerLoader.Init(fmt.Sprintf("cars/%s.json", makerNorm))
	makerLoader.SetIsMap(GenericLocale)

	// Load maker's data
	data := f.LoadGenericLocale(&makerLoader)
	model := data.RandomWeightedItem(f)
	return model["model"]
}

// A random car model
func (f *Faker) CarModel() string {

	make := f.CarMake()
	return f.makeFromModel(make)
}

func (f *Faker) CarMakeAndModel() (string, string) {

	make := f.CarMake()
	return make, f.makeFromModel(make)
}

func (f *Faker) CarCategory() string {
	category := f.RandomString(carData.CarCategories)
	// 5/6 times add an adjective
	if f.RollDice() < 6 {
		adjective := f.RandomString(carData.CarCategoryAdjectives)
		if adjective != category {
			return fmt.Sprintf("%s %s", adjective, category)
		}
	}

	return category
}

func (f *Faker) CarType() string {
	return f.RandomString(carData.CarFuelTypes)
}

func (f *Faker) CarTransmission() string {
	transmission, _ := f.RandomWeightedItem(&transmissionTypes)
	return transmission
}

func (f *Faker) CarPlate() string {
	format := f.RandomString(carPlateFormats)
	return f.Alphify(f.Numerify(format))
}

// Random car
func (f *Faker) Car() *Car {
	var car Car

	make, model := f.CarMakeAndModel()
	car.Make = make
	car.Model = model
	car.Category = f.CarCategory()
	car.Series = f.CarSeries()
	car.Type = f.CarType()
	car.Transmission = f.CarTransmission()
	car.Year = f.RandIntBetween(1990, time.Now().Year())
	car.Plate = f.CarPlate()

	return &car
}

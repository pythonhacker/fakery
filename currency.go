package fakery

var (
	currencyLoader DataLoader
)

func init() {
	currencyLoader.Init("currency.json")
	// Indicate the data is an array of maps for generic locale
	currencyLoader.SetIsMap(GenericLocale)
}

type Currency struct {
	Name    string  `json:"name"`    // Rupee
	Code    string  `json:"code"`    // INR
	Country string  `json:"country"` // India
	Amount  float64 `json:"amount"`  // 986.50
	Base
}

func (c Currency) String() string {
	return c.Base.String(c)
}

func (f *Fakery) CurrencyCode() string {
	item := f.LoadGenericLocale(&currencyLoader).RandomWeightedItem(f)
	return item["code"]
}

func (f *Fakery) CurrencyName() string {
	item := f.LoadGenericLocale(&currencyLoader).RandomWeightedItem(f)
	return item["currency"]
}

func (f *Fakery) CurrencyCountry() string {
	item := f.LoadGenericLocale(&currencyLoader).RandomWeightedItem(f)
	return item["country"]
}

func (f *Fakery) Currency() *Currency {
	var c Currency

	// The data has to be consistent
	item := f.LoadGenericLocale(&currencyLoader).RandomWeightedItem(f)
	c.Name = item["currency"]
	c.Code = item["code"]
	c.Country = item["country"]

	// Amounts are in range 1 -> 1000
	amount := float64(f.RandIntBetween(1, 1000))
	fractional := 0.0

	// Pick a decimal in the range of 0.00 -> 0.95
	for i := 0; i < 100; i += 5 {
		if f.RollDice() == 6 {
			fractional = float64(i) / 100.0
			break
		}
	}

	amount += fractional
	c.Amount = amount

	return &c
}

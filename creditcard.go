// Generate fake credit card data
package fakery

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type CreditCard struct {
	Number     string `json:"number"`
	Type       string `json:"type"`
	CVV        string `json:"cvv"`
	ExpiryDate string `json:"expiry_date"`
	Name       string `json:"name"`
	Base
}

func (c CreditCard) String() string {
	return c.Base.String(c)
}

func (c *CreditCard) Validate() bool {
	return luhnCheck(c.Number)
}

var creditCardTypes = WeightedArray{
	Items: []WeightedItem{
		{Item: "VISA", Weight: 0.4},
		{Item: "MasterCard", Weight: 0.4},
		{Item: "AMEX", Weight: 0.1},
		{Item: "Discover", Weight: 0.1},
	},
}

// LuhnCheck computes if a number passes the Luhn algorithm
func luhnCheck(card string) bool {
	sum := 0
	double := false

	for i := len(card) - 1; i >= 0; i-- {
		digit := int(card[i] - '0')
		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		double = !double
	}
	return sum%10 == 0
}

// generateValid completes a valid card number from prefix and total length
func generateValid(prefix string, length int) string {
	for {
		num := prefix
		for i := 0; i < length-len(prefix)-1; i++ {
			num += strconv.Itoa(rand.Intn(10))
		}
		for d := 0; d <= 9; d++ {
			testNum := num + strconv.Itoa(d)
			if luhnCheck(testNum) {
				return testNum
			}
		}
	}
}

// generate returns a random but unchecked credit card number
func generate(prefix string, length int) string {
	num := prefix
	for i := 0; i < length-len(prefix); i++ {
		num += strconv.Itoa(rand.Intn(10))
	}

	//	fmt.Println(luhnCheck(num))
	return num
}

func (f *Fakery) CreditCardCompany() string {
	company, _ := f.RandomWeightedItem(&creditCardTypes)
	return company
}

func (f *Fakery) CreditCardType() string {
	return f.CreditCardCompany()
}

// Generates a card number for the given type
func (f *Fakery) CreditCardNumber(cardType string) string {

	switch strings.ToLower(cardType) {
	case "visa":
		return generate("4", 16)
	case "mastercard":
		// Use 51-55 or 2221-2720
		prefix := ""
		if rand.Intn(2) == 0 {
			prefix = strconv.Itoa(51 + rand.Intn(5)) // 51-55
		} else {
			prefix = strconv.Itoa(2221 + rand.Intn(500)) // 2221-2720
		}
		return generate(prefix, 16)
	case "amex":
		amexPrefixes := []string{"34", "37"}
		return generate(amexPrefixes[rand.Intn(2)], 15)
	case "discover":
		discoverPrefixes := []string{"6011", "65", "644", "645", "646", "647", "648", "649"}
		return generate(discoverPrefixes[rand.Intn(len(discoverPrefixes))], 16)
	}

	return "unsupported card type"
}

func (f *Fakery) CreditCardExpiryDate() string {

	month := f.RandIntBetween(1, 13)
	currentYear := time.Now().Year()
	year := f.RandIntBetween(currentYear+1, currentYear+10)

	return fmt.Sprintf("%02d/%02d", month, year%100)
}

func (f *Fakery) CreditCardCVV(cardType string) string {
	// For Amex CVV - 4 digits, for everyone else - 3 digits
	if strings.ToLower(cardType) == "amex" {
		return f.Numerify("####")
	} else {
		return f.Numerify("###")
	}
}

func (f *Fakery) CreditCard() *CreditCard {
	var c CreditCard

	c.Type = f.CreditCardType()
	c.Number = f.CreditCardNumber(c.Type)
	c.CVV = f.CreditCardCVV(c.Type)
	c.ExpiryDate = f.CreditCardExpiryDate()
	c.Name = f.Name()

	return &c
}

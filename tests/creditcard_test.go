package tests

import (
	"fakery"
	"strconv"
	"strings"
	"testing"
)

func TestCreditCard(t *testing.T) {
	c := fakery.New().CreditCard()
	Expect(t, true, c != nil)
	Expect(t, true, len(c.Type) > 0)
	Expect(t, true, len(c.Number) > 0)
	Expect(t, true, len(c.CVV) == 3 || len(c.CVV) == 4)
	Expect(t, true, len(c.Name) > 0)
	Expect(t, true, len(c.ExpiryDate) > 0)
	// Split expiry date
	items := strings.Split(c.ExpiryDate, "/")
	Expect(t, true, len(items) == 2)
	month, _ := strconv.Atoi(items[0])
	year, _ := strconv.Atoi(items[1])
	Expect(t, true, month >= 1 && month <= 12)
	Expect(t, true, year > 25)
	// Number should be invalid
	//	Expect(t, false, c.Validate())
}


package tests

import (
	"testing"
	"fakery"

	"fmt"
)

func TestSimpleLoading(t *testing.T) {
	f := fakery.NewFromLocale("in_IN")
	var loader fakery.DataLoader
	loader.Init("test.json")
	
	if f.LoadLocale(&loader) != nil {
		fmt.Println("Worked")
	}
}

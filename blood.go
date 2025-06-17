package fakery

import (
	"fmt"
)

type Blood struct {
	Type     string `json:"type"`
	RHFactor string `json:"rh_factor"`
	Base
}

func (b Blood) String() string {
	return fmt.Sprintf("%s%s", b.Type, b.RHFactor)
}

var bloodTypes = []string{"A", "B", "AB", "O"}

func (f *Fakery) BloodType() string {
	typ := f.RandomString(bloodTypes)
	factor := f.OneOf([]string{"+", "-"})
	return fmt.Sprintf("%s%s", typ, factor)
}

func (f *Fakery) Blood() *Blood {
	var b Blood

	b.Type = f.RandomString(bloodTypes)
	b.RHFactor = f.OneOf([]string{"+", "-"})
	return &b
}

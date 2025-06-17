package fakery

import (
	"strings"
)

func (f *Fakery) BinaryString(length int) string {
	var bs strings.Builder

	// First is always 1
	bs.WriteString("1")
	choices := []string{"0", "1"}
	for i := 1; i < length; i++ {
		bs.WriteString(f.OneOf(choices))
	}
	return bs.String()
}

package tests

import (
	"fakery"
	"testing"
)

func TestEmoji(t *testing.T) {
	e := fakery.New().Emoji()
	Expect(t, true, e != nil)
	// Test fields which are never empty
	Expect(t, true, len(e.Symbol) > 0)
	Expect(t, true, len(e.Description) > 0)
	Expect(t, true, len(e.Category) > 0)
	Expect(t, true, len(e.Alias) > 0)
}

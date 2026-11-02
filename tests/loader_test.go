package tests

import (
	"fakery"
	"os"
	"path/filepath"
	"fmt"
	"testing"
)

// writeData creates <root>/<locale>/<fileName> with the given contents.
func writeData(t *testing.T, root, locale, fileName, contents string) {
	t.Helper()
	full := filepath.Join(root, locale, filepath.FromSlash(fileName))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

// Check if the loader reversts to the base locale for a requested
// locale that doesn't exist.
func TestResolveLocaleDoesNotFallbackToBaseLanguage(t *testing.T) {

	// A base-language resource exists, but the requested regional
	// locale resource does not.
	var loader fakery.DataLoader
	loader.Init("names.json")
	
	// LoadLocale should request exactly "fr_CA".
	f := fakery.NewFromLocale("fr_CA")
	got := f.LoadLocale(&loader)

	if got == nil {
		t.Fatal("LoadLocale returned nil")
	}

	if got.GetLocale() != "fr_CA" {
		t.Errorf(
			"LoadLocale(%q) returned locale %q; want exact locale %q",
			"fr_CA",
			got.GetLocale(),
			"fr_CA",
		)
	} else {
		fmt.Println("TestResolveLocaleDoesNotFallbackToBaseLanguage passed")
	}


}

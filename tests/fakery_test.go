
package tests

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"fakery"
)

func TestResolveLocaleDoesNotFallbackToBaseLanguage(t *testing.T) {
	baseDir := t.TempDir()

	// A base-language resource exists, but the requested regional
	// locale resource does not.
	baseLocaleDir := filepath.Join(baseDir, "fr")
	if err := os.MkdirAll(baseLocaleDir, 0o755); err != nil {
		t.Fatal(err)
	}

	resource := filepath.Join(baseLocaleDir, "names.json")
	if err := os.WriteFile(
		resource,
		[]byte(`{"firstName":["Jean"]}`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	loader := fakery.NewDataLoaderWithBaseDir("names.json", baseDir)

	// Existing behavior: load the exact requested locale.
	// Do not silently substitute "fr" for "fr_CA".
	got := loader.ResolveLocale("fr_CA")

	if got == nil {
		t.Fatal("ResolveLocale returned nil")
	}

	if got.Locale() != "fr_CA" {
		t.Errorf(
			"ResolveLocale(%q) returned locale %q; want exact locale %q",
			"fr_CA",
			got.Locale(),
			"fr_CA",
		)
	}

	if err := got.Err(); !errors.Is(err, os.ErrNotExist) {
		t.Errorf(
			"ResolveLocale(%q) error = %v; want os.ErrNotExist",
			"fr_CA",
			err,
		)
	}
}

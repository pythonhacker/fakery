package tests

import (
	"gofakelib"
	"testing"
)

func TestBookTitle(t *testing.T) {
	Expect(t, true, len(gofakelib.New().BookTitle()) > 0)
}

func TestBookGenre(t *testing.T) {
	Expect(t, true, len(gofakelib.New().BookGenre()) > 0)
}

func TestBookPublisher(t *testing.T) {
	Expect(t, true, len(gofakelib.New().BookPublisher()) > 0)
}

func TestBookAuthor(t *testing.T) {
	Expect(t, true, len(gofakelib.New().BookAuthor()) > 0)
}

func TestBookFormat(t *testing.T) {
	Expect(t, true, len(gofakelib.New().BookFormat()) > 0)
}

func TestBookISBN(t *testing.T) {
	isbn := gofakelib.New().BookISBN()
	Expect(t, true, len(isbn) > 0)
	Expect(t, true, gofakelib.New().ValidateISBN(isbn))
}

func TestBook(t *testing.T) {
	b := gofakelib.New().Book()
	Expect(t, true, b != nil)
	// Test fields which are never empty
	Expect(t, true, len(b.Title) > 0)
	Expect(t, true, len(b.Publisher) > 0)
	Expect(t, true, len(b.Author) > 0)
	Expect(t, true, len(b.Format) > 0)
	Expect(t, true, len(b.Genre) > 0)
	isbn := b.ISBN10
	if len(isbn) == 0 {
		isbn = b.ISBN13
	}
	Expect(t, true, gofakelib.New().ValidateISBN(isbn))
	Expect(t, true, b.PageCount > 0)
	Expect(t, true, b.Year >= 1970)
}

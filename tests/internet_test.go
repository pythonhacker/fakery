package tests

import (
	"gofakelib"
	"net/mail"
	"testing"
)

func testEmail(email string) bool {
	if len(email) == 0 {
		return false
	}

	_, err := mail.ParseAddress(email)
	return err == nil
}

func TestEmail(t *testing.T) {
	Expect(t, true, testEmail(gofakelib.New().Email()))
}

func TestEmailWithName(t *testing.T) {
	f := gofakelib.New()
	firstName := f.FirstName()
	lastName := f.LastName()

	Expect(t, true, testEmail(f.EmailWithName(firstName, lastName)))
}

func TestTLD(t *testing.T) {
	Expect(t, true, len(gofakelib.New().GetRandomTLD()) > 0)
}

func TestEmailDomain(t *testing.T) {
	Expect(t, true, len(gofakelib.New().GetRandomEmailDomain()) > 0)
}

package tests

import (
	"fakery"
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
	Expect(t, true, testEmail(fakery.New().Email()))
}

func TestEmailWithName(t *testing.T) {
	f := fakery.New()
	firstName := f.FirstName()
	lastName := f.LastName()

	Expect(t, true, testEmail(f.EmailWithName(firstName, lastName)))
}

func TestTLD(t *testing.T) {
	Expect(t, true, len(fakery.New().TLD()) > 0)
}

func TestEmailDomain(t *testing.T) {
	Expect(t, true, len(fakery.New().EmailDomain()) > 0)
}

func TestFreeEmailDomain(t *testing.T) {
	Expect(t, true, len(fakery.New().FreeEmailDomain()) > 0)
}

func TestUserName(t *testing.T) {
	Expect(t, true, len(fakery.New().UserName()) > 0)
}

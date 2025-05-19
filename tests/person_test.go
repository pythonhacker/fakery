package tests

import (
	"fakery"
	"testing"
)

func TestGender(t *testing.T) {
	Expect(t, true, len(fakery.New().Gender()) > 0)
}

func TestName(t *testing.T) {
	Expect(t, true, len(fakery.New().Name()) > 0)
}

func TestFirstName(t *testing.T) {
	Expect(t, true, len(fakery.New().FirstName()) > 0)
}

func TestLastName(t *testing.T) {
	Expect(t, true, len(fakery.New().LastName()) > 0)
}

func TestJob(t *testing.T) {
	j := fakery.New().Job()
	Expect(t, true, j != nil)
	Expect(t, true, len(j.Title) > 0)
}

func TestPerson(t *testing.T) {
	p := fakery.New().Person()
	Expect(t, true, p != nil)
	// Test fields which are never empty
	Expect(t, true, len(p.FirstName) > 0)
	Expect(t, true, len(p.LastName) > 0)
	Expect(t, true, len(p.Name) > 0)
	Expect(t, true, len(p.Gender) > 0)
	//	Expect(t, true, len(p.Username) > 0)
	Expect(t, true, len(p.Email) > 0)
	Expect(t, true, len(p.Job) > 0)
}

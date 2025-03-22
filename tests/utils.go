package tests

import (
	"testing"
)

// Borrowed from github.com/jaswdr/faker
func Expect(t *testing.T, expected, got interface{}, values ...interface{}) {
	t.Helper()
	if expected != got {
		t.Errorf("\nExpected: (%T) %v \nGot:\t  (%T) %v", expected, expected, got, got)
		if len(values) > 0 {
			for _, v := range values {
				t.Errorf("\n%+v", v)
			}
		}

		t.FailNow()
	}
}

// Borrowed from github.com/jaswdr/faker
func NotExpect(t *testing.T, notExpected, got interface{}, values ...interface{}) {
	t.Helper()
	if notExpected == got {
		t.Errorf("\nNot Expecting: (%T) '%v', but it was", notExpected, notExpected)
		if len(values) > 0 {
			for _, v := range values {
				t.Errorf("\n%+v", v)
			}
		}

		t.FailNow()
	}
}

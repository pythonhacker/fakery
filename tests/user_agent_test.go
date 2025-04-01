package tests

import (
	"github.com/mileusna/useragent"
	"gofakelib"
	"testing"
)

func TestUserAgent(t *testing.T) {
	ua := gofakelib.New().UserAgent()
	Expect(t, true, len(ua) > 0)
	userAgent := useragent.Parse(ua)
	Expect(t, true, len(userAgent.String) > 0)
}

func TestChrome(t *testing.T) {
	Expect(t, true, len(gofakelib.New().Chrome()) > 0)
}

func TestFirefox(t *testing.T) {
	Expect(t, true, len(gofakelib.New().Firefox()) > 0)
}

func TestSafari(t *testing.T) {
	Expect(t, true, len(gofakelib.New().Safari()) > 0)
}

func TestIE(t *testing.T) {
	Expect(t, true, len(gofakelib.New().IE()) > 0)
}

func TestEdge(t *testing.T) {
	Expect(t, true, len(gofakelib.New().Edge()) > 0)
}

func TestOpera(t *testing.T) {
	Expect(t, true, len(gofakelib.New().Opera()) > 0)
}

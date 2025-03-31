// Fake user agents which look realistic
package gofakelib

import (
	"fmt"
	"strings"
)

// Chrome returns a realistic Chrome user agent
func (f *Faker) Chrome() string {
	majorVersion := f.RandIntBetween(45, 138)

	// Minor version, patch, and build
	patchVersion := f.IntRange(10000)
	buildVersion := f.IntRange(1000)

	platform := f.Platform()

	// Mobile-specific modifications
	if strings.Contains(platform, "Android") {
		// For Android, we need a different Chrome format
		chromeVersion := fmt.Sprintf("%d.0.%d.%d", majorVersion, patchVersion, buildVersion)
		return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Mobile Safari/537.36",
			platform, chromeVersion)
	} else if strings.Contains(platform, "iPhone") || strings.Contains(platform, "iPad") || strings.Contains(platform, "iPod") {
		// For iOS, Chrome uses a specific format
		chromeVersion := fmt.Sprintf("%d.0.%d.%d", majorVersion, patchVersion, buildVersion)
		return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/%s Mobile/15E148 Safari/604.1",
			platform, chromeVersion)
	} else {
		// Default desktop Chrome format
		chromeVersion := fmt.Sprintf("%d.0.%d.%d", majorVersion, patchVersion, buildVersion)
		return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Safari/537.36",
			platform, chromeVersion)
	}
}

// Firefox returns a realistic Firefox user agent
func (f *Faker) Firefox() string {

	majorVersion := f.RandIntBetween(45, 136)

	platform := f.Platform()

	// Special case for Android Firefox (Fennec)
	if strings.Contains(platform, "Android") {
		return fmt.Sprintf("Mozilla/5.0 (%s; rv:%d.0) Gecko/20100101 Firefox/%d.0",
			platform, majorVersion, majorVersion)
	} else if strings.Contains(platform, "iPhone") || strings.Contains(platform, "iPad") || strings.Contains(platform, "iPod") {
		// For iOS, Firefox uses a specific format
		return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/605.1.15 (KHTML, like Gecko) FxiOS/%d.0 Mobile/15E148 Safari/605.1.15",
			platform, majorVersion)
	} else {
		// Default desktop Firefox format
		return fmt.Sprintf("Mozilla/5.0 (%s; rv:%d.0) Gecko/20100101 Firefox/%d.0",
			platform, majorVersion, majorVersion)
	}
}

// Safari returns a realistic Safari user agent
func (f *Faker) Safari() string {
	platform := f.Platform()

	// Safari version (corresponds roughly to the OS version)
	majorVersion := f.RandIntBetween(11, 17)
	minorVersion := f.IntRange(4)
	patchVersion := f.IntRange(16)

	safariVersion := fmt.Sprintf("%d.%d.%d", majorVersion, minorVersion, patchVersion)

	// Special format for iOS Safari
	if strings.Contains(platform, "iPhone") || strings.Contains(platform, "iPad") || strings.Contains(platform, "iPod") {
		return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/%s Mobile/15E148 Safari/604.1",
			platform, safariVersion)
	} else if strings.Contains(platform, "Macintosh") {
		// macOS Safari
		return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/%s Safari/605.1.15",
			platform, safariVersion)
	} else {
		// Non-Apple platform (rare but can happen)
		return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Version/%s Safari/537.36",
			platform, safariVersion)
	}
}

// IE returns a realistic Internet Explorer user agent
func (f *Faker) IE() string {
	// IE only runs on Windows
	windows := f.Windows()

	// IE versions 8-11 (older versions are too outdated)
	version := f.RandIntBetween(7, 11)

	// Trident version maps to IE version
	tridentVersion := version - 4
	if tridentVersion < 4 {
		tridentVersion = 4
	}

	if version >= 11 {
		return fmt.Sprintf("Mozilla/5.0 (%s; Trident/%d.0; rv:11.0) like Gecko",
			windows, tridentVersion)
	} else {
		return fmt.Sprintf("Mozilla/5.0 (%s; Trident/%d.0; MSIE %d.0; rv:11.0) like Gecko",
			windows, tridentVersion, version)
	}
}

// Edge returns a realistic Microsoft Edge user agent
func (f *Faker) Edge() string {
	platform := f.Windows() // Edge was originally Windows-only

	// Decide between legacy Edge and Chromium-based Edge
	if f.RollDice() < 3 {
		// Legacy Edge (EdgeHTML)
		edgeVersion := fmt.Sprintf("%d.%d", f.RandIntBetween(12, 18), f.RandIntBetween(10000, 20000))
		return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3282.140 Safari/537.36 Edge/%s",
			platform, edgeVersion)
	} else {
		// Modern Edge (Chromium-based)
		majorVersion := f.RandIntBetween(78, 124) // Chromium Edge started at version 79
		//		minorVersion := f.IntRange(11)
		patchVersion := f.IntRange(1001)

		// Platform can be Windows, Mac, or even Linux for newer Edge
		if f.RollDice() < 3 {
			platform = f.Mac()
		} else {
			platform = f.Linux()
		}

		edgeVersion := fmt.Sprintf("%d.0.%d.%d", majorVersion, patchVersion, f.IntRange(1000))
		return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Safari/537.36 Edg/%s",
			platform, edgeVersion, edgeVersion)
	}
}

// Opera returns a realistic Opera user agent
func (f *Faker) Opera() string {
	platform := f.Platform()

	// Modern Opera is Chromium-based
	majorVersion := f.RandIntBetween(59, 100)
	minorVersion := f.IntRange(10)
	patchVersion := f.IntRange(10000)

	chromeVersion := fmt.Sprintf("%d.0.%d.%d", majorVersion, patchVersion, f.IntRange(1000))
	operaVersion := fmt.Sprintf("%d.%d.%d.%d", majorVersion, minorVersion, f.IntRange(1000), f.IntRange(1000))

	// Mobile Opera
	if strings.Contains(platform, "Android") {
		return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Mobile Safari/537.36 OPR/%s",
			platform, chromeVersion, operaVersion)
	} else if strings.Contains(platform, "iPhone") || strings.Contains(platform, "iPad") || strings.Contains(platform, "iPod") {
		return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0 Mobile/15E148 Safari/604.1 OPiOS/%s",
			platform, operaVersion)
	} else {
		// Desktop Opera
		return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Safari/537.36 OPR/%s",
			platform, chromeVersion, operaVersion)
	}
}

// Return a random browser user agent
func (f *Faker) UserAgent() string {

	switch f.IntRange(5) {
	case 0:
		return f.Chrome()
	case 1:
		return f.Firefox()
	case 2:
		return f.Safari()
	case 3:
		return f.IE()
	default:
		return f.Edge()
	}
}

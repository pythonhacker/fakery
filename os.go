// Random operating system related functions
package fakery

import (
	"fmt"
)

func (f *Fakery) WindowsVersion() string {
	versions := []string{
		"Windows NT 10.0; Win64; x64",
		"Windows NT 10.0; WOW64",
		"Windows NT 6.3; Win64; x64",
		"Windows NT 6.3; WOW64",
		"Windows NT 6.2; Win64; x64",
		"Windows NT 6.2; WOW64",
		"Windows NT 6.1; Win64; x64",
		"Windows NT 6.1; WOW64",
		"Windows NT 6.1",
		"Windows NT 6.0",
		"Windows NT 5.1",
	}
	return f.RandomString(versions)
}

// Mac returns a random Mac OS version string
func (f *Fakery) MacVersion() string {
	// Generate OS X or macOS version
	majorVersion := 10
	// Version 10.9 (Mavericks) through 10.15 (Catalina) or 11.0+ (Big Sur and newer)
	if f.RollDice() < 3 {
		majorVersion = f.RandIntBetween(10, 14) // macOS 11-14 (Big Sur through Sonoma)
		return fmt.Sprintf("Macintosh; Intel Mac OS X %d_%d_%d",
			majorVersion,
			f.IntRange(8),
			f.IntRange(20))
	} else {
		minorVersion := f.RandIntBetween(8, 15)
		patchVersion := f.IntRange(8)
		return fmt.Sprintf("Macintosh; Intel Mac OS X 10_%d_%d", minorVersion, patchVersion)
	}
}

// Linux returns a random Linux version string
// this is to be used with browser user agents only
func (f *Fakery) LinuxVersion() string {
	versions := []string{
		"X11; Linux x86_64",
		"X11; Ubuntu; Linux x86_64",
		"X11; Linux i686",
		"X11; Fedora; Linux x86_64",
		"X11; CentOS; Linux x86_64",
		"X11; Debian; Linux x86_64",
	}
	return f.RandomString(versions)
}

// Android returns a random Android version string
func (f *Fakery) AndroidVersion() string {
	versions := []string{
		"Android 13; Mobile",
		"Android 12; Mobile",
		"Android 11; Mobile",
		"Android 10; Mobile",
		"Android 9; Mobile",
		"Android 8.1.0; Mobile",
		"Android 8.0.0; Mobile",
		"Android 7.1.2; Mobile",
		"Android 7.0; Mobile",
		"Android 6.0.1; Mobile",
		"Android 6.0; Mobile",
		"Android 5.1.1; Mobile",
	}
	return f.RandomString(versions)
}

// iOS returns a random iOS version string
func (f *Fakery) iOSVersion() string {
	device := f.RandomString([]string{"iPhone", "iPad", "iPod"})

	// iOS versions with corresponding device generations
	iosVersions := []string{
		"17_4_1", "17_0", "16_6", "16_5", "16_4", "16_3_1", "16_0",
		"15_7", "15_6_1", "15_5", "15_4_1", "15_0",
		"14_8", "14_7_1", "14_6", "14_4_2", "14_0",
		"13_7", "13_6_1", "13_5_1", "13_4_1", "13_3_1", "13_0",
		"12_5_5", "12_4_1", "12_3_1", "12_2", "12_1", "12_0",
	}

	// CPU values for iOS devices
	cpus := []string{
		"CPU iPhone OS %s like Mac OS X",
		"CPU OS %s like Mac OS X",
	}

	iosVersion := f.RandomString(iosVersions)
	cpuFormat := f.RandomString(cpus)
	cpu := fmt.Sprintf(cpuFormat, iosVersion)

	return fmt.Sprintf("%s; %s", device, cpu)
}

// Platform returns a random platform version string
func (f *Fakery) PlatformVersion() string {
	switch f.IntRange(5) {
	case 0:
		return f.WindowsVersion()
	case 1:
		return f.MacVersion()
	case 2:
		return f.LinuxVersion()
	case 3:
		return f.AndroidVersion()
	default:
		return f.iOSVersion()
	}
}

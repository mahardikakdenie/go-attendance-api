package utils

import "time"

// WIB is the fixed zone for Western Indonesia Time (UTC+7)
var WIB = time.FixedZone("WIB", 7*3600)

// Now returns the current time in WIB (UTC+7)
func Now() time.Time {
	return time.Now().In(WIB)
}

// GetWIBLocation returns the *time.Location for WIB
func GetWIBLocation() *time.Location {
	return WIB
}

// ParseDateWIB parses a date string in "2006-01-02" format into WIB timezone.
func ParseDateWIB(dateStr string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", dateStr, WIB)
}

// ParseTimeWIB parses a time string with a given layout into WIB timezone.
func ParseTimeWIB(layout, timeStr string) (time.Time, error) {
	return time.ParseInLocation(layout, timeStr, WIB)
}

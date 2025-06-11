package utils

import (
	"fmt"
	"strings"
	"time"
)

// ParseDateTime parses various date/time string formats and returns a time.Time
// When no timezone is specified, it uses the server's local timezone
func ParseDateTime(s string) (time.Time, error) {
	// Formats with timezone information (try these first)
	formatsWithTZ := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.999999999Z07:00",
		"2006-01-02T15:04:05Z07:00",
	}

	// Formats without timezone information (use local timezone)
	formatsWithoutTZ := []string{
		"2006-01-02T15:04:05.999999999",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"20060102",
		"2006/01/02",
		"2006/1/2",
		"2006.01.02",
		"2006.1.2",
		"2006-01-02",
		"2006-1-2",
	}

	// Try formats with timezone first
	for _, format := range formatsWithTZ {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	// Try formats without timezone using local timezone
	for _, format := range formatsWithoutTZ {
		if t, err := time.ParseInLocation(format, s, time.Local); err == nil {
			return t, nil
		}
	}

	// Handle special compact format like "1971-0618"
	if len(s) >= 8 && strings.Contains(s, "-") {
		parts := strings.Split(s, "-")
		if len(parts) == 2 && len(parts[0]) == 4 {
			// Try to parse as YYYY-MMDD
			compactDate := parts[0] + parts[1]
			if len(compactDate) == 8 {
				if t, err := time.ParseInLocation("20060102", compactDate, time.Local); err == nil {
					return t, nil
				}
			}
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date/time: %s", s)
}
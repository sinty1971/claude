package models

import "time"

// TimeParseRequest represents a request to parse time string
// @Description Request for parsing various date/time formats
type TimeParseRequest struct {
	// Time string to parse
	TimeString string `json:"time_string" example:"2024-01-15T10:30:00"`
}

// TimeParseResponse represents the parsed time response
// @Description Response containing parsed time in various formats
type TimeParseResponse struct {
	// Original input string
	Original string `json:"original" example:"2024-01-15T10:30:00"`
	// Parsed time in RFC3339 format
	RFC3339 string `json:"rfc3339" example:"2024-01-15T10:30:00Z"`
	// Unix timestamp
	Unix int64 `json:"unix" example:"1705318200"`
	// Human readable format
	Readable string `json:"readable" example:"January 15, 2024 10:30 AM"`
	// Time zone used
	TimeZone string `json:"timezone" example:"Local"`
}

// TimeFormat represents supported time format
// @Description Supported time format information
type TimeFormat struct {
	// Format name
	Name string `json:"name" example:"RFC3339"`
	// Format pattern
	Pattern string `json:"pattern" example:"2006-01-02T15:04:05Z07:00"`
	// Example value
	Example string `json:"example" example:"2024-01-15T10:30:00Z"`
}

// SupportedFormatsResponse represents list of supported time formats
// @Description List of all supported date/time formats
type SupportedFormatsResponse struct {
	// List of supported formats
	Formats []TimeFormat `json:"formats"`
}

// ConvertDateTimeInstant converts a parsed time to an instant representation
func ConvertDateTimeInstant(t Timestamp) TimeParseResponse {
	return TimeParseResponse{
		Original: t.Time.Format("2006-01-02T15:04:05"),
		RFC3339:  t.Time.Format(time.RFC3339),
		Unix:     t.Time.Unix(),
		Readable: t.Time.Format("January 2, 2006 3:04 PM"),
		TimeZone: t.Time.Location().String(),
	}
}
package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Instant represents a precise moment in time with nanosecond precision
type Instant struct {
	Seconds        int64 `json:"seconds"`         // Unix timestamp in seconds
	Nanos          int32 `json:"nanos"`           // Nanosecond component (0-999,999,999)
	TimezoneOffset int32 `json:"timezone_offset"` // Timezone offset in seconds from UTC
}

// NewInstant creates a new Instant from the current time
func NewInstant() *Instant {
	now := time.Now()
	return &Instant{
		Seconds:        now.Unix(),
		Nanos:          int32(now.Nanosecond()),
		TimezoneOffset: int32(now.Local().Sub(now.UTC()).Seconds()),
	}
}

// NewInstantFromTime creates an Instant from a time.Time
func NewInstantFromTime(t time.Time) *Instant {
	_, offset := t.Zone()
	return &Instant{
		Seconds:        t.Unix(),
		Nanos:          int32(t.Nanosecond()),
		TimezoneOffset: int32(offset),
	}
}

// NewInstantFromUnixNanos creates an Instant from Unix nanoseconds
func NewInstantFromUnixNanos(nanos int64) *Instant {
	seconds := nanos / 1_000_000_000
	remainingNanos := nanos % 1_000_000_000
	return &Instant{
		Seconds:        seconds,
		Nanos:          int32(remainingNanos),
		TimezoneOffset: 0, // UTC by default
	}
}

// ParseInstant parses various date/time string formats with optional nanosecond precision
func ParseInstant(s string) (*Instant, error) {
	// List of supported formats in priority order
	formats := []string{
		time.RFC3339Nano,                 // 2006-01-02T15:04:05.999999999Z07:00
		time.RFC3339,                     // 2006-01-02T15:04:05Z07:00
		"2006-01-02T15:04:05",           // 2006-01-02T15:04:05
		"2006-01-02 15:04:05",           // 2006-01-02 15:04:05
		"2006-01-02",                    // 2006-01-02
		"2006-0102",                     // 2006-0102 (compact with dash)
		"20060102",                      // 20060102 (YYYYMMDD)
		"2006/01/02",                    // 2006/01/02
		"2006/01/02 15:04:05",           // 2006/01/02 15:04:05
		"2006.01.02",                    // 2006.01.02
		"2006.01.02 15:04:05",           // 2006.01.02 15:04:05
		"02-01-2006",                    // DD-MM-YYYY
		"02/01/2006",                    // DD/MM/YYYY
		"01/02/2006",                    // MM/DD/YYYY
		"02.01.2006",                    // DD.MM.YYYY
	}

	// Try each format
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			// If no timezone info in the original string, use local timezone
			if !hasTimezoneInfo(s) {
				t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.Local)
			}
			return NewInstantFromTime(t), nil
		}
	}
	
	// If no standard format works, try custom parsing for special cases
	if customInstant, err := parseCustomFormat(s); err == nil {
		return customInstant, nil
	}
	
	return nil, fmt.Errorf("failed to parse time: unsupported format '%s'", s)
}

// parseCustomFormat handles special date formats like "1971-0618"
func parseCustomFormat(s string) (*Instant, error) {
	// Handle YYYY-MMDD format (like "1971-0618")
	if len(s) == 9 && s[4] == '-' {
		year := s[0:4]
		monthDay := s[5:9]
		if len(monthDay) == 4 {
			month := monthDay[0:2]
			day := monthDay[2:4]
			reformatted := year + "-" + month + "-" + day
			if t, err := time.Parse("2006-01-02", reformatted); err == nil {
				// Use local timezone for custom parsed dates
				t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.Local)
				return NewInstantFromTime(t), nil
			}
		}
	}
	
	// Handle YYYY-M-D, YYYY.M.D, YYYY/M/D formats (like "1971-6-18", "1971.6.18", "1971/6/18")
	separators := []string{"-", ".", "/"}
	for _, sep := range separators {
		parts := strings.Split(s, sep)
		if len(parts) == 3 {
			year := parts[0]
			month := parts[1]
			day := parts[2]
			
			// Pad month and day with leading zeros if needed
			if len(month) == 1 {
				month = "0" + month
			}
			if len(day) == 1 {
				day = "0" + day
			}
			
			reformatted := year + "-" + month + "-" + day
			if t, err := time.Parse("2006-01-02", reformatted); err == nil {
				return NewInstantFromTime(t), nil
			}
		}
	}
	
	return nil, fmt.Errorf("unsupported custom format")
}

// hasTimezoneInfo checks if the time string contains timezone information
func hasTimezoneInfo(s string) bool {
	// Check for Z at the end (UTC indicator)
	if strings.HasSuffix(s, "Z") {
		return true
	}
	
	// Look for timezone offset patterns like +09:00, -05:00, +0900, -0500
	// These should appear after time part, not in date part
	timezonePattern := []string{"+", "-"}
	
	for _, pattern := range timezonePattern {
		if idx := strings.LastIndex(s, pattern); idx != -1 {
			// Check if this + or - appears after a time (contains T or space and numbers)
			beforePattern := s[:idx]
			afterPattern := s[idx+1:]
			
			// If it's timezone, there should be time before it and numbers after
			if (strings.Contains(beforePattern, "T") || strings.Contains(beforePattern, " ")) && 
			   len(afterPattern) >= 2 && isDigit(afterPattern[0]) && isDigit(afterPattern[1]) {
				return true
			}
		}
	}
	
	return false
}

// isDigit checks if a byte represents a digit
func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

// ToTime converts the Instant to a time.Time
func (i *Instant) ToTime() time.Time {
	location := time.FixedZone("", int(i.TimezoneOffset))
	return time.Unix(i.Seconds, int64(i.Nanos)).In(location)
}

// ToUTC converts the Instant to UTC time.Time
func (i *Instant) ToUTC() time.Time {
	return time.Unix(i.Seconds, int64(i.Nanos)).UTC()
}

// UnixNanos returns the time as Unix nanoseconds
func (i *Instant) UnixNanos() int64 {
	return i.Seconds*1_000_000_000 + int64(i.Nanos)
}

// AddNanos adds nanoseconds to the instant, handling overflow properly
func (i *Instant) AddNanos(nanos int64) *Instant {
	totalNanos := int64(i.Nanos) + nanos
	
	// Handle overflow/underflow
	extraSeconds := totalNanos / 1_000_000_000
	remainingNanos := totalNanos % 1_000_000_000
	
	// Ensure nanos is always positive
	if remainingNanos < 0 {
		extraSeconds--
		remainingNanos += 1_000_000_000
	}
	
	return &Instant{
		Seconds:        i.Seconds + extraSeconds,
		Nanos:          int32(remainingNanos),
		TimezoneOffset: i.TimezoneOffset,
	}
}

// AddDuration adds a time.Duration to the instant
func (i *Instant) AddDuration(d time.Duration) *Instant {
	return i.AddNanos(d.Nanoseconds())
}

// Sub returns the duration between two instants
func (i *Instant) Sub(other *Instant) time.Duration {
	deltaSeconds := i.Seconds - other.Seconds
	deltaNanos := int64(i.Nanos) - int64(other.Nanos)
	
	totalNanos := deltaSeconds*1_000_000_000 + deltaNanos
	return time.Duration(totalNanos)
}

// Before reports whether the instant i is before other
func (i *Instant) Before(other *Instant) bool {
	if i.Seconds != other.Seconds {
		return i.Seconds < other.Seconds
	}
	return i.Nanos < other.Nanos
}

// After reports whether the instant i is after other
func (i *Instant) After(other *Instant) bool {
	if i.Seconds != other.Seconds {
		return i.Seconds > other.Seconds
	}
	return i.Nanos > other.Nanos
}

// Equal reports whether i and other represent the same time instant
func (i *Instant) Equal(other *Instant) bool {
	return i.Seconds == other.Seconds && i.Nanos == other.Nanos
}

// Format formats the instant using the given layout
func (i *Instant) Format(layout string) string {
	return i.ToTime().Format(layout)
}

// RFC3339Nano returns the instant in RFC3339 format with nanosecond precision
func (i *Instant) RFC3339Nano() string {
	return i.ToTime().Format(time.RFC3339Nano)
}

// String returns a string representation of the instant
func (i *Instant) String() string {
	return i.RFC3339Nano()
}

// MillisTop3 returns the top 3 digits of nanoseconds (milliseconds part)
func (i *Instant) MillisTop3() int16 {
	return int16(i.Nanos / 1_000_000)
}

// MicrosTop6 returns the top 6 digits of nanoseconds (microseconds part)
func (i *Instant) MicrosTop6() int32 {
	return i.Nanos / 1_000
}

// SubsecToNanos converts subsecond string to nanoseconds with proper padding
func SubsecToNanos(subsec string) int32 {
	if subsec == "" {
		return 0
	}
	
	// Pad or truncate to 9 digits
	if len(subsec) < 9 {
		subsec = subsec + strings.Repeat("0", 9-len(subsec))
	} else if len(subsec) > 9 {
		subsec = subsec[:9]
	}
	
	nanos, _ := strconv.ParseInt(subsec, 10, 32)
	return int32(nanos)
}

// ToBytes serializes the instant to a 16-byte array (big-endian)
func (i *Instant) ToBytes() [16]byte {
	var bytes [16]byte
	
	// Seconds (8 bytes, big-endian)
	for j := 0; j < 8; j++ {
		bytes[j] = byte(i.Seconds >> (8 * (7 - j)))
	}
	
	// Nanos (4 bytes, big-endian)
	for j := 0; j < 4; j++ {
		bytes[8+j] = byte(i.Nanos >> (8 * (3 - j)))
	}
	
	// Timezone offset (4 bytes, big-endian)
	for j := 0; j < 4; j++ {
		bytes[12+j] = byte(i.TimezoneOffset >> (8 * (3 - j)))
	}
	
	return bytes
}

// FromBytes creates an Instant from a 16-byte array (big-endian)
func FromBytes(bytes [16]byte) *Instant {
	// Reconstruct seconds
	var seconds int64
	for i := 0; i < 8; i++ {
		seconds = (seconds << 8) | int64(bytes[i])
	}
	
	// Reconstruct nanos
	var nanos int32
	for i := 0; i < 4; i++ {
		nanos = (nanos << 8) | int32(bytes[8+i])
	}
	
	// Reconstruct timezone offset
	var timezoneOffset int32
	for i := 0; i < 4; i++ {
		timezoneOffset = (timezoneOffset << 8) | int32(bytes[12+i])
	}
	
	return &Instant{
		Seconds:        seconds,
		Nanos:          nanos,
		TimezoneOffset: timezoneOffset,
	}
}

// MarshalJSON implements json.Marshaler
func (i *Instant) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"seconds":         i.Seconds,
		"nanos":           i.Nanos,
		"timezone_offset": i.TimezoneOffset,
		"rfc3339":         i.RFC3339Nano(),
	})
}

// UnmarshalJSON implements json.Unmarshaler
func (i *Instant) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	
	if seconds, ok := raw["seconds"].(float64); ok {
		i.Seconds = int64(seconds)
	}
	if nanos, ok := raw["nanos"].(float64); ok {
		i.Nanos = int32(nanos)
	}
	if offset, ok := raw["timezone_offset"].(float64); ok {
		i.TimezoneOffset = int32(offset)
	}
	
	return nil
}
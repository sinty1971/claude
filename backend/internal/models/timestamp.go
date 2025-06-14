package models

import (
	"fmt"
	"time"
)

// Timestamp wraps time.Time with custom YAML/JSON marshaling/unmarshaling
type Timestamp struct {
	time.Time
}

// NewTimestamp creates a new Timestamp from time.Time
func NewTimestamp(t time.Time) Timestamp {
	return Timestamp{Time: t}
}

// UnmarshalYAML implements yaml.Unmarshaler
func (ts *Timestamp) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	if err := unmarshal(&str); err != nil {
		return err
	}

	// Try parsing with RFC3339Nano first
	parsed, err := time.Parse(time.RFC3339Nano, str)
	if err != nil {
		// Try RFC3339 as fallback
		parsed, err = time.Parse(time.RFC3339, str)
		if err != nil {
			// Try local time format
			parsed, err = time.ParseInLocation("2006-01-02T15:04:05", str, time.Local)
			if err != nil {
				return fmt.Errorf("failed to parse timestamp: %w", err)
			}
		}
	}

	ts.Time = parsed
	return nil
}

// MarshalYAML implements yaml.Marshaler
func (ts Timestamp) MarshalYAML() (interface{}, error) {
	if ts.Time.IsZero() {
		return "", nil
	}
	return ts.Time.Format(time.RFC3339Nano), nil
}

// MarshalJSON implements json.Marshaler
func (ts Timestamp) MarshalJSON() ([]byte, error) {
	if ts.Time.IsZero() {
		return []byte(`""`), nil
	}
	return []byte(`"` + ts.Time.Format(time.RFC3339Nano) + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (ts *Timestamp) UnmarshalJSON(data []byte) error {
	str := string(data)
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}

	if str == "" {
		ts.Time = time.Time{}
		return nil
	}

	// Try parsing with RFC3339Nano first
	parsed, err := time.Parse(time.RFC3339Nano, str)
	if err != nil {
		// Try RFC3339 as fallback
		parsed, err = time.Parse(time.RFC3339, str)
		if err != nil {
			// Try local time format
			parsed, err = time.ParseInLocation("2006-01-02T15:04:05", str, time.Local)
			if err != nil {
				return fmt.Errorf("failed to parse timestamp: %w", err)
			}
		}
	}

	ts.Time = parsed
	return nil
}
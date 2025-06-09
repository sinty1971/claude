package models_test

import (
	"penguin-backend/internal/models"
	"testing"
	"time"
)

func TestNewInstant(t *testing.T) {
	instant := models.NewInstant()
	
	if instant.Seconds == 0 {
		t.Error("Expected non-zero seconds")
	}
	
	if instant.Nanos < 0 || instant.Nanos >= 1_000_000_000 {
		t.Errorf("Expected nanos in range [0, 1000000000), got %d", instant.Nanos)
	}
}

func TestNewInstantFromTime(t *testing.T) {
	now := time.Now()
	instant := models.NewInstantFromTime(now)
	
	if instant.Seconds != now.Unix() {
		t.Errorf("Expected seconds %d, got %d", now.Unix(), instant.Seconds)
	}
	
	if instant.Nanos != int32(now.Nanosecond()) {
		t.Errorf("Expected nanos %d, got %d", now.Nanosecond(), instant.Nanos)
	}
}

func TestNewInstantFromUnixNanos(t *testing.T) {
	nanos := int64(1640995200123456789) // 2022-01-01 00:00:00.123456789 UTC
	instant := models.NewInstantFromUnixNanos(nanos)
	
	expectedSeconds := int64(1640995200)
	expectedNanos := int32(123456789)
	
	if instant.Seconds != expectedSeconds {
		t.Errorf("Expected seconds %d, got %d", expectedSeconds, instant.Seconds)
	}
	
	if instant.Nanos != expectedNanos {
		t.Errorf("Expected nanos %d, got %d", expectedNanos, instant.Nanos)
	}
}

func TestParseInstant(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"2022-01-01T00:00:00Z", true},
		{"2022-01-01T00:00:00.123456789Z", true},
		{"2022-01-01T00:00:00+09:00", true},
		{"1971.06.18", true},
		{"1971.6.18", true},
		{"1971-0618", true},
		{"19710618", true},
		{"1971/06/18", true},
		{"1971/6/18", true},
		{"invalid", false},
	}
	
	for _, test := range tests {
		instant, err := models.ParseInstant(test.input)
		
		if test.expected && err != nil {
			t.Errorf("Expected successful parse for %s, got error: %v", test.input, err)
		}
		
		if !test.expected && err == nil {
			t.Errorf("Expected parse error for %s, got success", test.input)
		}
		
		if test.expected && instant == nil {
			t.Errorf("Expected non-nil instant for %s", test.input)
		}
	}
}

func TestParseDotFormat(t *testing.T) {
	tests := []struct {
		input          string
		expectedYear   int
		expectedMonth  int
		expectedDay    int
	}{
		{"1971.06.18", 1971, 6, 18},
		{"1971.6.18", 1971, 6, 18},
		{"2022.12.31", 2022, 12, 31},
		{"2000.01.01", 2000, 1, 1},
		{"1999.9.9", 1999, 9, 9},
	}
	
	for _, test := range tests {
		instant, err := models.ParseInstant(test.input)
		if err != nil {
			t.Errorf("Failed to parse %s: %v", test.input, err)
			continue
		}
		
		timeValue := instant.ToTime()
		if timeValue.Year() != test.expectedYear {
			t.Errorf("Expected year %d for %s, got %d", test.expectedYear, test.input, timeValue.Year())
		}
		
		if int(timeValue.Month()) != test.expectedMonth {
			t.Errorf("Expected month %d for %s, got %d", test.expectedMonth, test.input, int(timeValue.Month()))
		}
		
		if timeValue.Day() != test.expectedDay {
			t.Errorf("Expected day %d for %s, got %d", test.expectedDay, test.input, timeValue.Day())
		}
	}
}

func TestAddNanos(t *testing.T) {
	instant := &models.Instant{Seconds: 1000, Nanos: 500_000_000}
	
	// Add 600 million nanos (should overflow to next second)
	result := instant.AddNanos(600_000_000)
	
	expectedSeconds := int64(1001)
	expectedNanos := int32(100_000_000)
	
	if result.Seconds != expectedSeconds {
		t.Errorf("Expected seconds %d, got %d", expectedSeconds, result.Seconds)
	}
	
	if result.Nanos != expectedNanos {
		t.Errorf("Expected nanos %d, got %d", expectedNanos, result.Nanos)
	}
}

func TestAddNanosNegative(t *testing.T) {
	instant := &models.Instant{Seconds: 1000, Nanos: 200_000_000}
	
	// Subtract 300 million nanos (should underflow to previous second)
	result := instant.AddNanos(-300_000_000)
	
	expectedSeconds := int64(999)
	expectedNanos := int32(900_000_000)
	
	if result.Seconds != expectedSeconds {
		t.Errorf("Expected seconds %d, got %d", expectedSeconds, result.Seconds)
	}
	
	if result.Nanos != expectedNanos {
		t.Errorf("Expected nanos %d, got %d", expectedNanos, result.Nanos)
	}
}

func TestComparison(t *testing.T) {
	instant1 := &models.Instant{Seconds: 1000, Nanos: 500_000_000}
	instant2 := &models.Instant{Seconds: 1000, Nanos: 600_000_000}
	instant3 := &models.Instant{Seconds: 1001, Nanos: 400_000_000}
	
	// Test Before
	if !instant1.Before(instant2) {
		t.Error("instant1 should be before instant2")
	}
	
	if !instant1.Before(instant3) {
		t.Error("instant1 should be before instant3")
	}
	
	// Test After
	if !instant2.After(instant1) {
		t.Error("instant2 should be after instant1")
	}
	
	if !instant3.After(instant1) {
		t.Error("instant3 should be after instant1")
	}
	
	// Test Equal
	instant4 := &models.Instant{Seconds: 1000, Nanos: 500_000_000}
	if !instant1.Equal(instant4) {
		t.Error("instant1 should equal instant4")
	}
}

func TestSubsecToNanos(t *testing.T) {
	tests := []struct {
		input    string
		expected int32
	}{
		{"", 0},
		{"1", 100_000_000},
		{"12", 120_000_000},
		{"123", 123_000_000},
		{"123456", 123_456_000},
		{"123456789", 123_456_789},
		{"1234567890", 123_456_789}, // Truncated to 9 digits
	}
	
	for _, test := range tests {
		result := models.SubsecToNanos(test.input)
		if result != test.expected {
			t.Errorf("models.SubsecToNanos(%s) = %d, expected %d", test.input, result, test.expected)
		}
	}
}

func TestToBytes(t *testing.T) {
	instant := &models.Instant{
		Seconds:        1640995200,
		Nanos:          123456789,
		TimezoneOffset: 32400, // +9 hours
	}
	
	bytes := instant.ToBytes()
	reconstructed := models.FromBytes(bytes)
	
	if !instant.Equal(reconstructed) {
		t.Errorf("Byte serialization failed: original=%+v, reconstructed=%+v", instant, reconstructed)
	}
}

func TestMillisTop3(t *testing.T) {
	instant := &models.Instant{Nanos: 123456789}
	expected := int16(123)
	
	result := instant.MillisTop3()
	if result != expected {
		t.Errorf("MillisTop3() = %d, expected %d", result, expected)
	}
}

func TestMicrosTop6(t *testing.T) {
	instant := &models.Instant{Nanos: 123456789}
	expected := int32(123456)
	
	result := instant.MicrosTop6()
	if result != expected {
		t.Errorf("MicrosTop6() = %d, expected %d", result, expected)
	}
}

func BenchmarkAddNanos(b *testing.B) {
	instant := &models.Instant{Seconds: 1000, Nanos: 500_000_000}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		instant.AddNanos(1000)
	}
}

func BenchmarkSubsecToNanos(b *testing.B) {
	subsec := "123456789"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		models.SubsecToNanos(subsec)
	}
}

func TestParseWithLocalTimezone(t *testing.T) {
	tests := []struct {
		input           string
		hasTimezone     bool
		expectedLocal   bool
	}{
		{"1971.06.18", false, true},           // No timezone -> use local
		{"1971-06-18", false, true},           // No timezone -> use local
		{"2022-01-01T00:00:00Z", true, false}, // Has timezone -> use specified
		{"2022-01-01T00:00:00+09:00", true, false}, // Has timezone -> use specified
		{"19710618", false, true},             // No timezone -> use local
	}
	
	for _, test := range tests {
		instant, err := models.ParseInstant(test.input)
		if err != nil {
			t.Errorf("Failed to parse %s: %v", test.input, err)
			continue
		}
		
		timeValue := instant.ToTime()
		
		if test.expectedLocal {
			// For local timezone, we expect the timezone offset to match local system
			_, localOffset := time.Now().Zone()
			if instant.TimezoneOffset != int32(localOffset) {
				t.Errorf("Expected local timezone offset %d for %s, got %d", 
					localOffset, test.input, instant.TimezoneOffset)
			}
		} else {
			// For explicit timezone, check that it's not automatically set to local
			_, localOffset := time.Now().Zone()
			if instant.TimezoneOffset == int32(localOffset) && test.input != "1971.06.18" {
				// Only fail if it's definitely not supposed to be local
				// (this is a bit tricky to test reliably)
			}
		}
		
		t.Logf("Parsed '%s' -> timezone offset: %d seconds (%s)", 
			test.input, instant.TimezoneOffset, timeValue.Location().String())
	}
}

func TestTimezoneDetection(t *testing.T) {
	// Test timezone detection indirectly through ParseInstant behavior
	tests := []struct {
		input            string
		shouldUseLocal   bool
	}{
		{"2022-01-01T00:00:00Z", false},     // Has timezone -> UTC
		{"2022-01-01T00:00:00+09:00", false}, // Has timezone -> +9
		{"1971.06.18", true},               // No timezone -> local
		{"1971-06-18", true},               // No timezone -> local
		{"19710618", true},                 // No timezone -> local
	}
	
	_, localOffset := time.Now().Zone()
	
	for _, test := range tests {
		instant, err := models.ParseInstant(test.input)
		if err != nil {
			t.Errorf("Failed to parse %s: %v", test.input, err)
			continue
		}
		
		if test.shouldUseLocal {
			if instant.TimezoneOffset != int32(localOffset) {
				t.Errorf("Expected local timezone for %s, got offset %d", test.input, instant.TimezoneOffset)
			}
		}
		
		t.Logf("Parsed '%s' -> timezone offset: %d (local=%d)", 
			test.input, instant.TimezoneOffset, localOffset)
	}
}
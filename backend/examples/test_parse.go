package main

import (
	"fmt"
	"penguin-backend/internal/utils"
	"time"
)

func main() {
	testCases := []string{
		"1971-0618",
		"1971-06-18",
		"1971-06-18T00:00:00Z",
		"19710618",
		"1971.06.18",
		"1971.6.18",
		"1971/06/18",
		"1971/6/18",
	}
	
	fmt.Println("=== Basic Date Parsing Tests ===")
	for _, test := range testCases {
		t, err := utils.ParseDateTime(test)
		if err != nil {
			fmt.Printf("'%s' -> Error: %v\n", test, err)
		} else {
			fmt.Printf("'%s' -> Success: %s (Location: %s, UTC: %s)\n", 
				test, 
				t.Format("2006-01-02 15:04:05 MST"),
				t.Location(),
				t.UTC().Format("2006-01-02 15:04:05"))
		}
	}
	
	// Offset tests
	offsetTestCases := []string{
		"2023-06-15T12:30:45Z",           // UTC
		"2023-06-15T12:30:45+09:00",      // JST
		"2023-06-15T12:30:45-05:00",      // EST
		"2023-06-15T12:30:45+00:00",      // UTC with explicit offset
		"2023-06-15T12:30:45.123Z",       // With milliseconds
		"2023-06-15T12:30:45.123456789Z", // With nanoseconds
		"2023-06-15T12:30:45.123456789+09:00", // With nanoseconds and offset
	}
	
	fmt.Println("\n=== DateTime with Offset Tests ===")
	for _, test := range offsetTestCases {
		t, err := utils.ParseDateTime(test)
		if err != nil {
			fmt.Printf("'%s' -> Error: %v\n", test, err)
		} else {
			fmt.Printf("'%s' -> Success: %s (UTC: %s)\n", 
				test, 
				t.Format("2006-01-02 15:04:05.999999999 MST"),
				t.UTC().Format("2006-01-02 15:04:05.999999999"))
		}
	}
	
	// Test local timezone handling
	fmt.Printf("\n=== Local Timezone Test ===\n")
	fmt.Printf("Server's local timezone: %s\n", time.Now().Location())
	
	localTestCases := []string{
		"2023-06-15T12:30:45",      // No timezone specified
		"2023-06-15 12:30:45",      // Space separator, no timezone
		"2023-06-15",               // Date only, no timezone
	}
	
	for _, test := range localTestCases {
		t, err := utils.ParseDateTime(test)
		if err != nil {
			fmt.Printf("'%s' -> Error: %v\n", test, err)
		} else {
			fmt.Printf("'%s' -> Parsed as: %s (Location: %s, UTC: %s)\n", 
				test, 
				t.Format("2006-01-02 15:04:05 MST"),
				t.Location(),
				t.UTC().Format("2006-01-02 15:04:05"))
		}
	}
}
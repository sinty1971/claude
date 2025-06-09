package main

import (
	"fmt"
	"penguin-backend/internal/models"
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
	
	for _, test := range testCases {
		instant, err := models.ParseInstant(test)
		if err != nil {
			fmt.Printf("'%s' -> Error: %v\n", test, err)
		} else {
			fmt.Printf("'%s' -> Success: %s\n", test, instant.String())
		}
	}
}
package models

import (
	"testing"
)

func TestIDGeneration(t *testing.T) {
	// Test ID generation from string
	id1 := NewIDFromString("2025-0618 豊田築炉 名和工場")
	id2 := NewIDFromString("2025-0618 豊田築炉 名和工場")

	// Same input should generate same ID
	if id1.Len5() != id2.Len5() {
		t.Errorf("Same input generated different IDs: %s != %s", id1.Len5(), id2.Len5())
	}

	// Test ID length
	if len(id1.Len5()) != 5 {
		t.Errorf("Len5() should return 5 characters, got %d", len(id1.Len5()))
	}

	// Test ID contains only valid characters
	validChars := "123456789ABCDEFGHJKLMNPRSTUVWXYZ"
	for _, char := range id1.Len5() {
		found := false
		for _, valid := range validChars {
			if char == valid {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ID contains invalid character: %c", char)
		}
	}

	t.Logf("Sample ID from folder name: %s", id1.Len5())
}

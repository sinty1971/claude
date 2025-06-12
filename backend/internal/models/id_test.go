package models

import (
	"testing"
	"time"
)

func TestIDGeneration(t *testing.T) {
	// Test ID generation from string
	id1 := NewIDFromString("2025-0618 豊田築炉 名和工場")
	id2 := NewIDFromString("2025-0618 豊田築炉 名和工場")
	
	// Same input should generate same ID
	if id1.Len5() != id2.Len5() {
		t.Errorf("Same input generated different IDs: %s != %s", id1.Len5(), id2.Len5())
	}
	
	// Test ID generation with time
	id3 := NewIDFromFolderWithTime("2025-0618 豊田築炉 名和工場")
	time.Sleep(1 * time.Millisecond)
	id4 := NewIDFromFolderWithTime("2025-0618 豊田築炉 名和工場")
	
	// Same folder at different times should generate different IDs
	if id3.Len5() == id4.Len5() {
		t.Errorf("Same folder at different times generated same IDs: %s == %s", id3.Len5(), id4.Len5())
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
	t.Logf("Sample ID with timestamp: %s", id3.Len5())
}

func TestIDUniqueness(t *testing.T) {
	// Generate multiple IDs to check uniqueness
	ids := make(map[string]bool)
	folders := []string{
		"2025-0618 豊田築炉 名和工場",
		"2025-0619 豊田築炉 別工場",
		"2025-0620 山田工業 本社工場",
		"2025-0621 田中建設 東京支店",
		"2025-0622 佐藤重工 大阪工場",
	}
	
	for _, folder := range folders {
		id := NewIDFromFolderWithTime(folder)
		idStr := id.Len5()
		
		if ids[idStr] {
			t.Errorf("Duplicate ID generated: %s for folder %s", idStr, folder)
		}
		ids[idStr] = true
		
		t.Logf("Folder: %s -> ID: %s", folder, idStr)
	}
}
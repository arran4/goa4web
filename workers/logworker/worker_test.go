package logworker

import (
	"strings"
	"testing"
)

func TestCleanData(t *testing.T) {
	// Create a string longer than 50 chars
	longString := strings.Repeat("a", 100)
	shortString := "Short string"

	data := map[string]any{
		"Body": longString,
		"Other": shortString,
		"Number": 123,
		"Title": "Some Title",
		"UTF8": strings.Repeat("日", 60),
		"Multiline": "Line 1\nLine 2\rLine 3",
	}

	cleaned := cleanData(data)

	cleanedBody, ok := cleaned["Body"].(string)
	if !ok {
		t.Fatal("Body should be a string")
	}

	if len(cleanedBody) >= len(longString) {
		t.Errorf("Expected Body to be truncated, got length %d", len(cleanedBody))
	}

	if !strings.HasSuffix(cleanedBody, "...") {
		t.Error("Expected truncated string to end with ...")
	}

	if cleaned["Other"].(string) != shortString {
		t.Errorf("Expected Other to remain unchanged, got %s", cleaned["Other"])
	}

	if cleaned["Number"].(int) != 123 {
		t.Errorf("Expected Number to remain unchanged, got %d", cleaned["Number"])
	}

	cleanedUTF8, ok := cleaned["UTF8"].(string)
	if !ok {
		t.Fatal("UTF8 should be a string")
	}
	// 50 chars * 3 bytes = 150 bytes. + 3 for "..." = 153.
	// But let's just check suffix and length roughly or rune count
	if !strings.HasSuffix(cleanedUTF8, "...") {
		t.Error("Expected truncated UTF8 string to end with ...")
	}
	// Check rune count of the part before "..."
	runes := []rune(strings.TrimSuffix(cleanedUTF8, "..."))
	if len(runes) != 50 {
		t.Errorf("Expected 50 runes, got %d", len(runes))
	}

	cleanedMultiline, ok := cleaned["Multiline"].(string)
	if !ok {
		t.Fatal("Multiline should be a string")
	}
	if strings.ContainsAny(cleanedMultiline, "\n\r") {
		t.Error("Expected newlines to be removed/replaced")
	}
}

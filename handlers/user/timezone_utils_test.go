package user

import (
	"strings"
	"testing"
)

func TestGetAvailableTimezones(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		tzs := getAvailableTimezones()
		if len(tzs) == 0 {
			t.Fatal("Expected some timezones, got none")
		}

		foundMelbourne := false
		for _, tz := range tzs {
			if tz == "Australia/Melbourne" {
				foundMelbourne = true
				break
			}
		}

		if !foundMelbourne {
			t.Error("Australia/Melbourne not found in available timezones")
		}

		// Ensure no duplicates and sorted
		for i := 1; i < len(tzs); i++ {
			if strings.Compare(tzs[i-1], tzs[i]) >= 0 {
				t.Errorf("Timezones not sorted or contains duplicates: %s >= %s", tzs[i-1], tzs[i])
			}
		}
	})
}

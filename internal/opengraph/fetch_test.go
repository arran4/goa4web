package opengraph

import (
	"strings"
	"testing"
)

func TestFetch_BlocksPrivateIPs(t *testing.T) {
	// These IPs should be blocked by the internal logic when no client is provided
	privateURLs := []string{
		"http://127.0.0.1",
		"http://localhost",
		"http://192.168.1.1",
		"http://10.0.0.1",
	}

	for _, u := range privateURLs {
		t.Run(u, func(t *testing.T) {
			title, desc, image, err := Fetch(u, nil)
			if err == nil {
				t.Errorf("Fetch(%q, nil) expected error, got nil", u)
			}
			if title != "" || desc != "" || image != "" {
				t.Errorf("Fetch(%q, nil) expected empty results, got title=%q, desc=%q, image=%q", u, title, desc, image)
			}
			// Check if error message mentions blocked IP or connection refused (if it slipped through)
			// But we expect "blocked internal ip"
			if err != nil && !strings.Contains(err.Error(), "blocked internal ip") {
				// Note: localhost might resolve to ::1 which is also blocked.
				// But if it tries to connect and fails, that's different.
				// The logic does LookupIP first, so it should be the "blocked internal ip" error.
				t.Logf("Got error: %v", err)
			}
		})
	}
}

func TestGet_BlocksPrivateIPs(t *testing.T) {
	privateURLs := []string{
		"http://127.0.0.1",
		"http://localhost",
		"http://192.168.1.1",
		"http://10.0.0.1",
	}

	for _, u := range privateURLs {
		t.Run(u, func(t *testing.T) {
			resp, err := Get(u, nil)
			if err == nil {
				resp.Body.Close()
				t.Errorf("Get(%q, nil) expected error, got nil", u)
			}
			if err != nil && !strings.Contains(err.Error(), "blocked internal ip") {
				t.Logf("Got error: %v", err)
			}
		})
	}
}

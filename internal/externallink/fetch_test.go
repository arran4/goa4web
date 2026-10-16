package externallink

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/arran4/goa4web/internal/opengraph"
)

func TestFetchMetadataRetriesWithExistingBackoff(t *testing.T) {
	attempts := 0
	var pauses []time.Duration

	info, err := FetchMetadata("https://example.test", func(string) (*opengraph.Info, error) {
		attempts++
		if attempts < 3 {
			return nil, errors.New("temporary failure")
		}
		return &opengraph.Info{Title: "Fetched title"}, nil
	}, func(d time.Duration) {
		pauses = append(pauses, d)
	})
	if err != nil {
		t.Fatalf("FetchMetadata() error = %v", err)
	}
	if info == nil || info.Title != "Fetched title" {
		t.Fatalf("FetchMetadata() info = %#v, want fetched metadata", info)
	}
	if attempts != 3 {
		t.Fatalf("FetchMetadata() attempts = %d, want 3", attempts)
	}
	if want := []time.Duration{2 * time.Second, 4 * time.Second}; !reflect.DeepEqual(pauses, want) {
		t.Fatalf("FetchMetadata() pauses = %v, want %v", pauses, want)
	}
}

func TestFetchMetadataPreservesFinalFailurePause(t *testing.T) {
	var pauses []time.Duration
	wantErr := errors.New("fetch failed")

	info, err := FetchMetadata("https://example.test", func(string) (*opengraph.Info, error) {
		return nil, wantErr
	}, func(d time.Duration) {
		pauses = append(pauses, d)
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("FetchMetadata() error = %v, want %v", err, wantErr)
	}
	if info != nil {
		t.Fatalf("FetchMetadata() info = %#v, want nil", info)
	}
	if want := []time.Duration{2 * time.Second, 4 * time.Second, 6 * time.Second}; !reflect.DeepEqual(pauses, want) {
		t.Fatalf("FetchMetadata() pauses = %v, want %v", pauses, want)
	}
}

package externallink

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/arran4/goa4web/internal/opengraph"
)

func TestFetchMetadataRetriesWithExistingBackoff(t *testing.T) {
	attempts := 0
	var pauses []time.Duration

	info, err := FetchMetadata(context.Background(), "https://example.test", func(context.Context, string) (*opengraph.Info, error) {
		attempts++
		if attempts < 3 {
			return nil, errors.New("temporary failure")
		}
		return &opengraph.Info{Title: "Fetched title"}, nil
	}, func(_ context.Context, d time.Duration) error {
		pauses = append(pauses, d)
		return nil
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

	info, err := FetchMetadata(context.Background(), "https://example.test", func(context.Context, string) (*opengraph.Info, error) {
		return nil, wantErr
	}, func(_ context.Context, d time.Duration) error {
		pauses = append(pauses, d)
		return nil
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

func TestFetchMetadataCancelledBeforeFirstAttempt(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	attempts := 0
	_, err := FetchMetadata(ctx, "https://example.test", func(context.Context, string) (*opengraph.Info, error) {
		attempts++
		return nil, errors.New("unexpected fetch")
	}, Wait)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("FetchMetadata() error = %v, want context.Canceled", err)
	}
	if attempts != 0 {
		t.Fatalf("FetchMetadata() attempts = %d, want 0", attempts)
	}
}

func TestFetchMetadataCancellationStopsRetryDelay(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	attempted := make(chan struct{})
	attempts := 0
	done := make(chan error, 1)
	go func() {
		_, err := FetchMetadata(ctx, "https://example.test", func(context.Context, string) (*opengraph.Info, error) {
			attempts++
			close(attempted)
			return nil, errors.New("temporary failure")
		}, Wait)
		done <- err
	}()
	<-attempted
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("FetchMetadata() error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("FetchMetadata did not stop during retry delay")
	}
	if attempts != 1 {
		t.Fatalf("FetchMetadata() attempts = %d, want 1", attempts)
	}
}

func TestFetchMetadataNoAttemptAfterFetchCancelsContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	attempts := 0
	pauses := 0
	_, err := FetchMetadata(ctx, "https://example.test", func(context.Context, string) (*opengraph.Info, error) {
		attempts++
		cancel()
		return nil, errors.New("temporary failure")
	}, func(context.Context, time.Duration) error {
		pauses++
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("FetchMetadata() error = %v, want context.Canceled", err)
	}
	if attempts != 1 || pauses != 0 {
		t.Fatalf("attempts = %d, pauses = %d; want 1, 0", attempts, pauses)
	}
}

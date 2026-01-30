package admin

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

func TestBuildImageFilesListing_MissingRoot(t *testing.T) {
	// Create a temporary directory for uploads
	tmpDir, err := os.MkdirTemp("", "uploads")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Note: We deliberately do NOT create the 'imagebbs' subdirectory
	// The function expects 'imagebbs' to exist under tmpDir.

	ctx := context.Background()
	// We pass nil for querier because we expect it to fail before using it.
	var queries db.Querier = nil

	// Test requesting the root path
	listing, err := BuildImageFilesListing(ctx, queries, tmpDir, "", "", nil, time.Hour)

	// Expect success (empty list) when root is missing
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}

	if len(listing.Entries) != 0 {
		t.Errorf("Expected empty entries, got %d", len(listing.Entries))
	}
}

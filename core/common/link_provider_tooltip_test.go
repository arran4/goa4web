package common

import (
	"context"
	"database/sql"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	"github.com/stretchr/testify/assert"
)

func TestRenderLink_Tooltips(t *testing.T) {
	mockDB := &MockQuerier{
		Links: map[string]*db.ExternalLink{
			"http://example.com/data": {
				Url:             "http://example.com/data",
				CardTitle:       sql.NullString{String: "DB Title", Valid: true},
				CardDescription: sql.NullString{String: "DB Description", Valid: true},
				CardImage:       sql.NullString{String: "http://example.com/image.jpg", Valid: true},
			},
			"http://example.com/nodata": {
				Url:             "http://example.com/nodata",
				CardTitle:       sql.NullString{String: "", Valid: false},
				CardDescription: sql.NullString{String: "", Valid: false},
			},
			"http://example.com/notitle": {
				Url:             "http://example.com/notitle",
				CardTitle:       sql.NullString{String: "", Valid: true},
				CardDescription: sql.NullString{String: "DB Description Only", Valid: true},
			},
		},
	}

	cd := NewCoreData(context.Background(), mockDB, &config.RuntimeConfig{
		BaseURL: "http://site.local",
	})
	WithLinkSignKey("test-key")(cd)

	provider := NewGoa4WebLinkProvider(cd, context.Background())

	tests := []struct {
		name             string
		rawURL           string
		isBlock          bool
		isImmediateClose bool // true if NO user title (e.g. [url])
		wantTitleAttr    string
		wantText         string // What's inside the <a> tag
	}{
		{
			name:             "Inline + User Title + Data",
			rawURL:           "http://example.com/data",
			isBlock:          false,
			isImmediateClose: false, // User provided title
			wantTitleAttr:    "http://example.com/data - DB Title - DB Description",
		},
		{
			name:             "Inline + No User Title + Data",
			rawURL:           "http://example.com/data",
			isBlock:          false,
			isImmediateClose: true,                        // No user title
			wantTitleAttr:    "http://example.com/data - DB Title - DB Description", // Changed expectation: should show full details in tooltip
			wantText:         "DB Title",
		},
		{
			name:             "Inline + User Title + No DB Title + DB Desc",
			rawURL:           "http://example.com/notitle",
			isBlock:          false,
			isImmediateClose: false,
			wantTitleAttr:    "http://example.com/notitle - DB Description Only", // Should not start with " - "
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOpen, _, hasContent := provider.RenderLink(tt.rawURL, tt.isBlock, tt.isImmediateClose)

			if tt.wantTitleAttr != "" {
				assert.Contains(t, gotOpen, `title="`+tt.wantTitleAttr+`"`)
				// Ensure no leading " - " if title is missing
				if tt.wantTitleAttr == "DB Description Only" {
					assert.NotContains(t, gotOpen, `title=" - `)
				}
			}

			if tt.isImmediateClose && hasContent {
				assert.Contains(t, gotOpen, ">"+tt.wantText+"</a>")
			}
		})
	}
}

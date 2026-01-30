package common

import (
	"context"
	"database/sql"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

type mockQuerier struct {
	db.QuerierStub
	getExternalLink func(ctx context.Context, url string) (*db.ExternalLink, error)
}

func (m *mockQuerier) GetExternalLink(ctx context.Context, url string) (*db.ExternalLink, error) {
	if m.getExternalLink != nil {
		return m.getExternalLink(ctx, url)
	}
	return nil, sql.ErrNoRows
}

func TestGoa4WebLinkProvider_RenderLink(t *testing.T) {
	ctx := context.Background()
	mq := &mockQuerier{
		getExternalLink: func(ctx context.Context, url string) (*db.ExternalLink, error) {
			if url == "https://example.com" {
				return &db.ExternalLink{
					Url:             url,
					CardTitle:       sql.NullString{String: "Example Domain", Valid: true},
					CardDescription: sql.NullString{String: "This domain is for use in illustrative examples.", Valid: true},
					CardImage:       sql.NullString{String: "https://example.com/image.png", Valid: true},
				}, nil
			}
			if url == "https://empty-title.com" {
				return &db.ExternalLink{
					Url:             url,
					CardTitle:       sql.NullString{String: "", Valid: true},
					CardDescription: sql.NullString{String: "Description Only", Valid: true},
				}, nil
			}
			return nil, sql.ErrNoRows
		},
	}

	cd := NewCoreData(ctx, mq, nil)
	provider := NewGoa4WebLinkProvider(cd, ctx)

	tests := []struct {
		name             string
		url              string
		isBlock          bool
		isImmediateClose bool
		wantOpen         string
		wantClose        string
		wantConsume      bool
	}{
		{
			name:             "Inline, User Title (ImmediateClose=false), With Data",
			url:              "https://example.com",
			isBlock:          false,
			isImmediateClose: false,
			// Current logic: title="Title - Description"
			wantOpen:    `<a href="https://example.com" target="_blank" rel="noopener noreferrer" title="Example Domain - This domain is for use in illustrative examples.">`,
			wantClose:   "</a>",
			wantConsume: false,
		},
		{
			name:             "Inline, No User Title (ImmediateClose=true), With Data",
			url:              "https://example.com",
			isBlock:          false,
			isImmediateClose: true,
			// New requirement: title="Title - Description" (was just Description)
			wantOpen:    `<a href="https://example.com" target="_blank" rel="noopener noreferrer" title="Example Domain - This domain is for use in illustrative examples.">Example Domain</a>`,
			wantClose:   "",
			wantConsume: true,
		},
		{
			name:             "Inline, No User Title, Empty Card Title",
			url:              "https://empty-title.com",
			isBlock:          false,
			isImmediateClose: true,
			// Fallback text is description. Tooltip is description (Title is empty).
			wantOpen:    `<a href="https://empty-title.com" target="_blank" rel="noopener noreferrer" title="Description Only">Description Only</a>`,
			wantClose:   "",
			wantConsume: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOpen, gotClose, gotConsume := provider.RenderLink(tt.url, tt.isBlock, tt.isImmediateClose)
			if gotOpen != tt.wantOpen {
				t.Errorf("RenderLink() open\n got:  %v\n want: %v", gotOpen, tt.wantOpen)
			}
			if gotClose != tt.wantClose {
				t.Errorf("RenderLink() close = %v, want %v", gotClose, tt.wantClose)
			}
			if gotConsume != tt.wantConsume {
				t.Errorf("RenderLink() consume = %v, want %v", gotConsume, tt.wantConsume)
			}
		})
	}
}

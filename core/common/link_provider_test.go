package common

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	"github.com/stretchr/testify/assert"
)

type MockQuerier struct {
	db.QuerierStub
	Links map[string]*db.ExternalLink
}

func (m *MockQuerier) GetExternalLink(ctx context.Context, url string) (*db.ExternalLink, error) {
	if l, ok := m.Links[url]; ok {
		return l, nil
	}
	return nil, sql.ErrNoRows
}

func TestRenderLink_RoutesThroughGoto(t *testing.T) {
	mockDB := &MockQuerier{
		Links: map[string]*db.ExternalLink{
			"http://example.com/card": {
				Url:             "http://example.com/card",
				CardTitle:       sql.NullString{String: "Card Title", Valid: true},
				CardDescription: sql.NullString{String: "Card Desc", Valid: true},
				CardImage:       sql.NullString{String: "http://example.com/image.jpg", Valid: true},
			},
		},
	}

	cd := NewCoreData(context.Background(), mockDB, &config.RuntimeConfig{
		BaseURL: "http://site.local",
	})
	// CoreData requires LinkSignKey for signing
	WithLinkSignKey("test-key")(cd)

	provider := NewGoa4WebLinkProvider(cd, context.Background())

	tests := []struct {
		name             string
		rawURL           string
		isBlock          bool
		isImmediateClose bool // true if NO title provided (e.g. [url])
		wantContains     string
		wantNotContains  string
	}{
		{
			name:             "Inline HTTP link (no title)",
			rawURL:           "http://example.com",
			isBlock:          false,
			isImmediateClose: true,
			wantContains:     "http://site.local/goto?u=http%3A%2F%2Fexample.com&sig=",
		},
		{
			name:             "Inline HTTPS link (with title)",
			rawURL:           "https://example.com",
			isBlock:          false,
			isImmediateClose: false,
			wantContains:     "http://site.local/goto?u=https%3A%2F%2Fexample.com&sig=",
		},
		{
			name:             "Block HTTP link (no title, no card data)",
			rawURL:           "http://example.com/nocard",
			isBlock:          true,
			isImmediateClose: true,
			wantContains:     "http://site.local/goto?u=http%3A%2F%2Fexample.com%2Fnocard&sig=",
		},
		{
			name:             "Card link (Block + No Title + Data)",
			rawURL:           "http://example.com/card",
			isBlock:          true,
			isImmediateClose: true,
			wantContains:     `href="http://example.com/card"`, // Should stay direct (uses sanitized rawURL)
			// The rule says "All links that don't have a card ... should be routed through goto".
			// So links WITH a card stay direct.
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOpen, gotClose, _ := provider.RenderLink(tt.rawURL, tt.isBlock, tt.isImmediateClose)
			full := gotOpen + gotClose
			if tt.wantContains != "" {
				assert.Contains(t, full, tt.wantContains)
			}
			if tt.wantNotContains != "" {
				assert.NotContains(t, full, tt.wantNotContains)
			}

            // Also check that it is properly signed if it contains goto
            if strings.Contains(full, "/goto?") {
                assert.Contains(t, full, "&sig=")
            }
		})
	}
}

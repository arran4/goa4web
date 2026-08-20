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
				CardImage:       sql.NullString{String: "http://example.com/image.jpg?sig=asset_sig_123", Valid: true},
				FaviconCache:    sql.NullString{String: "favicon.ico", Valid: true},
			},
			"https://www.bloomberg.com/news/sample?accessToken=token123": {
				Url:             "https://www.bloomberg.com/news/sample?accessToken=token123",
				CardTitle:       sql.NullString{String: "Bloomberg Article", Valid: true},
				CardDescription: sql.NullString{String: "Market Analysis", Valid: true},
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
		isImmediateClose bool
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
			wantContains:     "http://site.local/goto?u=http%3A%2F%2Fexample.com%2Fcard&sig=",
		},
		{
			name:             "Card link with tracking param strips tracking from goto, tooltip, and matches canonical in db",
			rawURL:           "http://example.com/card?utm_source=twitter&utm_medium=social",
			isBlock:          true,
			isImmediateClose: true,
			wantContains:     "http://site.local/goto?u=http%3A%2F%2Fexample.com%2Fcard&sig=",
			wantNotContains:  "utm_source",
		},
		{
			name:             "Bloomberg accessToken retained in goto, tooltip, and db lookup",
			rawURL:           "https://www.bloomberg.com/news/sample?accessToken=token123&utm_source=email",
			isBlock:          true,
			isImmediateClose: true,
			wantContains:     "accessToken=token123",
			wantNotContains:  "utm_source",
		},
		{
			name:             "AWS Presigned URL with X-Amz-Signature is not stripped or rewritten",
			rawURL:           "https://s3.amazonaws.com/bucket/doc.pdf?X-Amz-Signature=sig123&utm_source=tracked",
			isBlock:          false,
			isImmediateClose: true,
			wantContains:     "X-Amz-Signature=sig123",
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

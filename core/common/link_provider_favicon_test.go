package common

import (
	"context"
	"database/sql"
	"html"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	"github.com/stretchr/testify/assert"
)

func TestRenderLink_Favicon(t *testing.T) {
	mockDB := &MockQuerier{
		Links: map[string]*db.ExternalLink{
			"http://example.com/favicon": {
				Url:          "http://example.com/favicon",
				CardTitle:    sql.NullString{String: "Favicon Title", Valid: true},
				FaviconCache: sql.NullString{String: "cache:favicon.ico", Valid: true},
			},
			"http://example.com/nofavicon": {
				Url:       "http://example.com/nofavicon",
				CardTitle: sql.NullString{String: "No Favicon Title", Valid: true},
			},
		},
	}

	cd := NewCoreData(context.Background(), mockDB, &config.RuntimeConfig{
		BaseURL: "http://site.local",
	})
	WithLinkSignKey("test-key")(cd)
	WithImageSignKey("img-key")(cd)

	provider := NewGoa4WebLinkProvider(cd, context.Background())

	tests := []struct {
		name             string
		rawURL           string
		isBlock          bool
		isImmediateClose bool // true if [link] (no title), false if [link=...]...
		wantFavicon      bool
		wantTooltip      string
	}{
		{
			name:             "Inline with favicon (no user title)",
			rawURL:           "http://example.com/favicon",
			isBlock:          false,
			isImmediateClose: true,
			wantFavicon:      true,
			wantTooltip:      "http://example.com/favicon - Favicon Title",
		},
		{
			name:             "Inline with favicon (with user title)",
			rawURL:           "http://example.com/favicon",
			isBlock:          false,
			isImmediateClose: false,
			wantFavicon:      true,
			wantTooltip:      "http://example.com/favicon - Favicon Title",
		},
		{
			name:             "Inline without favicon",
			rawURL:           "http://example.com/nofavicon",
			isBlock:          false,
			isImmediateClose: true,
			wantFavicon:      false,
			wantTooltip:      "http://example.com/nofavicon - No Favicon Title",
		},
		{
			name:             "Inline no data",
			rawURL:           "http://example.com/nodata",
			isBlock:          false,
			isImmediateClose: true,
			wantFavicon:      false,
			wantTooltip:      "http://example.com/nodata",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOpen, gotClose, _ := provider.RenderLink(tt.rawURL, tt.isBlock, tt.isImmediateClose)
			full := gotOpen + gotClose

			// Check target is goto
			assert.Contains(t, gotOpen, "/goto?u=")

			// Check tooltip
			assert.Contains(t, gotOpen, `title="`+html.EscapeString(tt.wantTooltip)+`"`)

			if tt.wantFavicon {
				assert.Contains(t, full, `class="a4code-inline-favicon"`)
				assert.Contains(t, full, `<img src="`)
			} else {
				assert.NotContains(t, full, `class="a4code-inline-favicon"`)
			}
		})
	}
}

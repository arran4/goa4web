package handlers

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	db "github.com/arran4/goa4web/internal/db"
)

func TestGetPageSize(t *testing.T) {
	orig := config.AppRuntimeConfig
	defer func() { config.AppRuntimeConfig = orig }()

	tests := []struct {
		name string
		pref *db.Preference
		want int
	}{
		{"default", nil, 15},
		{"stored pref", &db.Preference{PageSize: 20}, 20},
		{"below min", &db.Preference{PageSize: 2}, 5},
		{"above max", &db.Preference{PageSize: 60}, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.AppRuntimeConfig.PageSizeMin = 5
			config.AppRuntimeConfig.PageSizeMax = 50
			config.AppRuntimeConfig.PageSizeDefault = 15

			r := httptest.NewRequest("GET", "/", nil)
			ctx := r.Context()
			if tt.pref != nil {
				ctx = context.WithValue(ctx, ContextKey("preference"), tt.pref)
			}
			r = r.WithContext(ctx)

			got := GetPageSize(r)
			if got != tt.want {
				t.Fatalf("want %d got %d", tt.want, got)
			}
		})
	}
}

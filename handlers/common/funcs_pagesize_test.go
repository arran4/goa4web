package common

import (
	"context"
	"net/http/httptest"
	"testing"

	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/runtimeconfig"
)

func TestGetPageSize(t *testing.T) {
	orig := runtimeconfig.AppRuntimeConfig
	defer func() { runtimeconfig.AppRuntimeConfig = orig }()

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
			runtimeconfig.AppRuntimeConfig.PageSizeMin = 5
			runtimeconfig.AppRuntimeConfig.PageSizeMax = 50
			runtimeconfig.AppRuntimeConfig.PageSizeDefault = 15

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

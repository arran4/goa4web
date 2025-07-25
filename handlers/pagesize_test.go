package handlers

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestGetPageSize(t *testing.T) {
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
			cfg := config.AppRuntimeConfig
			cfg.PageSizeMin = 5
			cfg.PageSizeMax = 50
			cfg.PageSizeDefault = 15

			cd := common.NewCoreData(context.Background(), nil, common.WithConfig(cfg))
			if tt.pref != nil {
				common.WithPreference(tt.pref)(cd)
			}
			r := httptest.NewRequest("GET", "/", nil)
			r = r.WithContext(context.WithValue(r.Context(), consts.KeyCoreData, cd))

			got := GetPageSize(r)
			if got != tt.want {
				t.Fatalf("want %d got %d", tt.want, got)
			}
		})
	}
}

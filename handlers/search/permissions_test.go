package search

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

func TestCanSearch(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	cases := []struct {
		name string
		cd   *common.CoreData
		want bool
	}{
		{
			name: "no grants",
			cd: func() *common.CoreData {
				cd := common.NewCoreData(context.Background(), nil, cfg)
				cd.UserID = 1
				return cd
			}(),
			want: false,
		},
		{
			name: "global grant",
			cd: func() *common.CoreData {
				q := &db.QuerierStub{
					GetAdministratorUserRoleReturns: &db.UserRole{},
				}
				cd := common.NewCoreData(context.Background(), q, cfg, common.WithUserRoles([]string{"administrator"}))
				cd.UserID = 1
				cd.AdminMode = true
				return cd
			}(),
			want: true,
		},
		{
			name: "section grant",
			cd: func() *common.CoreData {
				q := &db.QuerierStub{
					GetAdministratorUserRoleReturns: &db.UserRole{},
				}
				cd := common.NewCoreData(context.Background(), q, cfg, common.WithUserRoles([]string{"administrator"}))
				cd.UserID = 1
				cd.AdminMode = true
				return cd
			}(),
			want: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := common.CanSearch(tc.cd, "news"); got != tc.want {
				t.Fatalf("CanSearch() = %v, want %v", got, tc.want)
			}
		})
	}
}

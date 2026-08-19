package blogs

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"context"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/config"
)

type MyQuerierStub struct {
	db.QuerierStub
}

func (m *MyQuerierStub) AdminListUserIDsByRole(ctx context.Context, role string) ([]int32, error) {
	return []int32{1, 2, 3}, nil
}

func BenchmarkAdminPage(b *testing.B) {
	// Setup mock data
	var roles []*db.Role
	for i := 0; i < 5; i++ {
		roles = append(roles, &db.Role{ID: int32(i), Name: "Role" + string(rune(i))})
	}
	var grants []*db.Grant
	for i := 0; i < 1000; i++ {
		grants = append(grants, &db.Grant{
			Section: "blogs",
			UserID: sql.NullInt32{Int32: int32(i), Valid: true},
		})
	}

	// Create mock querier
	baseQ := db.QuerierStub{
		AdminListRolesFn: func(ctx context.Context) ([]*db.Role, error) {
			return roles, nil
		},
		ListGrantsFn: func(ctx context.Context) ([]*db.Grant, error) {
			return grants, nil
		},
		AdminListGrantsByRoleIDFn: func(ctx context.Context, roleID sql.NullInt32) ([]*db.Grant, error) {
			return nil, nil
		},
		SystemGetUserByIDFn: func(ctx context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
			return &db.SystemGetUserByIDRow{
				Idusers: id,
				Username: sql.NullString{String: "User", Valid: true},
				Email: sql.NullString{String: "user@example.com", Valid: true},
			}, nil
		},
		SystemGetUsersByIDsFn: func(ctx context.Context, ids []int32) ([]*db.SystemGetUsersByIDsRow, error) {
			var rows []*db.SystemGetUsersByIDsRow
			for _, id := range ids {
				rows = append(rows, &db.SystemGetUsersByIDsRow{
					Idusers: id,
					Username: sql.NullString{String: "User", Valid: true},
					Email: sql.NullString{String: "user@example.com", Valid: true},
				})
			}
			return rows, nil
		},
	}
	q := &MyQuerierStub{QuerierStub: baseQ}

	cd := common.NewCoreData(context.Background(), q, &config.RuntimeConfig{})

	req, _ := http.NewRequest("GET", "/admin/blogs", nil)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		AdminPage(w, req)
	}
}

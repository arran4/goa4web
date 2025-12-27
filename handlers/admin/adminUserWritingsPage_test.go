package admin

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

type userWritingsQueries struct {
	db.Querier
	userID   int32
	user     *db.SystemGetUserByIDRow
	writings []*db.AdminGetAllWritingsByAuthorRow
}

func (q *userWritingsQueries) SystemGetUserByID(_ context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
	if id != q.userID {
		return nil, fmt.Errorf("unexpected user id: %d", id)
	}
	return q.user, nil
}

func (q *userWritingsQueries) AdminGetAllWritingsByAuthor(_ context.Context, id int32) ([]*db.AdminGetAllWritingsByAuthorRow, error) {
	if id != q.userID {
		return nil, fmt.Errorf("unexpected author id: %d", id)
	}
	return q.writings, nil
}

func (q *userWritingsQueries) SystemCheckGrant(_ context.Context, arg db.SystemCheckGrantParams) (int32, error) {
	if arg.Section == common.AdminAccessSection && arg.Action == common.AdminAccessAction {
		return 1, nil
	}
	return 0, fmt.Errorf("unexpected grant check: %#v", arg)
}

func TestAdminUserWritingsPage(t *testing.T) {
	queries := &userWritingsQueries{
		userID: 1,
		user: &db.SystemGetUserByIDRow{
			Idusers:                1,
			Email:                  sql.NullString{String: "u@test", Valid: true},
			Username:               sql.NullString{String: "user", Valid: true},
			PublicProfileEnabledAt: sql.NullTime{},
		},
		writings: []*db.AdminGetAllWritingsByAuthorRow{{
			Idwriting:         1,
			UsersIdusers:      1,
			ForumthreadID:     0,
			LanguageID:        sql.NullInt32{},
			WritingCategoryID: 2,
			Title:             sql.NullString{String: "Title", Valid: true},
			Published:         sql.NullTime{Time: time.Now(), Valid: true},
			Timezone:          sql.NullString{String: time.Local.String(), Valid: true},
			Writing:           sql.NullString{String: "", Valid: true},
			Abstract:          sql.NullString{String: "", Valid: true},
			Private:           sql.NullBool{Bool: false, Valid: true},
			DeletedAt:         sql.NullTime{},
			LastIndex:         sql.NullTime{},
			Username:          sql.NullString{String: "user", Valid: true},
			Comments:          0,
		}},
	}

	req := httptest.NewRequest("GET", "/admin/user/1/writings", nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, queries, config.NewRuntimeConfig(), common.WithUserRoles([]string{}))
	cd.SetCurrentProfileUserID(1)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	adminUserWritingsPage(rr, req)

	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
	body := rr.Body.String()
	if !strings.Contains(body, `<td>1</td>`) {
		t.Fatalf("missing id: %s", body)
	}
	if !strings.Contains(body, "Title") {
		t.Fatalf("missing title: %s", body)
	}
}

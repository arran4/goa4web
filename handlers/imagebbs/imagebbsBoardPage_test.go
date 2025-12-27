package imagebbs

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestBoardPageRendersSubBoards(t *testing.T) {
	boardID := int32(3)
	q := &db.QuerierStub{
		ListBoardsByParentIDForListerReturn: map[sql.NullInt32][]*db.Imageboard{
			{Int32: boardID, Valid: true}: {
				{
					Idimageboard:           4,
					ImageboardIdimageboard: sql.NullInt32{Int32: boardID, Valid: true},
					Title:                  sql.NullString{String: "child", Valid: true},
					Description:            sql.NullString{String: "sub", Valid: true},
				},
			},
		},
		ListImagePostsByBoardForListerReturn: map[sql.NullInt32][]*db.ListImagePostsByBoardForListerRow{
			{Int32: boardID, Valid: true}: {
				{
					Idimagepost:            1,
					ForumthreadID:          1,
					UsersIdusers:           1,
					ImageboardIdimageboard: sql.NullInt32{Int32: boardID, Valid: true},
					Posted:                 sql.NullTime{Time: time.Unix(0, 0), Valid: true},
					Timezone:               sql.NullString{String: time.Local.String(), Valid: true},
					Description:            sql.NullString{String: "desc", Valid: true},
					Thumbnail:              sql.NullString{String: "/t", Valid: true},
					Fullimage:              sql.NullString{String: "/f", Valid: true},
					FileSize:               10,
					Approved:               true,
					Username:               sql.NullString{String: "alice", Valid: true},
					Comments:               sql.NullInt32{Int32: 0, Valid: true},
				},
			},
		},
	}

	req := httptest.NewRequest("GET", "/imagebbs/board/3", nil)
	req = mux.SetURLVars(req, map[string]string{"board": "3"})
	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}))
	cd.AdminMode = true
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	ImagebbsBoardPage(rr, req)

	if got := len(q.ListBoardsByParentIDForListerCalls); got != 1 {
		t.Fatalf("expected one sub-board query, got %d", got)
	}
	if got := len(q.ListImagePostsByBoardForListerCalls); got != 1 {
		t.Fatalf("expected one post listing query, got %d", got)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Sub-Boards") {
		t.Fatalf("expected sub boards in output: %s", body)
	}
	if !strings.Contains(body, "Pictures:") {
		t.Fatalf("expected pictures in output: %s", body)
	}
}

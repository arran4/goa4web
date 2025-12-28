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
	qs := &db.QuerierStub{}

	req := httptest.NewRequest("GET", "/imagebbs/board/3", nil)
	req = mux.SetURLVars(req, map[string]string{"board": "3"})

	// Create CoreData with the QuerierStub
	cd := common.NewCoreData(req.Context(), qs, config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}))
	cd.AdminMode = true

	// Wrap with FakeCoreData
	fakeCD := NewFakeCoreData(cd)

	// Prepare test data
	subBoards := []*db.Imageboard{
		{
			Idimageboard:     4,
			ImageboardIdimageboard: sql.NullInt32{Int32: 3, Valid: true},
			Title:            sql.NullString{String: "child", Valid: true},
			Description:      sql.NullString{String: "sub", Valid: true},
			ApprovalRequired: false,
		},
	}

	posts := []*db.ListImagePostsByBoardForListerRow{
		{
			Idimagepost:           1,
			ForumthreadID:         1,
			UsersIdusers:          1,
			ImageboardIdimageboard: sql.NullInt32{Int32: 3, Valid: true},
			Posted:                sql.NullTime{Time: time.Unix(0, 0), Valid: true},
			Timezone:              sql.NullString{String: time.Local.String(), Valid: true},
			Description:           sql.NullString{String: "desc", Valid: true},
			Thumbnail:             sql.NullString{String: "/t", Valid: true},
			Fullimage:             sql.NullString{String: "/f", Valid: true},
			FileSize:              10,
			Approved:              true,
			Username:              sql.NullString{String: "alice", Valid: true},
			Comments:              sql.NullInt32{Int32: 0, Valid: true},
		},
	}

	// Stub the fetches
	fakeCD.StubSubImageBoards(3, subBoards)
	fakeCD.StubImageBoardPosts(3, posts)
    fakeCD.StubSystemCheckGrant(1)

	// Inject CoreData into context
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	ImagebbsBoardPage(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, "Sub-Boards") {
		t.Fatalf("expected sub boards in output: %s", body)
	}
	if !strings.Contains(body, "Pictures:") {
		t.Fatalf("expected pictures in output: %s", body)
	}
	if !strings.Contains(body, "child") {
		t.Errorf("expected sub-board title 'child' in output")
	}
	if !strings.Contains(body, "desc") {
		t.Errorf("expected post description 'desc' in output")
	}
}

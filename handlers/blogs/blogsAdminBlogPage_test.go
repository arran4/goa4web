package blogs

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestAdminBlogPage_UsesURLParam(t *testing.T) {
	blogID := int32(7)
	now := time.Now()
	queries := testhelpers.NewQuerierStub()
	queries.GetBlogEntryForListerByIDRow = &db.GetBlogEntryForListerByIDRow{
		Idblogs:       blogID,
		ForumthreadID: sql.NullInt32{},
		UsersIdusers:  1,
		LanguageID:    sql.NullInt32{Int32: 1, Valid: true},
		Blog:          sql.NullString{String: "body", Valid: true},
		Written:       now,
		Timezone:      sql.NullString{String: time.Local.String(), Valid: true},
		Username:      sql.NullString{String: "user", Valid: true},
		Comments:      0,
		IsOwner:       true,
		Title:         "body",
	}
	queries.AdminListRolesReturns = []*db.Role{
		{ID: 42, Name: "editor"},
	}
	queries.ListGrantsReturns = []*db.Grant{
		{
			Section: "blogs",
			Item:    sql.NullString{String: "entry", Valid: true},
			ItemID:  sql.NullInt32{Int32: blogID, Valid: true},
			RoleID:  sql.NullInt32{Int32: 42, Valid: true},
		},
	}

	req := httptest.NewRequest("GET", "/admin/blogs/blog/"+strconv.Itoa(int(blogID)), nil)
	req = mux.SetURLVars(req, map[string]string{"blog": strconv.Itoa(int(blogID))})
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(req.Context(), queries, cfg)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	rr := httptest.NewRecorder()

	AdminBlogPage(rr, req.WithContext(ctx))
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	if len(queries.GetBlogEntryForListerByIDCalls) != 1 {
		t.Fatalf("expected blog lookup once, got %d calls", len(queries.GetBlogEntryForListerByIDCalls))
	}
	if got := queries.GetBlogEntryForListerByIDCalls[0].ID; got != blogID {
		t.Fatalf("expected blog ID %d, got %d", blogID, got)
	}
}

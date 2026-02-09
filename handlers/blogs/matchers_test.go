package blogs

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

type mockQuerier struct {
	*db.QuerierStub
	grants map[string]bool
}

func (m *mockQuerier) SystemCheckGrant(ctx context.Context, arg db.SystemCheckGrantParams) (int32, error) {
	// Construct key similar to test setup: section:item:action:id
	// arg.Item is NullString, arg.ItemID is NullInt32
	item := ""
	if arg.Item.Valid {
		item = arg.Item.String
	}
	id := int32(0)
	if arg.ItemID.Valid {
		id = arg.ItemID.Int32
	}

	key := fmt.Sprintf("%s:%s:%s:%d", arg.Section, item, arg.Action, id)
	if m.grants[key] {
		return 1, nil
	}
	return 0, sql.ErrNoRows
}

func TestRequireBlogCommentAccess(t *testing.T) {
	blogID := 123
	tests := []struct {
		name           string
		grants         []string
		canReplyThread bool
		expectedStatus int
		blogExists     bool
	}{
		{
			name:           "Allowed via View Grant",
			grants:         []string{"blogs:entry:view:123"},
			expectedStatus: http.StatusOK,
			blogExists:     true,
		},
		{
			name:           "Allowed via Reply Grant",
			grants:         []string{"blogs:entry:reply:123"},
			expectedStatus: http.StatusOK,
			blogExists:     true,
		},
		{
			name:           "Allowed via Thread Reply",
			grants:         []string{},
			canReplyThread: true,
			expectedStatus: http.StatusOK,
			blogExists:     true,
		},
		{
			name:           "Denied",
			grants:         []string{},
			expectedStatus: http.StatusForbidden,
			blogExists:     true,
		},
		{
			name:           "Blog Not Found",
			grants:         []string{"blogs:entry:view:123"},
			expectedStatus: http.StatusForbidden,
			blogExists:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qStub := testhelpers.NewQuerierStub()
			if tt.blogExists {
				qStub.GetBlogEntryForListerByIDRow = &db.GetBlogEntryForListerByIDRow{
					Idblogs:       int32(blogID),
					ForumthreadID: sql.NullInt32{Int32: 1, Valid: true},
					UsersIdusers:  1,
				}
			} else {
				qStub.GetBlogEntryForListerByIDErr = sql.ErrNoRows
			}

			qStub.GetThreadLastPosterAndPermsReturns = &db.GetThreadLastPosterAndPermsRow{
				Idforumthread: 1,
			}

			if tt.canReplyThread {
				qStub.GetThreadBySectionThreadIDForReplierReturn = &db.Forumthread{
					Idforumthread: 1,
				}
			}

			grantMap := make(map[string]bool)
			for _, g := range tt.grants {
				grantMap[g] = true
			}

			q := &mockQuerier{
				QuerierStub: qStub,
				grants:      grantMap,
			}

			req := httptest.NewRequest("GET", "/blog/"+strconv.Itoa(blogID)+"/comments", nil)
			req = mux.SetURLVars(req, map[string]string{"blog": strconv.Itoa(blogID)})
			cfg := config.NewRuntimeConfig()
			cd := common.NewCoreData(req.Context(), q, cfg)

			ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
			rr := httptest.NewRecorder()

			handler := RequireBlogCommentAccess(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(rr, req.WithContext(ctx))

			if rr.Code != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.expectedStatus)
			}
		})
	}
}

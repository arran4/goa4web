package blogs

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestRequireBlogCommentAccess(t *testing.T) {
	tests := []struct {
		name           string
		grants         map[string]bool
		blogEntry      *db.GetBlogEntryForListerByIDRow
		blogEntryErr   error
		thread         *db.GetThreadLastPosterAndPermsRow
		threadErr      error
		expectedStatus int
	}{
		{
			name: "View Grant",
			grants: map[string]bool{
				testhelpers.GrantKey("blogs", "entry", "view"): true,
			},
			blogEntry:      &db.GetBlogEntryForListerByIDRow{Idblogs: 1, UsersIdusers: 1},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Reply Grant",
			grants: map[string]bool{
				testhelpers.GrantKey("blogs", "entry", "reply"): true,
			},
			blogEntry:      &db.GetBlogEntryForListerByIDRow{Idblogs: 1, UsersIdusers: 1},
			expectedStatus: http.StatusOK,
		},
		{
			name: "SelectedThreadCanReply",
			grants:         map[string]bool{},
			blogEntry:      &db.GetBlogEntryForListerByIDRow{Idblogs: 1, UsersIdusers: 1, ForumthreadID: sql.NullInt32{Int32: 10, Valid: true}},
			thread:         &db.GetThreadLastPosterAndPermsRow{Idforumthread: 10},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "No Access",
			grants:         map[string]bool{},
			blogEntry:      &db.GetBlogEntryForListerByIDRow{Idblogs: 1, UsersIdusers: 1},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Blog Not Found",
			blogEntryErr:   sql.ErrNoRows,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := testhelpers.NewQuerierStub(
				testhelpers.WithGrants(tt.grants),
			)
			stub.GetBlogEntryForListerByIDRow = tt.blogEntry
			stub.GetBlogEntryForListerByIDErr = tt.blogEntryErr

            if tt.thread != nil {
                stub.GetThreadLastPosterAndPermsReturns = tt.thread
            } else if tt.threadErr != nil {
                stub.GetThreadLastPosterAndPermsErr = tt.threadErr
            }

            if tt.name == "SelectedThreadCanReply" {
                 stub.GetThreadBySectionThreadIDForReplierReturn = &db.Forumthread{
                     Idforumthread: 10,
                 }
            }

			ctx := context.Background()
			cd := common.NewCoreData(ctx, stub, config.NewRuntimeConfig())
			cd.UserID = 1

			r := httptest.NewRequest("GET", "/blog/1/comments", nil)
			r = mux.SetURLVars(r, map[string]string{"blog": "1"})
			r = r.WithContext(context.WithValue(r.Context(), consts.KeyCoreData, cd))

			w := httptest.NewRecorder()

			handler := RequireBlogCommentAccess(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(w, r)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

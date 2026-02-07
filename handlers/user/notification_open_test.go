package user

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

func TestUserNotificationOpenPage_SetsTitle(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		tests := []struct {
			name          string
			notifID       string
			expectedTitle string
		}{
			{
				name:          "No Link",
				notifID:       "123",
				expectedTitle: "Notification",
			},
			{
				name:          "With Link Public",
				notifID:       "124",
				expectedTitle: "Notification: Thread Title 1",
			},
			{
				name:          "With Link Private",
				notifID:       "125",
				expectedTitle: "Notification: Thread Title 2",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				qs := testhelpers.NewQuerierStub()

				qs.GetNotificationForListerFn = func(ctx context.Context, arg db.GetNotificationForListerParams) (*db.Notification, error) {
					link := ""
					if arg.ID == 124 {
						link = "/topic/1/thread/1"
					} else if arg.ID == 125 {
						link = "/private/topic/2/thread/2"
					}
					return &db.Notification{
						ID:      arg.ID,
						Message: sql.NullString{String: "Test Notification Message", Valid: true},
						Link:    sql.NullString{String: link, Valid: link != ""},
					}, nil
				}

				qs.GetForumTopicByIdForUserFn = func(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
					handler := "forum"
					if arg.Idforumtopic == 2 {
						handler = "private"
					}
					return &db.GetForumTopicByIdForUserRow{
						Idforumtopic: arg.Idforumtopic,
						Handler:      handler,
					}, nil
				}
				qs.GetForumTopicByIdFn = func(ctx context.Context, id int32) (*db.Forumtopic, error) {
					handler := "forum"
					if id == 2 {
						handler = "private"
					}
					return &db.Forumtopic{Idforumtopic: id, Handler: handler}, nil
				}

				qs.GetThreadLastPosterAndPermsFn = func(ctx context.Context, arg db.GetThreadLastPosterAndPermsParams) (*db.GetThreadLastPosterAndPermsRow, error) {
					return &db.GetThreadLastPosterAndPermsRow{
						Idforumthread: arg.ThreadID,
						Firstpost:     100 + arg.ThreadID,
					}, nil
				}

				qs.GetCommentByIdForUserFn = func(ctx context.Context, arg db.GetCommentByIdForUserParams) (*db.GetCommentByIdForUserRow, error) {
					text := "Thread Title 1"
					if arg.ID == 102 {
						text = "Thread Title 2"
					}
					return &db.GetCommentByIdForUserRow{
						Idcomments: arg.ID,
						Text:       sql.NullString{String: text, Valid: true},
					}, nil
				}

				req := httptest.NewRequest("GET", "/usr/notifications/open/"+tt.notifID, nil)
				req = mux.SetURLVars(req, map[string]string{"id": tt.notifID})

				// Setup session
				store := sessions.NewCookieStore([]byte("secret"))
				core.Store = store
				session, _ := store.New(req, "session")
				session.Values["UID"] = int32(1)
				ctx := context.WithValue(req.Context(), core.ContextValues("session"), session)

				// Setup CoreData
				cd := common.NewCoreData(ctx, qs, config.NewRuntimeConfig())
				cd.Config.NotificationsEnabled = true
				cd.UserID = 1
				ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
				req = req.WithContext(ctx)

				rr := httptest.NewRecorder()

				userNotificationOpenPage(rr, req)

				if cd.PageTitle != tt.expectedTitle {
					t.Errorf("Expected PageTitle to be '%s', got '%s'", tt.expectedTitle, cd.PageTitle)
				}
			})
		}
	})
}

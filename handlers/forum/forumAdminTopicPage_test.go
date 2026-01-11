package forum

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
)

func TestAdminTopicPage(t *testing.T) {
	topicID := 4
	queries := &db.QuerierStub{
		GetForumTopicByIdReturns: &db.Forumtopic{
			Idforumtopic:                 int32(topicID),
			ForumcategoryIdforumcategory: 1,
			Title:                        sql.NullString{String: "t", Valid: true},
			Description:                  sql.NullString{String: "d", Valid: true},
			Threads:                      sql.NullInt32{Int32: 2, Valid: true},
			Comments:                     sql.NullInt32{Int32: 3, Valid: true},
			Lastaddition:                 sql.NullTime{Time: time.Now(), Valid: true},
			Handler:                      "",
		},
		AdminListForumTopicGrantsByTopicIDReturns: []*db.AdminListForumTopicGrantsByTopicIDRow{
			{ID: 1, Section: "forum", Action: "see", RoleName: sql.NullString{String: "Anyone", Valid: true}},
		},
	}

	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
	r := httptest.NewRequest("GET", "/admin/forum/topics/topic/"+strconv.Itoa(topicID), nil)
	r = mux.SetURLVars(r, map[string]string{"topic": strconv.Itoa(topicID)})
	r = r.WithContext(context.WithValue(r.Context(), consts.KeyCoreData, cd))
	w := httptest.NewRecorder()
	AdminTopicPage(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d", w.Code, http.StatusOK)
	}
}

func TestAdminTopicEditFormPage(t *testing.T) {
	topicID := 4
	queries := &db.QuerierStub{
		GetForumTopicByIdReturns: &db.Forumtopic{
			Idforumtopic:                 int32(topicID),
			ForumcategoryIdforumcategory: 1,
			Title:                        sql.NullString{String: "t", Valid: true},
			Description:                  sql.NullString{String: "d", Valid: true},
			Threads:                      sql.NullInt32{Int32: 2, Valid: true},
			Comments:                     sql.NullInt32{Int32: 3, Valid: true},
			Lastaddition:                 sql.NullTime{Time: time.Now(), Valid: true},
			Handler:                      "",
		},
		GetAllForumCategoriesReturns: []*db.Forumcategory{
			{
				Idforumcategory:              1,
				ForumcategoryIdforumcategory: 0,
				Title:                        sql.NullString{String: "cat", Valid: true},
				Description:                  sql.NullString{String: "desc", Valid: true},
			},
		},
		AdminListRolesReturns: []*db.Role{
			{ID: 1, Name: "role", CanLogin: true},
		},
	}

	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
	r := httptest.NewRequest("GET", "/admin/forum/topics/topic/"+strconv.Itoa(topicID)+"/edit", nil)
	r = mux.SetURLVars(r, map[string]string{"topic": strconv.Itoa(topicID)})
	r = r.WithContext(context.WithValue(r.Context(), consts.KeyCoreData, cd))
	w := httptest.NewRecorder()
	AdminTopicEditFormPage(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d", w.Code, http.StatusOK)
	}
}

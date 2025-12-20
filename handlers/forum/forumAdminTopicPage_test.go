package forum

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestAdminTopicPage(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	mock.MatchExpectationsInOrder(false)

	topicID := 4
	rows := sqlmock.NewRows([]string{"idforumtopic", "lastposter", "forumcategory_idforumcategory", "language_id", "title", "description", "threads", "comments", "lastaddition", "handler", "lastposterusername"}).
		AddRow(topicID, 0, 1, 0, "t", "d", 2, 3, time.Now(), "", nil)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT t.idforumtopic, t.lastposter, t.forumcategory_idforumcategory, t.language_id, t.title, t.description, t.threads, t.comments, t.lastaddition, t.handler, lu.username AS LastPosterUsername FROM forumtopic t")).WillReturnRows(rows)

	grantsRows := sqlmock.NewRows([]string{"id", "section", "action", "role_name", "username"}).
		AddRow(1, "forum", "see", "Anyone", nil)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT g.id, g.section, g.action, r.name AS role_name, u.username FROM grants g")).WillReturnRows(grantsRows)

	cd := common.NewCoreData(context.Background(), db.New(conn), config.NewRuntimeConfig())
	r := httptest.NewRequest("GET", "/admin/forum/topics/topic/"+strconv.Itoa(topicID), nil)
	r = mux.SetURLVars(r, map[string]string{"topic": strconv.Itoa(topicID)})
	r = r.WithContext(context.WithValue(r.Context(), consts.KeyCoreData, cd))
	w := httptest.NewRecorder()
	AdminTopicPage(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d", w.Code, http.StatusOK)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestAdminTopicEditFormPage(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	mock.MatchExpectationsInOrder(false)

	topicID := 4
	topicRows := sqlmock.NewRows([]string{"idforumtopic", "lastposter", "forumcategory_idforumcategory", "language_id", "title", "description", "threads", "comments", "lastaddition", "handler", "lastposterusername"}).
		AddRow(topicID, 0, 1, 0, "t", "d", 2, 3, time.Now(), "", nil)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT t.idforumtopic, t.lastposter, t.forumcategory_idforumcategory, t.language_id, t.title, t.description, t.threads, t.comments, t.lastaddition, t.handler, lu.username AS LastPosterUsername FROM forumtopic t")).WillReturnRows(topicRows)

	catRows := sqlmock.NewRows([]string{"idforumcategory", "forumcategory_idforumcategory", "language_id", "title", "description"}).
		AddRow(1, 0, 0, "cat", "desc")
	mock.ExpectQuery("SELECT").WillReturnRows(catRows)

	roleRows := sqlmock.NewRows([]string{"id", "name", "can_login", "is_admin", "private_labels", "public_profile_allowed_at"}).
		AddRow(1, "role", true, false, true, nil)
	mock.ExpectQuery("SELECT").WillReturnRows(roleRows)

	cd := common.NewCoreData(context.Background(), db.New(conn), config.NewRuntimeConfig())
	r := httptest.NewRequest("GET", "/admin/forum/topics/topic/"+strconv.Itoa(topicID)+"/edit", nil)
	r = mux.SetURLVars(r, map[string]string{"topic": strconv.Itoa(topicID)})
	r = r.WithContext(context.WithValue(r.Context(), consts.KeyCoreData, cd))
	w := httptest.NewRecorder()
	AdminTopicEditFormPage(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d", w.Code, http.StatusOK)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

package forum

import (
	"context"
	"net/http"
	"net/http/httptest"
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
	rows := sqlmock.NewRows([]string{"idforumtopic", "lastposter", "forumcategory_idforumcategory", "language_idlanguage", "title", "description", "threads", "comments", "lastaddition", "handler", "LastPosterUsername"}).
		AddRow(topicID, 0, 1, 0, "t", "d", 2, 3, time.Now(), "", nil)
	mock.ExpectQuery("WITH").WillReturnRows(rows)

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
	topicRows := sqlmock.NewRows([]string{"idforumtopic", "lastposter", "forumcategory_idforumcategory", "language_idlanguage", "title", "description", "threads", "comments", "lastaddition", "handler", "LastPosterUsername"}).
		AddRow(topicID, 0, 1, 0, "t", "d", 2, 3, time.Now(), "", nil)
	mock.ExpectQuery("WITH").WillReturnRows(topicRows)

	catRows := sqlmock.NewRows([]string{"idforumcategory", "forumcategory_idforumcategory", "language_idlanguage", "title", "description"}).
		AddRow(1, 0, 0, "cat", "desc")
	mock.ExpectQuery("SELECT").WillReturnRows(catRows)

	roleRows := sqlmock.NewRows([]string{"id", "name", "can_login", "is_admin", "public_profile_allowed_at"}).
		AddRow(1, "role", true, false, nil)
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

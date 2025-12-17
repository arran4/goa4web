package privateforum

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestPage_NoAccess(t *testing.T) {
	cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig())
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	w := httptest.NewRecorder()
	PrivateForumPage(w, req)

	if body := w.Body.String(); !strings.Contains(body, "Forbidden") {
		t.Fatalf("expected no access message, got %q", body)
	}
}

func TestPage_Access(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}))
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	mock.ExpectQuery("SELECT 1 FROM grants").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectQuery("SELECT 1 FROM grants").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	topicRows := sqlmock.NewRows([]string{"idforumtopic", "lastposter", "forumcategory_idforumcategory", "language_id", "title", "description", "threads", "comments", "lastaddition", "handler", "LastPosterUsername"})
	mock.ExpectQuery(regexp.QuoteMeta("SELECT t.idforumtopic")).
		WithArgs(sql.NullInt32{}).
		WillReturnRows(topicRows)
	mock.ExpectQuery("SELECT 1 FROM grants").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectQuery("SELECT 1 FROM grants").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	w := httptest.NewRecorder()
	PrivateForumPage(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "Private Topics") {
		t.Fatalf("expected private topics page, got %q", body)
	}
	if !strings.Contains(body, "<form id=\"private-form\"") {
		t.Fatalf("expected create form, got %q", body)
	}
}

func TestPage_SeeNoCreate(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
	cd.UserID = 1

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	mock.ExpectQuery("SELECT 1 FROM grants").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectQuery("SELECT 1 FROM grants").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectQuery("SELECT 1 FROM grants").WillReturnError(sql.ErrNoRows)

	w := httptest.NewRecorder()
	PrivateForumPage(w, req)

	body := w.Body.String()
	if strings.Contains(body, "Start conversation") {
		t.Fatalf("unexpected create form, got %q", body)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

package privateforum

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"time"

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
	cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}))
	cd.AdminMode = true
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	w := httptest.NewRecorder()
	PrivateForumPage(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "Private Topics") {
		t.Fatalf("expected private topics page, got %q", body)
	}
	if !strings.Contains(body, "topic-controls") {
		t.Fatalf("expected create form, got %q", body)
	}
}

func TestPage_SeeNoCreate(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	mock.MatchExpectationsInOrder(false)

	queries := db.New(conn)
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())
	cd.UserID = 1
	cd.AdminMode = false
	cachePrivateTopics(cd, nil)

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

func TestPage_AdminLinks(t *testing.T) {
	cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}))
	cd.AdminMode = true

	// Inject a mock topic
	topic := &common.PrivateTopic{
		ListPrivateTopicsByUserIDRow: &db.ListPrivateTopicsByUserIDRow{},
	}
	topic.Idforumtopic = 123
	topic.Title.String = "Secret Plans"
	topic.Title.Valid = true
	topic.Lastaddition.Time = time.Now()
	topic.Lastaddition.Valid = true
	cachePrivateTopics(cd, []*common.PrivateTopic{topic})

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	// Mock the template engine to handle tableTopics if needed, but since we are running in unit test without full template loading,
	// we might run into template missing error if we don't set it up.
	// However, PrivateForumPage uses handlers.TemplateHandler which uses global templates.
	// We need to ensure templates are loaded or at least available.
	// The existing TestPage_Access asserts body content, so templates must be working or mocked in some way.
	// core/templates/site/* are embedded.

	// But wait, the previous tests (TestPage_Access) pass. They use PrivateForumPage.
	// This implies the standard template loading works in these tests or is bypassed?
	// The tests are in package privateforum, so they use the same package.
	// But `handlers.TemplateHandler` relies on `core.templates`.
	// If `TestMain` is not setting up templates, they might be empty.
	// Let's check `handlers/privateforum/main_test.go`.

	w := httptest.NewRecorder()
	PrivateForumPage(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "/admin/forum/topics/topic/123/edit") {
		t.Errorf("expected admin link for topic 123, got body length %d", len(body))
	}
}

func cachePrivateTopics(cd *common.CoreData, topics []*common.PrivateTopic) {
	v := reflect.ValueOf(cd).Elem().FieldByName("privateForumTopics")
	ptr := reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
	method := ptr.Addr().MethodByName("Set")
	if !method.IsValid() {
		return
	}
	method.Call([]reflect.Value{reflect.ValueOf(topics)})
}

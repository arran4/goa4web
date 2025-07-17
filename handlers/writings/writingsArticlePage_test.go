package writings

import (
	"context"
	"database/sql"
	corecommon "github.com/arran4/goa4web/core/common"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/core"
	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
)

func TestArticleReplyActionPage_UsesArticleParam(t *testing.T) {
	dbconn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer dbconn.Close()

	queries := db.New(dbconn)
	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test-session"

	form := url.Values{}
	form.Set("replytext", "hi")
	form.Set("language", "1")
	req := httptest.NewRequest("POST", "/writings/article/1", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = mux.SetURLVars(req, map[string]string{"article": "1"})

	w := httptest.NewRecorder()
	sess, _ := store.Get(req, core.SessionName)
	sess.Values["UID"] = int32(1)
	sess.Save(req, w)
	for _, c := range w.Result().Cookies() {
		req.AddCookie(c)
	}

	ctx := context.WithValue(req.Context(), handlers.KeyQueries, queries)
	ctx = context.WithValue(ctx, handlers.KeyCoreData, &corecommon.CoreData{})
	req = req.WithContext(ctx)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT w.idwriting")).
		WithArgs(int32(1), int32(1), sql.NullInt32{Int32: 1, Valid: true}).
		WillReturnError(sqlmock.ErrCancelled)

	rr := httptest.NewRecorder()
	ArticleReplyActionPage(rr, req)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	if rr.Result().StatusCode != http.StatusInternalServerError {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
}

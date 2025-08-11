package languages

import (
	"context"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestDeleteLanguageTask_PreventDeletion(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer dbMock.Close()
	queries := db.New(dbMock)

	form := url.Values{}
	form.Set("cid", "1")

	req := httptest.NewRequest("POST", "/admin/languages/language/1/edit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(context.Background(), queries, cfg, common.WithUserRoles([]string{"administrator"}))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	langRows := sqlmock.NewRows([]string{"idlanguage", "nameof"}).AddRow(1, "en")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT idlanguage, nameof\nFROM language")).WillReturnRows(langRows)

	countRows := sqlmock.NewRows([]string{"comments", "writings", "blogs", "news", "links"}).AddRow(1, 0, 0, 0, 0)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT\n    (SELECT COUNT(*) FROM comments WHERE comments.language_idlanguage = ?) AS comments,\n    (SELECT COUNT(*) FROM writing WHERE writing.language_idlanguage = ?) AS writings,\n    (SELECT COUNT(*) FROM blogs WHERE blogs.language_idlanguage = ?) AS blogs,\n    (SELECT COUNT(*) FROM site_news WHERE site_news.language_idlanguage = ?) AS news,\n    (SELECT COUNT(*) FROM linker WHERE linker.language_id = ?) AS links")).WithArgs(int32(1), int32(1), int32(1), int32(1), int32(1)).WillReturnRows(countRows)

	result := deleteLanguageTask.Action(rr, req)
	if result == nil {
		t.Fatalf("expected error, got nil")
	}
	if _, ok := result.(error); !ok {
		t.Fatalf("expected error result, got %T", result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

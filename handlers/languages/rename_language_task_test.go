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

func TestRenameLanguageTask_Action(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer dbMock.Close()
	queries := db.New(dbMock)

	form := url.Values{}
	form.Set("cid", "1")
	form.Set("cname", "fr")

	req := httptest.NewRequest("POST", "/admin/languages/language/1/edit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(context.Background(), queries, cfg, common.WithUserRoles([]string{"administrator"}))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	langRows := sqlmock.NewRows([]string{"idlanguage", "nameof"}).AddRow(1, "en")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT idlanguage, nameof\nFROM language")).WillReturnRows(langRows)

	idRow := sqlmock.NewRows([]string{"idlanguage"}).AddRow(1)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT idlanguage FROM language WHERE nameof = ?")).WithArgs("en").WillReturnRows(idRow)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE language\nSET nameof = ?\nWHERE idlanguage = ?")).WithArgs("fr", int32(1)).WillReturnResult(sqlmock.NewResult(0, 1))

	result := renameLanguageTask.Action(rr, req)
	if result != nil {
		t.Fatalf("expected nil, got %v", result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

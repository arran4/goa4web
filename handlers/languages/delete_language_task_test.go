package languages

import (
	"context"
	"database/sql"
	"fmt"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

type deleteLanguageQueries struct {
	db.Querier
	languages   []*db.Language
	usageCounts *db.AdminLanguageUsageCountsRow
}

func (q *deleteLanguageQueries) SystemListLanguages(context.Context) ([]*db.Language, error) {
	return q.languages, nil
}

func (q *deleteLanguageQueries) AdminLanguageUsageCounts(ctx context.Context, arg db.AdminLanguageUsageCountsParams) (*db.AdminLanguageUsageCountsRow, error) {
	if arg.LangID.Int32 != 1 {
		return nil, fmt.Errorf("unexpected language id: %v", arg.LangID.Int32)
	}
	return q.usageCounts, nil
}

func (q *deleteLanguageQueries) SystemCheckGrant(_ context.Context, arg db.SystemCheckGrantParams) (int32, error) {
	if arg.Section == common.AdminAccessSection && arg.Action == common.AdminAccessAction {
		return 1, nil
	}
	return 0, fmt.Errorf("unexpected grant check: %#v", arg)
}

func (q *deleteLanguageQueries) SystemCheckRoleGrant(context.Context, db.SystemCheckRoleGrantParams) (int32, error) {
	return 0, sql.ErrNoRows
}

func TestDeleteLanguageTask_PreventDeletion(t *testing.T) {
	queries := &deleteLanguageQueries{
		languages:   []*db.Language{{ID: 1, Nameof: sql.NullString{String: "en", Valid: true}}},
		usageCounts: &db.AdminLanguageUsageCountsRow{Comments: 1},
	}

	form := url.Values{}
	form.Set("cid", "1")

	req := httptest.NewRequest("POST", "/admin/languages/language/1/edit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(context.Background(), queries, cfg, common.WithUserRoles([]string{}))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	result := deleteLanguageTask.Action(rr, req)
	if result == nil {
		t.Fatalf("expected error, got nil")
	}
	if _, ok := result.(error); !ok {
		t.Fatalf("expected error result, got %T", result)
	}
}

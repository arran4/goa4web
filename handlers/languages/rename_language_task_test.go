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

type renameLanguageQueries struct {
	db.Querier
	languages  []*db.Language
	renameArgs []db.AdminRenameLanguageParams
}

func (q *renameLanguageQueries) SystemListLanguages(context.Context) ([]*db.Language, error) {
	return q.languages, nil
}

func (q *renameLanguageQueries) SystemGetLanguageIDByName(_ context.Context, name sql.NullString) (int32, error) {
	if name.String != "en" {
		return 0, fmt.Errorf("unexpected old name: %q", name.String)
	}
	return 1, nil
}

func (q *renameLanguageQueries) AdminRenameLanguage(_ context.Context, arg db.AdminRenameLanguageParams) error {
	q.renameArgs = append(q.renameArgs, arg)
	return nil
}

func (q *renameLanguageQueries) SystemCheckGrant(_ context.Context, arg db.SystemCheckGrantParams) (int32, error) {
	return 0, fmt.Errorf("unexpected grant check: %#v", arg)
}

func (q *renameLanguageQueries) SystemCheckRoleGrant(context.Context, db.SystemCheckRoleGrantParams) (int32, error) {
	return 0, sql.ErrNoRows
}

func (q *renameLanguageQueries) GetAdministratorUserRole(ctx context.Context, usersIdusers int32) (*db.UserRole, error) {
	return &db.UserRole{}, nil
}

func TestRenameLanguageTask_Action(t *testing.T) {
	queries := &renameLanguageQueries{
		languages: []*db.Language{{ID: 1, Nameof: sql.NullString{String: "en", Valid: true}}},
	}

	form := url.Values{}
	form.Set("cid", "1")
	form.Set("cname", "fr")

	req := httptest.NewRequest("POST", "/admin/languages/language/1/edit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(context.Background(), queries, cfg, common.WithUserRoles([]string{}))
	cd.UserID = 1
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	result := renameLanguageTask.Action(rr, req)
	if result != nil {
		t.Fatalf("expected nil, got %v", result)
	}
	if len(queries.renameArgs) != 1 {
		t.Fatalf("expected rename call, got %d", len(queries.renameArgs))
	}
	if arg := queries.renameArgs[0]; arg.ID != 1 || arg.Nameof.String != "fr" {
		t.Fatalf("unexpected rename args: %#v", arg)
	}
}

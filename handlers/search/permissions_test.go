package search

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

const systemCheckGrant = `-- name: SystemCheckGrant :one
WITH role_ids AS (
    SELECT DISTINCT ur.role_id AS id FROM user_roles ur WHERE ur.users_idusers = ?
)
SELECT 1 FROM grants g
WHERE g.section = ?
  AND (g.item = ? OR g.item IS NULL)
  AND g.action = ?
  AND g.active = 1
  AND (g.item_id = ? OR g.item_id IS NULL)
  AND (g.user_id = ? OR g.user_id IS NULL)
  AND (g.role_id IS NULL OR g.role_id IN (SELECT id FROM role_ids))
LIMIT 1
`

func TestCanSearch(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())

	// No grants
	mock.ExpectQuery(regexp.QuoteMeta(systemCheckGrant)).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta(systemCheckGrant)).WillReturnError(sql.ErrNoRows)
	if common.CanSearch(cd, "news") {
		t.Fatalf("expected false")
	}

	// Global grant only
	mock.ExpectQuery(regexp.QuoteMeta(systemCheckGrant)).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta(systemCheckGrant)).WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	if !common.CanSearch(cd, "news") {
		t.Fatalf("expected true with global grant")
	}

	// Grant present for section
	mock.ExpectQuery(regexp.QuoteMeta(systemCheckGrant)).WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	if !common.CanSearch(cd, "news") {
		t.Fatalf("expected true with section grant")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

type mockQuerierSearchPage struct {
	db.Querier
	grants map[string]bool
}

func (m *mockQuerierSearchPage) SystemCheckGrant(ctx context.Context, arg db.SystemCheckGrantParams) (int32, error) {
	parts := []string{arg.Section, arg.Item.String, arg.Action}
	key := strings.Join(parts, "_")
	if m.grants[key] {
		return 1, nil
	}
	return 0, fmt.Errorf("no grant for %s", key)
}

func (m *mockQuerierSearchPage) SystemCheckRoleGrant(ctx context.Context, arg db.SystemCheckRoleGrantParams) (int32, error) {
	return 0, nil
}

func TestGetSearchPageData(t *testing.T) {
	testCases := []struct {
		name              string
		grants            map[string]bool
		expectAnySearch   bool
		expectForumSearch bool
	}{
		{
			name:              "no grants",
			grants:            map[string]bool{},
			expectAnySearch:   false,
			expectForumSearch: false,
		},
		{
			name: "general search grant only",
			grants: map[string]bool{
				"search__search": true,
			},
			expectAnySearch:   false,
			expectForumSearch: false,
		},
		{
			name: "forum see grant only",
			grants: map[string]bool{
				"forum__see": true,
			},
			expectAnySearch:   false,
			expectForumSearch: false,
		},
		{
			name: "general search and forum see grant",
			grants: map[string]bool{
				"search__search": true,
				"forum__see":     true,
			},
			expectAnySearch:   true,
			expectForumSearch: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			querier := &mockQuerierSearchPage{grants: tc.grants}
			cd := common.NewCoreData(context.Background(), querier, config.NewRuntimeConfig())

			data := GetSearchPageData(cd)

			if data.AnySearch != tc.expectAnySearch {
				t.Errorf("expected AnySearch to be %v, but got %v", tc.expectAnySearch, data.AnySearch)
			}
			if data.CanSearchForum != tc.expectForumSearch {
				t.Errorf("expected CanSearchForum to be %v, but got %v", tc.expectForumSearch, data.CanSearchForum)
			}
		})
	}
}

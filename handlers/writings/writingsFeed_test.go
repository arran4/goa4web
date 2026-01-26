package writings

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/config"
	"github.com/DATA-DOG/go-sqlmock"
	"regexp"
)

func TestWritingsFeed(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	q := db.New(conn)

	req := httptest.NewRequest("GET", "http://example.com/writings/rss", nil)
	cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig(), common.WithSiteTitle("Site"))
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	mock.ExpectQuery(regexp.QuoteMeta("-- name: GetPublicWritings :many")).
		WillReturnRows(sqlmock.NewRows([]string{
			"idwriting", "users_idusers", "forumthread_id", "language_id",
			"writing_category_id", "title", "published", "timezone",
			"writing", "abstract", "private", "deleted_at", "last_index",
		}).
			AddRow(1, 1, 1, 1, 1, "Title", time.Now(), "UTC", "Content", "Abstract", false, nil, nil))

	// Mock HasGrant check if called. The feedGen calls LatestWritings which does check permissions.
	// But LatestWritings in CoreData calls HasGrant which calls Permissions() which queries DB.
	// However, LatestWritings logic:
	/*
		rows, err := cd.queries.GetPublicWritings(cd.ctx, params)
		...
		for _, row := range rows {
			if !cd.HasGrant("writing", "article", "see", row.Idwriting) {
				continue
			}
			writings = append(writings, row)
		}
	*/
	// Permissions() -> GetPermissionsByUserID. If UserID is 0, it returns nil and HasGrant returns false unless "anyone" role has permission.
	// cd.UserRoles() appends "anyone".
	// HasGrant -> HasRole -> ...

	// Actually, HasGrant implementation:
	/*
	func (cd *CoreData) HasGrant(section, itemType, action string, itemID int32) bool {
        // ... checks grants ...
		// checks roles ...
	}
	*/

	// Since creating proper mocks for permissions is complex, I will try to make HasGrant pass or ensure the test setup allows it.
	// If UserID is 0, it relies on "anyone" role.
	// HasGrant calls Permissions() which returns empty for UserID 0.

	// Wait, HasGrant also checks:
	/*
		if cd.queries != nil {
			g, err := cd.queries.SystemCheckGrant(cd.ctx, ...)
	*/
	// SystemCheckGrant is called.

	mock.ExpectQuery(regexp.QuoteMeta("-- name: SystemCheckGrant :one")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	feed, err := feedGen(req, cd)
	if err != nil {
		t.Fatalf("feedGen: %v", err)
	}

	if feed.Title != "Site - Latest writings" {
		t.Errorf("feed title incorrect: %s", feed.Title)
	}
	if len(feed.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(feed.Items))
	}
}

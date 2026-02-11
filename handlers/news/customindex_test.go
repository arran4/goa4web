package news

import (
	"context"
	"database/sql"
	"net/http"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

type NewsTestStub struct {
	*db.QuerierStub
	GetForumThreadIdByNewsPostIdFn func(ctx context.Context, idsitenews int32) (*db.GetForumThreadIdByNewsPostIdRow, error)
}

func (s *NewsTestStub) GetForumThreadIdByNewsPostId(ctx context.Context, idsitenews int32) (*db.GetForumThreadIdByNewsPostIdRow, error) {
	if s.GetForumThreadIdByNewsPostIdFn != nil {
		return s.GetForumThreadIdByNewsPostIdFn(ctx, idsitenews)
	}
	return nil, sql.ErrNoRows
}

func TestNewsCustomIndexItems_FetchAuthor(t *testing.T) {
	t.Run("Suppress 'new' label for author", func(t *testing.T) {
		q := &NewsTestStub{
			QuerierStub: testhelpers.NewQuerierStub(),
		}

		newsID := int32(123)
		authorID := int32(456)
		viewerID := authorID // Viewer is the author

		q.GetForumThreadIdByNewsPostIdFn = func(ctx context.Context, id int32) (*db.GetForumThreadIdByNewsPostIdRow, error) {
			if id == newsID {
				return &db.GetForumThreadIdByNewsPostIdRow{
					ForumthreadID: 1,
					Idusers:       sql.NullInt32{Int32: authorID, Valid: true},
				}, nil
			}
			return nil, sql.ErrNoRows
		}

		// ListContentPrivateLabels should return empty if we don't have explicit labels.
		// If the code passes 0 as authorID, logic in PrivateLabels (in core/common) adds "new" because 0 != viewerID.
		// If the code passes authorID (456), logic adds nothing because 456 == viewerID.
		// We rely on default behavior of PrivateLabels.

		// We need to ensure ListContentPrivateLabels returns empty list so we rely on generated labels.
		q.ListContentPrivateLabelsFn = func(arg db.ListContentPrivateLabelsParams) ([]*db.ListContentPrivateLabelsRow, error) {
			// Simulate that the user has marked it as read, so "unread" is suppressed.
			// This isolates the test to the "new" label which is controlled by author ID.
			return []*db.ListContentPrivateLabelsRow{
				{
					Label:  "unread",
					Invert: true,
				},
			}, nil
		}

		cd := common.NewCoreData(context.Background(), q, nil)
		cd.UserID = viewerID

		r, _ := http.NewRequest("GET", "/news/123", nil)
		r = mux.SetURLVars(r, map[string]string{"news": "123"})

		items := NewsCustomIndexItems(cd, r, nil)

		// Expectation: "Mark as read" should NOT be present because "new" label should be suppressed.
		found := false
		for _, item := range items {
			if item.Name == "Mark as read" {
				found = true
				break
			}
		}

		// Currently the code passes 0, so it SHOULD be found (failure case for desired behavior).
		// We assert False to verify the fix.
		assert.False(t, found, "Should NOT find 'Mark as read' item for author")
	})
}

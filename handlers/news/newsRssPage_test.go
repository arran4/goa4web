package news

import (
	"context"
	"database/sql"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestNewsRssPage(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		q := testhelpers.NewQuerierStub()
		q.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
			return 1, nil
		}
		q.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingReturns = []*db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow{
			{
				Writername:    sql.NullString{String: "Writer", Valid: true},
				Writerid:      sql.NullInt32{Int32: 1, Valid: true},
				Idsitenews:    1,
				ForumthreadID: 1,
				LanguageID:    sql.NullInt32{Int32: 1, Valid: true},
				UsersIdusers:  1,
				News:          sql.NullString{String: "News", Valid: true},
				Occurred:      sql.NullTime{Time: time.Now(), Valid: true},
				Timezone:      sql.NullString{String: "UTC", Valid: true},
				Comments:      sql.NullInt32{Int32: 0, Valid: true},
			},
		}

		req := httptest.NewRequest("GET", "http://example.com/news/rss", nil)
		cd := common.NewCoreData(req.Context(), q, config.NewRuntimeConfig(), common.WithSiteTitle("Site"))
		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		NewsRssPage(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status code: %d", w.Code)
		}

		var v struct {
			XMLName xml.Name
			Channel struct {
				Title string `xml:"title"`
			} `xml:"channel"`
		}
		if err := xml.Unmarshal(w.Body.Bytes(), &v); err != nil {
			t.Fatalf("xml parse: %v", err)
		}
		if v.Channel.Title != "Site - News feed" {
			t.Errorf("expected title 'Site - News feed' got %q", v.Channel.Title)
		}
	})
}

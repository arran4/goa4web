package news

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/handlers/handlertest"
	"github.com/arran4/goa4web/internal/db"
)

func TestNewsSearchFiltersUnauthorized(t *testing.T) {
	cd, stub, cleanup := handlertest.NewCoreData(t, context.Background())
	defer cleanup()

	stub.ListSiteNewsSearchFirstForListerReturns = []int32{1, 2}
	stub.GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCountReturns = []*db.GetNewsPostsByIdsForUserWithWriterIdAndThreadCommentCountRow{
		{
			Writername:    sql.NullString{String: "bob", Valid: true},
			Writerid:      sql.NullInt32{Int32: 1, Valid: true},
			Idsitenews:    1,
			ForumthreadID: 0,
			LanguageID:    sql.NullInt32{Int32: 1, Valid: true},
			UsersIdusers:  1,
			News:          sql.NullString{String: "text", Valid: true},
			Occurred:      sql.NullTime{Time: time.Unix(0, 0), Valid: true},
			Timezone:      sql.NullString{String: time.Local.String(), Valid: true},
			Comments:      sql.NullInt32{Int32: 0, Valid: true},
		},
	}

	form := url.Values{"searchwords": {"foo"}}
	req := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := req.Context()
	req = req.WithContext(ctx)
	news, emptyWords, noResults, err := cd.SearchNews(req, 1)
	if err != nil {
		t.Fatalf("NewsSearch: %v", err)
	}
	if emptyWords || noResults {
		t.Fatalf("unexpected flags")
	}
	if len(news) != 1 {
		t.Fatalf("expected 1 result got %d", len(news))
	}
	if news[0].Idsitenews != 1 {
		t.Errorf("unexpected id %d", news[0].Idsitenews)
	}
}

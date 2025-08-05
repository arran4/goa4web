package db

import (
	"strings"
	"testing"
)

func TestBlogQueriesAllowGlobalGrants(t *testing.T) {
	cases := []struct {
		name  string
		query string
	}{
		{"listBlogEntriesForLister", listBlogEntriesForLister},
		{"listBlogEntriesByAuthorForLister", listBlogEntriesByAuthorForLister},
		{"listBlogEntriesByIDsForLister", listBlogEntriesByIDsForLister},
		{"getBlogEntryForListerByID", getBlogEntryForListerByID},
		{"listBlogIDsBySearchWordFirstForLister", listBlogIDsBySearchWordFirstForLister},
		{"listBlogIDsBySearchWordNextForLister", listBlogIDsBySearchWordNextForLister},
		{"listBloggersForLister", listBloggersForLister},
		{"listBloggersSearchForLister", listBloggersSearchForLister},
	}

	itemSub := "(g.item = 'entry' OR g.item IS NULL)"
	idSub := "(g.item_id = b.idblogs OR g.item_id IS NULL)"

	for _, c := range cases {
		if !strings.Contains(c.query, itemSub) {
			t.Errorf("%s missing global item check", c.name)
		}
		if !strings.Contains(c.query, idSub) {
			t.Errorf("%s missing global item_id check", c.name)
		}
	}
}

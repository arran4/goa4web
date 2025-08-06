package db

import (
	"strings"
	"testing"
)

func TestForumQueriesAllowGlobalGrants(t *testing.T) {
	cases := []struct {
		name  string
		query string
	}{
		{"getForumTopicByIdForUser", getForumTopicByIdForUser},
		{"getForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostText", getForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostText},
		{"getAllForumTopicsByCategoryIdForUserWithLastPosterName", getAllForumTopicsByCategoryIdForUserWithLastPosterName},
		{"getAllForumTopicsForUser", getAllForumTopicsForUser},
	}

	for _, c := range cases {
		if !strings.Contains(c.query, "g.item='topic' OR g.item IS NULL") {
			t.Errorf("%s missing global item check", c.name)
		}
		if !strings.Contains(c.query, "g.item_id = t.idforumtopic OR g.item_id IS NULL") {
			t.Errorf("%s missing global item_id check", c.name)
		}
		if !strings.Contains(c.query, "WITH role_ids") {
			t.Errorf("%s missing role lookup", c.name)
		}
	}
}

package db

import (
	"fmt"
	"strings"
	"testing"

	"github.com/arran4/goa4web/core/consts"
)

func TestGetThreadLastPosterAndPermsForUser_AllowsGlobalGrants(t *testing.T) {
	if !strings.Contains(getThreadLastPosterAndPermsForUser, "g.item='topic' OR g.item IS NULL") {
		t.Errorf("missing global item check")
	}
	if !strings.Contains(getThreadLastPosterAndPermsForUser, "g.item_id = t.idforumtopic OR g.item_id IS NULL") {
		t.Errorf("missing global item_id check")
	}
}

func TestGetThreadLastPosterAndPermsForUser_UsesPrivateThreadGrants(t *testing.T) {
	want := fmt.Sprintf("g.section='%s'", consts.PermissionSectionPrivateForumThread)
	if !strings.Contains(getThreadLastPosterAndPermsForUser, want) {
		t.Errorf("private thread access does not use privateforum_thread grants")
	}
}

func TestPrivateThreadReadQueriesRequireExactThreadViewGrant(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		threadID string
	}{
		{"thread list", getForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostText, "th.idforumthread"},
		{"comment", getCommentByIdForUser, "th.idforumthread"},
		{"comment list", getCommentsByIdsForUserWithThreadInfo, "th.idforumthread"},
		{"thread comments", getCommentsByThreadIdForUser, "th.idforumthread"},
		{"section thread comments", getCommentsBySectionThreadIdForUser, "th.idforumthread"},
		{"private topic labels", getPrivateTopicThreadsAndLabelsForUser, "th.idforumthread"},
		{"unread thread list", listUnreadPrivateThreadsForUser, "th.idforumthread"},
		{"unread thread count", countUnreadPrivateThreadsForUser, "th.idforumthread"},
		{"first unrestricted search", listCommentIDsBySearchWordFirstForListerNotInRestrictedTopic, "fth.idforumthread"},
		{"next unrestricted search", listCommentIDsBySearchWordNextForListerNotInRestrictedTopic, "fth.idforumthread"},
		{"first restricted search", listCommentIDsBySearchWordFirstForListerInRestrictedTopic, "fth.idforumthread"},
		{"next restricted search", listCommentIDsBySearchWordNextForListerInRestrictedTopic, "fth.idforumthread"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			checks := []string{
				fmt.Sprintf("thread_grant.section = '%s'", consts.PermissionSectionPrivateForumThread),
				fmt.Sprintf("thread_grant.item = '%s'", consts.PermissionItemThread),
				fmt.Sprintf("thread_grant.action = '%s'", consts.PermissionActionView),
				"thread_grant.item_id = " + test.threadID,
				"thread_grant.active = 1",
			}
			for _, check := range checks {
				if !strings.Contains(test.query, check) {
					t.Errorf("private thread read query missing %q", check)
				}
			}
		})
	}
}

func TestPrivateThreadReadQueryNamesDescribeUserContext(t *testing.T) {
	queries := map[string]string{
		"thread":               getThreadLastPosterAndPermsForUser,
		"thread list":          getForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostText,
		"comment":              getCommentByIdForUser,
		"comment list":         getCommentsByIdsForUserWithThreadInfo,
		"thread comments":      getCommentsByThreadIdForUser,
		"section comments":     getCommentsBySectionThreadIdForUser,
		"private topic labels": getPrivateTopicThreadsAndLabelsForUser,
		"unread thread list":   listUnreadPrivateThreadsForUser,
		"unread thread count":  countUnreadPrivateThreadsForUser,
		"search":               listCommentIDsBySearchWordFirstForListerInRestrictedTopic,
	}

	for description, query := range queries {
		queryHeader, _, _ := strings.Cut(query, "\n")
		if !strings.Contains(queryHeader, "ForUser") && !strings.Contains(queryHeader, "ForLister") {
			t.Errorf("%s query does not identify its user authorization context: %q", description, queryHeader)
		}
	}
}

func TestSystemCopyPrivateTopicGrantsToThreadPreservesPrincipalsAndActions(t *testing.T) {
	checks := []string{
		"topic_grant.user_id, topic_grant.role_id",
		"thread_row.idforumthread, NULL, topic_grant.action",
		"topic_grant.action IN ('view', 'reply')",
		"thread_grant.action = topic_grant.action",
		"thread_grant.user_id <=> topic_grant.user_id",
		"thread_grant.role_id <=> topic_grant.role_id",
	}
	for _, check := range checks {
		if !strings.Contains(systemCopyPrivateTopicGrantsToThread, check) {
			t.Errorf("private topic grant copy query missing %q", check)
		}
	}
}

func TestPrivateCommentReadQueriesBindParentGrantToTopicNamespace(t *testing.T) {
	queries := map[string]string{
		"comment":      getCommentByIdForUser,
		"comment list": getCommentsByIdsForUserWithThreadInfo,
		"thread":       getCommentsByThreadIdForUser,
	}
	checks := []string{
		"((t.handler = 'private' AND g.section = 'privateforum') OR (t.handler <> 'private' AND g.section = 'forum'))",
		"((t.handler = 'private' AND g.item_id = t.idforumtopic) OR (t.handler <> 'private' AND (g.item_id = t.idforumtopic OR g.item_id IS NULL)))",
	}
	for name, query := range queries {
		for _, check := range checks {
			if !strings.Contains(query, check) {
				t.Errorf("%s query does not bind its parent grant with %q", name, check)
			}
		}
	}
}

func TestGetThreadLastPosterAndPermsForUserBindsParentGrantToHandler(t *testing.T) {
	checks := []string{
		"((t.handler = 'private' AND g.section = 'privateforum') OR (t.handler <> 'private' AND g.section = 'forum'))",
		"((t.handler = 'private' AND g.item_id = t.idforumtopic) OR (t.handler <> 'private' AND (g.item_id = t.idforumtopic OR g.item_id IS NULL)))",
	}
	for _, check := range checks {
		if !strings.Contains(getThreadLastPosterAndPermsForUser, check) {
			t.Errorf("ForUser thread query does not bind its parent grant with %q", check)
		}
	}
}

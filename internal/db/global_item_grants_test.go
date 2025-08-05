package db

import (
	"strings"
	"testing"
)

func TestGlobalItemGrantQueries(t *testing.T) {
	tests := []struct {
		name  string
		query string
		item  string
	}{
		{"UpdatePublicProfileEnabledAtForUser", updatePublicProfileEnabledAtForUser, "public_profile"},
		{"GetForumTopicByIdForUser", getForumTopicByIdForUser, "topic"},
		{"GetThreadLastPosterAndPerms", getThreadLastPosterAndPerms, "topic"},
		{"GetImagePostByIDForLister", getImagePostByIDForLister, "board"},
		{"GetWritingForListerByID", getWritingForListerByID, "article"},
		{"UpdateNewsPostForWriter", updateNewsPostForWriter, "post"},
		{"GetActiveAnnouncementWithNewsForLister", getActiveAnnouncementWithNewsForLister, "post"},
		{"GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUser", getLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingForUser, "link"},
		{"ListCommentIDsBySearchWordFirstForListerNotInRestrictedTopic", listCommentIDsBySearchWordFirstForListerNotInRestrictedTopic, "topic"},
	}

	for _, tt := range tests {
		expectedNoSpace := "g.item='" + tt.item + "' OR g.item IS NULL"
		expectedSpace := "g.item = '" + tt.item + "' OR g.item IS NULL"
		if !strings.Contains(tt.query, expectedNoSpace) && !strings.Contains(tt.query, expectedSpace) {
			t.Errorf("%s query missing global item grant clause", tt.name)
		}
	}
}

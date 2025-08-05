package news

import (
	"github.com/arran4/goa4web/core/common"
)

// canEditNewsPost reports whether cd can modify the specified news post.
func canEditNewsPost(cd *common.CoreData, postID int32) bool {
	if cd == nil {
		return false
	}
	if cd.HasGrant(common.SectionNews, common.ItemPost, common.ActionEdit, postID) && (cd.AdminMode || cd.UserID != 0) {
		return true
	}
	return false
}

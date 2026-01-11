package news

import (
	"github.com/arran4/goa4web/core/common"
)

// canEditNewsPost reports whether cd can modify the specified news post.
func canEditNewsPost(cd *common.CoreData, postID int32) bool {
	if cd == nil {
		return false
	}
	if cd.HasGrant("news", "post", "edit", postID) && (cd.AdminMode || cd.UserID != 0) {
		return true
	}
	return false
}

// CanPostNews reports whether the current user can create news posts.
func CanPostNews(cd *common.CoreData) bool {
	if cd == nil {
		return false
	}
	if cd.HasAdminRole() && cd.IsAdminMode() {
		return true
	}
	return cd.HasGrant("news", "post", "post", 0)
}

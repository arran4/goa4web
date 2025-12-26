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
// Administrators must be in Admin Mode.
// Users with explicit "news writer" or "content writer" roles can post anytime.
func CanPostNews(cd *common.CoreData) bool {
	if cd == nil {
		return false
	}
	if cd.IsAdmin() {
		return true
	}
	for _, r := range cd.UserRoles() {
		if r == "news writer" || r == "content writer" {
			return true
		}
	}
	return false
}

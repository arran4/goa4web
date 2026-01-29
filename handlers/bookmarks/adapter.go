package bookmarks

import (
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/gobookmarks"
)

type GoBookmarksUserProvider struct{}

func (p *GoBookmarksUserProvider) CurrentUser(r *http.Request) (*gobookmarks.User, error) {
	cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !ok || cd == nil {
		return nil, nil // Or error? gobookmarks might expect nil if not logged in
	}

	if cd.UserID == 0 {
		return nil, nil
	}

	// We need to fetch the full user object to get the username
	// Assuming CoreData exposes a way to get the user or we can use the ID
	// Step 267 showed cd.user is a lazy.Value[*db.User] but it is unexported?
	// No, it is `user lazy.Value[*db.User]`. Unexported field.
	// But `common.CoreData` usually has methods?
	// It has `Announcement`, `BlogEntryByID`, etc.
	// Does it have `User()`?
	// Step 267 didn't show `User()` method.
	// But it has `PagedUsers`, `PublicProfile` (takes userID).

	// `PublicProfile` returns `*db.SystemGetUserByIDRow`.
	u, err := cd.PublicProfile(cd.UserID)
	if err != nil {
		log.Printf("Error fetching user profile for ID %d: %v", cd.UserID, err)
		return nil, err
	}
	if u == nil {
		return nil, nil
	}

	return &gobookmarks.User{
		Login: u.Username.String,
	}, nil
}

func (p *GoBookmarksUserProvider) IsLoggedIn(r *http.Request) bool {
	cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	return ok && cd != nil && cd.UserID != 0
}

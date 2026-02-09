package blogs

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	navpkg "github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestBlogEditPageAccessControl(t *testing.T) {
	// Setup Router
	r := mux.NewRouter()
	navReg := navpkg.NewRegistry()
	cfg := config.NewRuntimeConfig()
	RegisterRoutes(r, cfg, navReg)

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	})

	queries := testhelpers.NewQuerierStub()
	blogID := int32(101)
	queries.GetBlogEntryForListerByIDRow = &db.GetBlogEntryForListerByIDRow{
		Idblogs:      blogID,
		UsersIdusers: 999, // Author is user 999
		IsOwner:      false, // This field seems to be derived in SQL, stub sets it directly
	}
	queries.ListGrantsReturns = []*db.Grant{}

	store := sessions.NewCookieStore([]byte("secret"))

	makeRequest := func(userID int32, grants []*db.Grant) *httptest.ResponseRecorder {
		// Mock SystemCheckGrant
		queries.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
			if grants == nil {
				return 0, sql.ErrNoRows
			}
			for _, g := range grants {
				if g.Section == arg.Section && g.Action == arg.Action {
					// Check Item match
					itemMatch := false
					if !g.Item.Valid && !arg.Item.Valid {
						itemMatch = true
					} else if g.Item.Valid && arg.Item.Valid && g.Item.String == arg.Item.String {
						itemMatch = true
					}

					// Check ItemID match
					itemIDMatch := false
					if !g.ItemID.Valid && !arg.ItemID.Valid {
						itemIDMatch = true
					} else if g.ItemID.Valid && arg.ItemID.Valid && g.ItemID.Int32 == arg.ItemID.Int32 {
						itemIDMatch = true
					}

					if itemMatch && itemIDMatch {
						return 1, nil
					}
				}
			}
			return 0, sql.ErrNoRows
		}

		req := httptest.NewRequest("GET", "/blogs/blog/101/edit", nil)
		req = mux.SetURLVars(req, map[string]string{"blog": "101"})

		session := sessions.NewSession(store, "session")
		session.Values["UID"] = userID

		cd := common.NewCoreData(req.Context(), queries, cfg, common.WithSession(session))
		cd.UserID = userID

		ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
		ctx = context.WithValue(ctx, core.ContextValues("session"), session)

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req.WithContext(ctx))
		return rr
	}

	// Case 1: No Grant. Should be 404 because Matcher fails.
	t.Run("No Grant", func(t *testing.T) {
		rr := makeRequest(123, nil)
		if rr.Code != http.StatusNotFound {
			t.Errorf("Expected 404 Not Found, got %d", rr.Code)
		}
	})

	// Case 2: Has Edit Grant but not Author.
	t.Run("Has Edit Grant Not Author", func(t *testing.T) {
		grant := &db.Grant{
			Section: "blogs",
			Item:    sql.NullString{String: "entry", Valid: true},
			ItemID:  sql.NullInt32{Int32: blogID, Valid: true},
			Action:  "edit",
		}

		rr := makeRequest(123, []*db.Grant{grant})

		if rr.Code != http.StatusNotFound {
			t.Errorf("Expected 404 (due to middleware authorship check), got %d", rr.Code)
		}
	})

	// Case 3: Author and Grant. Should be 200 (or at least passed middleware).
	t.Run("Author and Grant", func(t *testing.T) {
		grant := &db.Grant{
			Section: "blogs",
			Item:    sql.NullString{String: "entry", Valid: true},
			ItemID:  sql.NullInt32{Int32: blogID, Valid: true},
			Action:  "edit",
		}

		rr := makeRequest(999, []*db.Grant{grant})

		if rr.Code == http.StatusNotFound || rr.Code == http.StatusForbidden {
			t.Errorf("Expected success (or 500), got %d. Body: %s", rr.Code, rr.Body.String())
		}
	})
}

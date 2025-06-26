package router

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	adminhandlers "github.com/arran4/goa4web/handlers/admin"
	auth "github.com/arran4/goa4web/handlers/auth"
	blogs "github.com/arran4/goa4web/handlers/blogs"
	bookmarks "github.com/arran4/goa4web/handlers/bookmarks"
	hcommon "github.com/arran4/goa4web/handlers/common"
	faq "github.com/arran4/goa4web/handlers/faq"
	forum "github.com/arran4/goa4web/handlers/forum"
	imagebbs "github.com/arran4/goa4web/handlers/imagebbs"
	information "github.com/arran4/goa4web/handlers/information"
	linker "github.com/arran4/goa4web/handlers/linker"
	news "github.com/arran4/goa4web/handlers/news"
	search "github.com/arran4/goa4web/handlers/search"
	writings "github.com/arran4/goa4web/handlers/writings"

	corecommon "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
	userhandlers "github.com/arran4/goa4web/handlers/user"
	dbpkg "github.com/arran4/goa4web/internal/db"
	handlers "github.com/arran4/goa4web/pkg/handlers"
)

// RegisterRoutes sets up all application routes on r.
func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/main.css", handlers.MainCSS).Methods("GET")

	news.RegisterRoutes(r)
	faq.RegisterRoutes(r)
	blogs.RegisterRoutes(r)
	forum.RegisterRoutes(r)
	linker.RegisterRoutes(r)
	bookmarks.RegisterRoutes(r)
	imagebbs.RegisterRoutes(r)
	search.RegisterRoutes(r)
	writings.RegisterRoutes(r)
	information.RegisterRoutes(r)
	userhandlers.RegisterRoutes(r)
	auth.RegisterRoutes(r)
	registerAdminRoutes(r)

	// legacy redirects
	r.PathPrefix("/writing").HandlerFunc(handlers.RedirectPermanentPrefix("/writing", "/writings"))
	r.PathPrefix("/links").HandlerFunc(handlers.RedirectPermanentPrefix("/links", "/linker"))
}

// roleAllowed checks if the current request has one of the provided roles.
func roleAllowed(r *http.Request, roles ...string) bool {
	cd, ok := r.Context().Value(hcommon.KeyCoreData).(*corecommon.CoreData)
	if ok && cd != nil {
		for _, lvl := range roles {
			if cd.HasRole(lvl) {
				return true
			}
		}
		return false
	}

	user, uok := r.Context().Value(hcommon.KeyUser).(*dbpkg.User)
	queries, qok := r.Context().Value(hcommon.KeyQueries).(*dbpkg.Queries)
	if !uok || !qok {
		return false
	}
	section := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")[0]
	perm, err := queries.GetPermissionsByUserIdAndSectionAndSectionAll(r.Context(), dbpkg.GetPermissionsByUserIdAndSectionAndSectionAllParams{
		UsersIdusers: user.Idusers,
		Section:      sql.NullString{String: section, Valid: true},
	})
	if err != nil || !perm.Level.Valid {
		return false
	}
	cd = &corecommon.CoreData{SecurityLevel: perm.Level.String}
	for _, lvl := range roles {
		if cd.HasRole(lvl) {
			return true
		}
	}
	return false
}

// RoleCheckerMiddleware ensures the user has one of the supplied roles.
func RoleCheckerMiddleware(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !roleAllowed(r, roles...) {
				err := templates.GetCompiledTemplates(corecommon.NewFuncs(r)).ExecuteTemplate(w, "noAccessPage.gohtml", r.Context().Value(hcommon.KeyCoreData).(*corecommon.CoreData))
				if err != nil {
					log.Printf("Template Error: %s", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// AdminCheckerMiddleware ensures the requester has administrator rights.
func AdminCheckerMiddleware(next http.Handler) http.Handler {
	return RoleCheckerMiddleware("administrator")(next)
}

func registerAdminRoutes(r *mux.Router) {
	ar := r.PathPrefix("/admin").Subrouter()
	ar.Use(AdminCheckerMiddleware)
	adminhandlers.RegisterRoutes(ar)
}

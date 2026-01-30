package bookmarks

import (
	"net/http"

	"github.com/arran4/gobookmarks"
	"github.com/arran4/gobookmarks/app"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"

	"github.com/arran4/goa4web/internal/router"

	navpkg "github.com/arran4/goa4web/internal/navigation"
)

// SectionWeight defines the order in the navigation menu.
const SectionWeight = 50

// RegisterRoutes attaches the public bookmarks endpoints to r.
func RegisterRoutes(r *mux.Router, cfg *config.RuntimeConfig, navReg *navpkg.Registry) {
	navReg.RegisterIndexLink("Bookmarks", "/bookmarks", SectionWeight)

	provider := &GoBookmarksUserProvider{}

	// Initialize gobookmarks configuration
	gbCfg := &app.Config{
		SessionName:          cfg.SessionName,
		BaseURL:              "/bookmarks",
		ExternalURL:          cfg.HTTPHostname + "/bookmarks",
		DevMode:              cfg.DevMode,
		GithubClientID:       cfg.GithubClientID,
		GithubSecret:         cfg.GithubSecret,
		GithubServer:         cfg.GithubServer,
		GitlabClientID:       cfg.GitlabClientID,
		GitlabSecret:         cfg.GitlabSecret,
		GitlabServer:         cfg.GitlabServer,
		Title:                cfg.BookmarksTitle,
		CssColumns:           cfg.BookmarksCssColumns,
		NoFooter:             cfg.BookmarksNoFooter,
		LocalGitPath:         cfg.BookmarksLocalGitPath,
		CommitsPerPage:       cfg.BookmarksCommitsPerPage,
		FaviconCacheDir:      cfg.BookmarksFaviconCacheDir,
		FaviconCacheSize:     int64(cfg.BookmarksFaviconCacheSize),
		FaviconMaxCacheCount: cfg.BookmarksFaviconMaxCacheCount,
	}

	repo := &ContextRepo{}

	gobookmarks.RegisterProvider(repo)

	application := app.NewApp(nil, core.Store, repo, provider, gbCfg)

	router := gobookmarks.NewRouter(application)

	// Mount it
	r.PathPrefix("/bookmarks").Handler(http.StripPrefix("/bookmarks", router))
}

// Register registers the bookmarks router module.
func Register(reg *router.Registry) {
	reg.RegisterModule("bookmarks", nil, RegisterRoutes)
}

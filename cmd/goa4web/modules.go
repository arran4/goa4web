package main

import (
	"github.com/arran4/goa4web/handlers/admin"
	"github.com/arran4/goa4web/handlers/auth"
	"github.com/arran4/goa4web/handlers/blogs"
	"github.com/arran4/goa4web/handlers/bookmarks"
	"github.com/arran4/goa4web/handlers/faq"
	"github.com/arran4/goa4web/handlers/forum"
	"github.com/arran4/goa4web/handlers/imagebbs"
	"github.com/arran4/goa4web/handlers/images"
	"github.com/arran4/goa4web/handlers/languages"
	"github.com/arran4/goa4web/handlers/linker"
	"github.com/arran4/goa4web/handlers/news"
	"github.com/arran4/goa4web/handlers/search"
	"github.com/arran4/goa4web/handlers/user"
	"github.com/arran4/goa4web/handlers/writings"

	"github.com/arran4/goa4web/internal/router"
)

var extraRegistrations []func(*router.Registry)

// registerModules registers all router modules used by the application.
func registerModules(reg *router.Registry, ah *admin.Handlers) {
	ah.Register(reg)
	auth.Register(reg)
	blogs.Register(reg)
	bookmarks.Register(reg)
	faq.Register(reg)
	forum.Register(reg)
	imagebbs.Register(reg)
	languages.Register(reg)
	linker.Register(reg)
	news.Register(reg)
	search.Register(reg)
	images.Register(reg)
	user.Register(reg)
	writings.Register(reg)
	for _, fn := range extraRegistrations {
		fn(reg)
	}
}

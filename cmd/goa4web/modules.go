package main

import (
	admin "github.com/arran4/goa4web/handlers/admin"
	auth "github.com/arran4/goa4web/handlers/auth"
	blogs "github.com/arran4/goa4web/handlers/blogs"
	bookmarks "github.com/arran4/goa4web/handlers/bookmarks"
	faq "github.com/arran4/goa4web/handlers/faq"
	forum "github.com/arran4/goa4web/handlers/forum"
	imagebbs "github.com/arran4/goa4web/handlers/imagebbs"
	images "github.com/arran4/goa4web/handlers/images"
	languages "github.com/arran4/goa4web/handlers/languages"
	linker "github.com/arran4/goa4web/handlers/linker"
	news "github.com/arran4/goa4web/handlers/news"
	search "github.com/arran4/goa4web/handlers/search"
	user "github.com/arran4/goa4web/handlers/user"
	writings "github.com/arran4/goa4web/handlers/writings"

	"github.com/arran4/goa4web/internal/router"
)

var extraRegistrations []func(*router.Registry)

// registerModules registers all router modules used by the application.
func registerModules(reg *router.Registry) {
	admin.Register(reg)
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

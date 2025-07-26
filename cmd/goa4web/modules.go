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

	nav "github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/router"
)

var extraRegistrations []func(*router.Registry)

// registerModules registers all router modules used by the application.
func registerModules(reg *router.Registry, navReg *nav.Registry) {
	admin.Register(reg, navReg)
	auth.Register(reg, navReg)
	blogs.Register(reg, navReg)
	bookmarks.Register(reg, navReg)
	faq.Register(reg, navReg)
	forum.Register(reg, navReg)
	imagebbs.Register(reg, navReg)
	languages.Register(reg, navReg)
	linker.Register(reg, navReg)
	news.Register(reg, navReg)
	search.Register(reg, navReg)
	images.Register(reg, navReg)
	user.Register(reg, navReg)
	writings.Register(reg, navReg)
	for _, fn := range extraRegistrations {
		fn(reg)
	}
}

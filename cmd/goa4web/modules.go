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
)

// init registers all router modules used by the application.
func init() {
	admin.Register()
	auth.Register()
	blogs.Register()
	bookmarks.Register()
	faq.Register()
	forum.Register()
	imagebbs.Register()
	languages.Register()
	linker.Register()
	news.Register()
	search.Register()
	images.Register()
	user.Register()
	writings.Register()
}

package main

import (
	"github.com/arran4/goa4web/handlers/admin"
	"github.com/arran4/goa4web/handlers/auth"
	"github.com/arran4/goa4web/handlers/blogs"
	"github.com/arran4/goa4web/handlers/bookmarks"
	"github.com/arran4/goa4web/handlers/externallink"
	"github.com/arran4/goa4web/handlers/faq"
	"github.com/arran4/goa4web/handlers/forum"
	"github.com/arran4/goa4web/handlers/imagebbs"
	"github.com/arran4/goa4web/handlers/images"
	"github.com/arran4/goa4web/handlers/languages"
	"github.com/arran4/goa4web/handlers/linker"
	"github.com/arran4/goa4web/handlers/news"
	"github.com/arran4/goa4web/handlers/privateforum"
	"github.com/arran4/goa4web/handlers/search"
	"github.com/arran4/goa4web/handlers/user"
	"github.com/arran4/goa4web/handlers/writings"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/pkg/moduleapi"
)

// compiledModuleManifests returns the statically linked module manifest list.
//
// Distribution-specific repos can replace this assembly with their own package
// that imports manifests from external module repositories.
func compiledModuleManifests(ah *admin.Handlers) []moduleapi.Manifest {
	registerTasks := func(section string, ts []tasks.NamedTask) func(*tasks.Registry) {
		return func(reg *tasks.Registry) {
			for _, t := range ts {
				reg.Register(section, t)
			}
		}
	}

	return []moduleapi.Manifest{
		{
			Name:           "admin",
			RegisterRoutes: ah.Register,
			RegisterTasks:  registerTasks("admin", ah.RegisterTasks()),
		},
		{
			Name:           "auth",
			RegisterRoutes: auth.Register,
			RegisterTasks:  registerTasks("auth", auth.RegisterTasks()),
		},
		{
			Name:           "blogs",
			RegisterRoutes: blogs.Register,
			RegisterTasks:  registerTasks("blogs", blogs.RegisterTasks()),
		},
		{
			Name:           "bookmarks",
			RegisterRoutes: bookmarks.Register,
			RegisterTasks:  registerTasks("bookmarks", bookmarks.RegisterTasks()),
		},
		{
			Name:           "faq",
			RegisterRoutes: faq.Register,
			RegisterTasks:  registerTasks("faq", faq.RegisterTasks()),
		},
		{
			Name:           "forum",
			RegisterRoutes: forum.Register,
			RegisterTasks:  registerTasks("forum", forum.RegisterTasks()),
		},
		{
			Name:           "privateforum",
			RegisterRoutes: privateforum.Register,
			RegisterTasks:  registerTasks("privateforum", privateforum.RegisterTasks()),
		},
		{
			Name:           "images",
			RegisterRoutes: images.Register,
			RegisterTasks:  registerTasks("images", images.RegisterTasks()),
		},
		{
			Name:           "imagebbs",
			RegisterRoutes: imagebbs.Register,
			RegisterTasks:  registerTasks("imagebbs", imagebbs.RegisterTasks()),
		},
		{
			Name:           "linker",
			RegisterRoutes: linker.Register,
			RegisterTasks:  registerTasks("linker", linker.RegisterTasks()),
		},
		{
			Name:           "news",
			RegisterRoutes: news.Register,
			RegisterTasks:  registerTasks("news", news.RegisterTasks()),
		},
		{
			Name:           "search",
			RegisterRoutes: search.Register,
			RegisterTasks:  registerTasks("search", search.RegisterTasks()),
		},
		{
			Name:           "user",
			RegisterRoutes: user.Register,
			RegisterTasks:  registerTasks("user", user.RegisterTasks()),
		},
		{
			Name:           "writing",
			RegisterRoutes: writings.Register,
			RegisterTasks:  registerTasks("writing", writings.RegisterTasks()),
		},
		{
			Name:           "languages",
			RegisterRoutes: languages.Register,
		},
		{
			Name:           "externallink",
			RegisterRoutes: externallink.Register,
		},
	}
}

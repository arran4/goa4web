package main

import (
	"github.com/arran4/goa4web/handlers/admin"
	"github.com/arran4/goa4web/internal/router"
)

// registerModules registers all router modules used by the application.
func registerModules(reg *router.Registry, ah *admin.Handlers) {
	for _, manifest := range compiledModuleManifests(ah) {
		if manifest.RegisterRoutes == nil {
			continue
		}
		manifest.RegisterRoutes(reg)
	}
}

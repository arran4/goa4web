package goa4web

import (
	"net/http"

	corepkg "github.com/arran4/goa4web/core"
)

func handleDie(w http.ResponseWriter, message string) { corepkg.HandleDie(w, message) }

type IndexItem = corepkg.IndexItem
type CoreData = corepkg.CoreData
type ContextValues = corepkg.ContextValues
type Configuration = corepkg.Configuration

func CoreAdderMiddleware(next http.Handler) http.Handler {
	corepkg.GetSession = GetSession
	corepkg.SessionErrorRedirect = sessionErrorRedirect
	corepkg.NotificationsEnabled = notificationsEnabled
	corepkg.FeedsEnabled = appRuntimeConfig.FeedsEnabled
	return corepkg.CoreAdderMiddleware(next)
}

func DBAdderMiddleware(next http.Handler) http.Handler {
	corepkg.DBPool = dbPool
	corepkg.DBLogVerbosity = dbLogVerbosity
	return corepkg.DBAdderMiddleware(next)
}

func NewConfiguration() *Configuration { return corepkg.NewConfiguration() }

func X2c(what string) byte { return corepkg.X2c(what) }

func getPageSize(r *http.Request) int {
	corepkg.PageSizeDefault = appRuntimeConfig.PageSizeDefault
	corepkg.PageSizeMin = appRuntimeConfig.PageSizeMin
	corepkg.PageSizeMax = appRuntimeConfig.PageSizeMax
	return corepkg.GetPageSize(r)
}

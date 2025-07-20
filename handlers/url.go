package handlers

import (
	"net/http"
	"strings"

	"github.com/arran4/goa4web/config"
)

// AbsoluteURL returns an absolute URL using either the configured hostname or the request host.
func AbsoluteURL(r *http.Request, path string) string {
	base := "http://" + r.Host
	if config.AppRuntimeConfig.HTTPHostname != "" {
		base = strings.TrimRight(config.AppRuntimeConfig.HTTPHostname, "/")
	}
	return base + path
}

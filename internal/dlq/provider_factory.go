package dlq

import (
	"strings"

	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/runtimeconfig"
)

// ProviderFromConfig returns a DLQ implementation configured from cfg.
func ProviderFromConfig(cfg runtimeconfig.RuntimeConfig, q *dbpkg.Queries) DLQ {
	switch strings.ToLower(cfg.DLQProvider) {
	case "file":
		return &FileDLQ{Path: cfg.DLQFile}
	case "db":
		return DBDLQ{Queries: q}
	default:
		if cfg.DLQProvider != "" && cfg.DLQProvider != "log" {
			// unrecognised provider -> fallback to log
		}
		return LogDLQ{}
	}
}

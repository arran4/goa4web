package dlq

import (
	"log"
	"strings"

	"github.com/arran4/goa4web/config"
	dbpkg "github.com/arran4/goa4web/internal/db"
)

// ProviderFromConfig returns a DLQ implementation configured from cfg.
func (r *Registry) ProviderFromConfig(cfg config.RuntimeConfig, q *dbpkg.Queries) DLQ {
	names := strings.Split(cfg.DLQProvider, ",")
	var qs []DLQ
	for _, name := range names {
		n := strings.ToLower(strings.TrimSpace(name))
		if n == "" {
			continue
		}
		if f := r.lookupProvider(n); f != nil {
			qs = append(qs, f(cfg, q))
		} else {
			if n != "log" {
				log.Printf("unrecognised DLQ provider %q, falling back to log", n)
			}
			qs = append(qs, LogDLQ{})
		}
	}
	if len(qs) == 0 {
		return LogDLQ{}
	}
	if len(qs) == 1 {
		return qs[0]
	}
	return NewMulti(qs...)
}

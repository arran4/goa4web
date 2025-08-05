package middleware

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

func normalizeIP(ip string) string {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ip
	}
	if v4 := parsed.To4(); v4 != nil {
		return v4.String()
	}
	return parsed.String()
}

func requestIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		if comma := strings.IndexByte(ip, ','); comma >= 0 {
			ip = ip[:comma]
		}
		ip = strings.TrimSpace(ip)
		if host, _, err := net.SplitHostPort(ip); err == nil {
			ip = host
		}
		return normalizeIP(ip)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return normalizeIP(r.RemoteAddr)
	}
	return normalizeIP(host)
}

// SecurityHeadersMiddleware enforces IP bans and sets common security headers.
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := requestIP(r)
		if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
			bans, err := cd.Queries().ListActiveBans(r.Context())
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				w.WriteHeader(http.StatusInternalServerError)
				handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
				return
			}
			addr, parseErr := netip.ParseAddr(ip)
			if parseErr == nil {
				for _, b := range bans {
					if p, err := netip.ParsePrefix(b.IpNet); err == nil {
						if p.Contains(addr) {
							w.WriteHeader(http.StatusForbidden)
							handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
							return
						}
					} else if b.IpNet == ip {
						w.WriteHeader(http.StatusForbidden)
						handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
						return
					}
				}
			}
		}
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		var cfg *config.RuntimeConfig
		if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
			cfg = cd.Config
		}
		var hsts string
		if cfg != nil {
			hsts = cfg.HSTSHeaderValue
		}
		if hsts != "" {
			if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
				w.Header().Set("Strict-Transport-Security", hsts)
			} else if cfg != nil && strings.HasPrefix(strings.ToLower(cfg.HTTPHostname), "https://") {
				w.Header().Set("Strict-Transport-Security", hsts)
			}
		}
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

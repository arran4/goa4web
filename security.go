package goa4web

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/handlers/common"
	"net"
	"net/http"
	"net/netip"
	"strings"
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

func normalizeIPNet(ip string) string {
	ip = strings.TrimSpace(ip)
	if strings.Contains(ip, "/") {
		if p, err := netip.ParsePrefix(ip); err == nil {
			return p.String()
		}
		return ip
	}
	return normalizeIP(ip)
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

func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := requestIP(r)
		if queries, ok := r.Context().Value(common.KeyQueries).(*Queries); ok {
			bans, err := queries.ListActiveBans(r.Context())
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			addr, parseErr := netip.ParseAddr(ip)
			if parseErr == nil {
				for _, b := range bans {
					if p, err := netip.ParsePrefix(b.IpNet); err == nil {
						if p.Contains(addr) {
							http.Error(w, "Forbidden", http.StatusForbidden)
							return
						}
					} else if b.IpNet == ip {
						http.Error(w, "Forbidden", http.StatusForbidden)
						return
					}
				}
			}
		}
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		next.ServeHTTP(w, r)
	})
}

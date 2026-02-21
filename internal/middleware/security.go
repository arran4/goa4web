package middleware

import (
	"database/sql"
	"errors"
	"log"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

var (
	lastUntrustedProxyLogTime time.Time
	untrustedProxyLogMu       sync.Mutex
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

func requestIP(r *http.Request, cfg *config.RuntimeConfig) string {
	remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		remoteIP = r.RemoteAddr
	}
	remoteIP = normalizeIP(remoteIP)

	if cfg == nil || len(cfg.TrustedProxiesParsed) == 0 {
		if r.Header.Get("X-Forwarded-For") != "" {
			untrustedProxyLogMu.Lock()
			if time.Since(lastUntrustedProxyLogTime) > 24*time.Hour {
				log.Printf("Security Warning: X-Forwarded-For header detected but no trusted proxies configured. Ignoring header. Configure TRUSTED_PROXIES to trust specific proxies.")
				lastUntrustedProxyLogTime = time.Now()
			}
			untrustedProxyLogMu.Unlock()
		}
		return remoteIP
	}

	addr, err := netip.ParseAddr(remoteIP)
	if err != nil {
		return remoteIP
	}

	isTrusted := false
	for _, cidr := range cfg.TrustedProxiesParsed {
		if cidr.Contains(addr) {
			isTrusted = true
			break
		}
	}

	if !isTrusted {
		return remoteIP
	}

	xff := r.Header.Get("X-Forwarded-For")
	if xff == "" {
		return remoteIP
	}

	ips := strings.Split(xff, ",")
	currentIP := addr

	for i := len(ips) - 1; i >= 0; i-- {
		ipStr := strings.TrimSpace(ips[i])
		if host, _, err := net.SplitHostPort(ipStr); err == nil {
			ipStr = host
		}
		parsedIP, err := netip.ParseAddr(ipStr)
		if err != nil {
			return normalizeIP(currentIP.String())
		}

		isIpTrusted := false
		for _, cidr := range cfg.TrustedProxiesParsed {
			if cidr.Contains(parsedIP) {
				isIpTrusted = true
				break
			}
		}

		if isIpTrusted {
			currentIP = parsedIP
			continue
		} else {
			return normalizeIP(ipStr)
		}
	}

	return normalizeIP(currentIP.String())
}

// SecurityHeadersMiddleware enforces IP bans and sets common security headers.
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var cfg *config.RuntimeConfig
		if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
			cfg = cd.Config
		}
		ip := requestIP(r, cfg)
		if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
			bans, err := cd.Queries().ListActiveBans(r.Context())
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				w.WriteHeader(http.StatusInternalServerError)
				handlers.RenderErrorPage(w, r, common.ErrInternalServerError)
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
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' https://static.cloudflareinsights.com; style-src 'self' 'unsafe-inline'; img-src 'self' data:;")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		var hsts string
		if cfg != nil {
			hsts = cfg.HSTSHeaderValue
		}
		if hsts != "" {
			if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
				w.Header().Set("Strict-Transport-Security", hsts)
			} else if cfg != nil && strings.HasPrefix(strings.ToLower(cfg.BaseURL), "https://") {
				w.Header().Set("Strict-Transport-Security", hsts)
			}
		}
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

package common

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/arran4/goa4web/internal/sign"
)

// SanitizeBackURL validates raw and returns a safe back URL.
// Absolute URLs are allowed only when the host matches an allowed hostname
// or when accompanied by a valid signature via back_ts and back_sig.
func (cd *CoreData) SanitizeBackURL(r *http.Request, raw string) string {
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil {
		log.Printf("invalid back url %q: %v", raw, err)
		return ""
	}
	if !u.IsAbs() {
		return raw
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		log.Printf("invalid back scheme %q", raw)
		return ""
	}

	allowed := map[string]struct{}{}
	if r != nil && r.Host != "" {
		allowed[strings.ToLower(r.Host)] = struct{}{}
	}
	if cd != nil {
		hosts := strings.Fields(cd.Config.BaseURL)
		for _, h := range hosts {
			h = strings.TrimSpace(h)
			if h == "" {
				continue
			}
			if pu, err := url.Parse(h); err == nil && pu.Host != "" {
				h = pu.Host
			} else {
				h = strings.TrimSuffix(h, "/")
			}
			allowed[strings.ToLower(h)] = struct{}{}
		}
	}

	if _, ok := allowed[strings.ToLower(u.Host)]; ok {
		result := u.Path
		if u.RawQuery != "" {
			result += "?" + u.RawQuery
		}
		if u.Fragment != "" {
			result += "#" + u.Fragment
		}
		return result
	}

	sig := r.FormValue("back_sig")
	if cd.ImageSignKey != "" && sig != "" {
		data := "back:" + raw
		if err := sign.Verify(data, sig, cd.ImageSignKey, sign.WithOutNonce()); err == nil {
			return raw
		}
	}
	if sig != "" {
		log.Printf("invalid back signature url=%q sig=%s", raw, sig)
	} else {
		log.Printf("disallowed back host url=%q", raw)
	}
	return ""
}

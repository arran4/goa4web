package common

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/arran4/goa4web/internal/sign"
)

var (
	ErrInvalidBackURL          = errors.New("invalid back url")
	ErrProtocolRelativeBackURL = errors.New("protocol relative back url")
	ErrInvalidBackScheme       = errors.New("invalid back scheme")
	ErrInvalidBackSignature    = errors.New("invalid back signature")
	ErrDisallowedBackHost      = errors.New("disallowed back host")
)

// IsAllowedHost checks if a given hostname is considered allowed by comparing
// against the BaseURL configuration.
func (cd *CoreData) IsAllowedHost(host string) bool {
	if cd == nil || host == "" {
		return false
	}
	host = strings.ToLower(host)
	for h := range strings.FieldsSeq(cd.Config.BaseURL) {
		h = strings.TrimSpace(h)
		if pu, err := url.Parse(h); err == nil && pu.Host != "" {
			h = pu.Host
		} else {
			h = strings.TrimSuffix(h, "/")
		}
		if strings.ToLower(h) == host {
			return true
		}
	}
	return false
}

// SanitizeBackURL validates raw and returns a safe back URL.
// Absolute URLs are allowed only when the host matches an allowed hostname
// or when accompanied by a valid signature via back_ts and back_sig.
// Returns the sanitized URL string and nil on success.
// Returns an empty string and an error explaining why the URL was rejected on failure.
func (cd *CoreData) SanitizeBackURL(r *http.Request, raw string) (string, error) {
	if raw == "" {
		return "", nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		log.Printf("invalid back url %q: %v", raw, err)
		return "", fmt.Errorf("%w: %v", ErrInvalidBackURL, err)
	}
	if !u.IsAbs() {
		if u.Host != "" {
			log.Printf("invalid back host (protocol relative) %q", raw)
			return "", ErrProtocolRelativeBackURL
		}
		return raw, nil
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		log.Printf("invalid back scheme %q", raw)
		return "", fmt.Errorf("%w: %s", ErrInvalidBackScheme, u.Scheme)
	}

	host := strings.ToLower(u.Host)
	if (r != nil && r.Host != "" && strings.ToLower(r.Host) == host) || cd.IsAllowedHost(host) {
		result := u.Path
		if u.RawQuery != "" {
			result += "?" + u.RawQuery
		}
		if u.Fragment != "" {
			result += "#" + u.Fragment
		}
		return result, nil
	}

	sig := r.FormValue("back_sig")
	if cd.ImageSignKey != "" && sig != "" {
		data := "back:" + raw
		if err := sign.Verify(data, sig, cd.ImageSignKey, sign.WithOutNonce()); err == nil {
			return raw, nil
		}
	}
	if sig != "" {
		log.Printf("invalid back signature url=%q sig=%s", raw, sig)
		return "", ErrInvalidBackSignature
	} else {
		log.Printf("disallowed back host url=%q", raw)
		return "", fmt.Errorf("%w: %s", ErrDisallowedBackHost, u.Host)
	}
}

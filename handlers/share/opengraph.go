package share

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// OpenGraphData contains the metadata for an OpenGraph preview page.
type OpenGraphData struct {
	Title       string
	Description string
	ImageURL    template.URL
	ContentURL  template.URL
	ImageWidth  int
	ImageHeight int
	TwitterSite string
}

// RenderOpenGraph renders an OpenGraph preview page with the provided metadata.
func RenderOpenGraph(w http.ResponseWriter, r *http.Request, data OpenGraphData) error {
	tmpl, err := template.New("og").Parse(`
<!DOCTYPE html>
<html>
<head>
	<meta property="og:title" content="{{.Title}}" />
	<meta property="og:description" content="{{.Description}}" />
	{{.ImageMeta}}
	{{.SecureImageMeta}}
	{{.ImageWidthMeta}}
	{{.ImageHeightMeta}}
	{{.URLMeta}}
	<meta name="twitter:card" content="summary_large_image" />
	<meta name="twitter:title" content="{{.Title}}" />
	<meta name="twitter:description" content="{{.Description}}" />
	{{if .TwitterSite}}<meta name="twitter:site" content="{{.TwitterSite}}" />{{end}}
	{{.TwitterImageMeta}}
	<meta http-equiv="refresh" content="0;url={{.ContentURL}}" />
</head>
<body>
	<h1>Redirecting...</h1>
	<p>If you are not redirected automatically, <a href="{{.ContentURL}}">click here</a>.</p>
</body>
</html>
`)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.Execute(w, data)
}

func (d OpenGraphData) URLMeta() template.HTML {
	return template.HTML(fmt.Sprintf(`<meta property="og:url" content="%s" />`, d.ContentURL))
}

func (d OpenGraphData) ImageMeta() template.HTML {
	return template.HTML(fmt.Sprintf(`<meta property="og:image" content="%s" />`, d.ImageURL))
}

func (d OpenGraphData) SecureImageMeta() template.HTML {
	return template.HTML(fmt.Sprintf(`<meta property="og:image:secure_url" content="%s" />`, d.ImageURL))
}

func (d OpenGraphData) ImageWidthMeta() template.HTML {
	return template.HTML(fmt.Sprintf(`<meta property="og:image:width" content="%d" />`, d.ImageWidth))
}

func (d OpenGraphData) ImageHeightMeta() template.HTML {
	return template.HTML(fmt.Sprintf(`<meta property="og:image:height" content="%d" />`, d.ImageHeight))
}

func (d OpenGraphData) TwitterImageMeta() template.HTML {
	return template.HTML(fmt.Sprintf(`<meta name="twitter:image" content="%s" />`, d.ImageURL))
}

// VerifyAndGetPath verifies the signature and returns the content path without query parameters.
// Returns empty string if verification fails.
func VerifyAndGetPath(r *http.Request, signer SignatureVerifier) string {
	ts := r.URL.Query().Get("ts")
	sig := r.URL.Query().Get("sig")

	// Get path without query params
	path := r.URL.Path

	if ts == "" || sig == "" {
		vars := mux.Vars(r)
		ts = vars["ts"]
		sig = vars["sign"]

		if ts != "" && sig != "" {
			suffix := fmt.Sprintf("/ts/%s/sign/%s", ts, sig)
			path = strings.TrimSuffix(path, suffix)
		}
	}

	query := r.URL.Query()
	query.Del("ts")
	query.Del("sig")
	if encoded := query.Encode(); encoded != "" {
		path = path + "?" + encoded
	}

	log.Printf("Verifying signature. Path: %s, TS: %s, Sig: %s", path, ts, sig)

	if !signer.Verify(path, ts, sig) {
		log.Printf("Signature verification failed for path: %s", path)
		return ""
	}

	return path
}

// SignatureVerifier is an interface for signature verification.
type SignatureVerifier interface {
	Verify(data, ts, sig string) bool
}

// URLSigner is an interface for signing URLs.
type URLSigner interface {
	Sign(data string, exp ...time.Time) (int64, string)
}

// MakeImageURL creates an OpenGraph image URL for the given title with a specific expiration.
// If usePathAuth is true, it generates a URL with auth parameters in the path (/ts/.../sign/...).
// Expiration is optional. If not provided, the link will not expire.
func MakeImageURL(baseURL, title string, signer URLSigner, usePathAuth bool, expiration ...time.Time) string {
	encodedTitle := url.QueryEscape(title)
	path := fmt.Sprintf("/api/og-image?title=%s", encodedTitle)

	var exp time.Time
	if len(expiration) > 0 {
		exp = expiration[0]
	} else {
		// If no expiration provided, explicitly set to 0 (no expiry)
		// We pass time.Unix(0, 0) because Signer defaults to 24h if we pass nothing (empty slice).
		// Wait, Signer.Sign takes ...time.Time.
		// If we pass empty 'expiration' slice to Signer.Sign, it defaults to 24h.
		// We want NO expiry by default. So we MUST pass time.Unix(0, 0).
		exp = time.Unix(0, 0)
	}

	ts, sig := signer.Sign(path, exp)

	if usePathAuth {
		// Output: /api/og-image/ts/{ts}/sign/{sign}?title=...
		return fmt.Sprintf("%s/api/og-image/ts/%d/sign/%s?title=%s", baseURL, ts, sig, encodedTitle)
	}

	return fmt.Sprintf("%s%s&ts=%d&sig=%s", baseURL, path, ts, sig)
}

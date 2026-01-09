package share

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// OpenGraphData contains the metadata for an OpenGraph preview page.
type OpenGraphData struct {
	Title       string
	Description string
	ImageURL    template.URL
	ContentURL  template.URL
}

// RenderOpenGraph renders an OpenGraph preview page with the provided metadata.
func RenderOpenGraph(w http.ResponseWriter, r *http.Request, data OpenGraphData) error {
	tmpl, err := template.New("og").Parse(`
<!DOCTYPE html>
<html>
<head>
	<meta property="og:title" content="{{.Title}}" />
	<meta property="og:description" content="{{.Description}}" />
	<meta property="og:image" content="{{.ImageURL}}" />
	<meta property="og:image:secure_url" content="{{.ImageURL}}" />
	<meta property="og:url" content="{{.ContentURL}}" />
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

// VerifyAndGetPath verifies the signature and returns the content path without query parameters.
// Returns empty string if verification fails.
func VerifyAndGetPath(r *http.Request, signer SignatureVerifier) string {
	ts := r.URL.Query().Get("ts")
	sig := r.URL.Query().Get("sig")

	// Get path without query params
	path := r.URL.Path

	if !signer.Verify(path, ts, sig) {
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

// MakeImageURL creates an OpenGraph image URL for the given title.
func MakeImageURL(baseURL, title string, signer URLSigner) string {
	encodedTitle := strings.ReplaceAll(url.QueryEscape(title), "+", "%20")
	path := fmt.Sprintf("/api/og-image?title=%s", encodedTitle)
	ts, sig := signer.Sign(path, time.Now().Add(24*time.Hour))
	return fmt.Sprintf("%s%s&ts=%d&sig=%s", baseURL, path, ts, sig)
}

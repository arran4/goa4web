package share

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
)

// OpenGraphData contains the metadata for an OpenGraph preview page.
type OpenGraphData struct {
	Title       string
	Description string
	ImageURL    string
	ContentURL  string
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

// RenderLoginWithOG renders a login page with OpenGraph metadata.
// Social media bots can scrape the OG tags while users are prompted to login.
func RenderLoginWithOG(w http.ResponseWriter, r *http.Request, data OpenGraphData, redirectURL string) error {
	tmpl, err := template.New("og_login").Parse(`
<!DOCTYPE html>
<html>
<head>
	<meta property="og:title" content="{{.Title}}" />
	<meta property="og:description" content="{{.Description}}" />
	<meta property="og:image" content="{{.ImageURL}}" />
	<meta property="og:url" content="{{.ContentURL}}" />
	<title>{{.Title}} - Login Required</title>
	<style>
		body {
			font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
			display: flex;
			justify-content: center;
			align-items: center;
			min-height: 100vh;
			margin: 0;
			background: #f5f5f5;
		}
		.container {
			background: white;
			padding: 2rem;
			border-radius: 8px;
			box-shadow: 0 2px 10px rgba(0,0,0,0.1);
			max-width: 400px;
			text-align: center;
		}
		h1 {
			margin-top: 0;
			color: #333;
		}
		p {
			color: #666;
			line-height: 1.6;
		}
		a {
			display: inline-block;
			margin-top: 1rem;
			padding: 0.75rem 2rem;
			background: #007bff;
			color: white;
			text-decoration: none;
			border-radius: 4px;
			transition: background 0.2s;
		}
		a:hover {
			background: #0056b3;
		}
	</style>
</head>
<body>
	<div class="container">
		<h1>Login Required</h1>
		<p><strong>{{.Title}}</strong></p>
		<p>{{.Description}}</p>
		<p>Please login to view this content.</p>
		<a href="/login?redirect={{.RedirectURL}}">Login</a>
	</div>
</body>
</html>
`)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.Execute(w, struct {
		OpenGraphData
		RedirectURL string
	}{
		OpenGraphData: data,
		RedirectURL:   url.QueryEscape(redirectURL),
	})
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

// MakeImageURL creates an OpenGraph image URL for the given title.
func MakeImageURL(baseURL, title string) string {
	return fmt.Sprintf("%s/api/og-image?title=%s", baseURL, url.QueryEscape(title))
}

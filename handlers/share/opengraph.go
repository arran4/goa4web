package share

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/internal/sharesign"
	"github.com/arran4/goa4web/internal/sign"
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
func VerifyAndGetPath(r *http.Request, signer *sharesign.Signer) string {
	ts := r.URL.Query().Get("ts")
	sig := r.URL.Query().Get("sig")
	nonce := r.URL.Query().Get("nonce")

	// Get path without query params
	path := r.URL.Path

	if (ts == "" && nonce == "") || sig == "" {
		vars := mux.Vars(r)
		ts = vars["ts"]
		sig = vars["sign"]
		nonce = vars["nonce"]

		if ts != "" && sig != "" {
			suffix := fmt.Sprintf("/ts/%s/sign/%s", ts, sig)
			path = strings.TrimSuffix(path, suffix)
		} else if nonce != "" && sig != "" {
			suffix := fmt.Sprintf("/nonce/%s/sign/%s", nonce, sig)
			path = strings.TrimSuffix(path, suffix)
		}
	}

	query := r.URL.Query()
	query.Del("ts")
	query.Del("sig")
	query.Del("nonce")
	if encoded := query.Encode(); encoded != "" {
		path = path + "?" + encoded
	}

	log.Printf("Verifying signature. Path: %s, TS: %s, Nonce: %s, Sig: %s", path, ts, nonce, sig)

	var valid bool
	var err error
	if nonce != "" {
		valid, err = signer.Verify(path, sig, sign.WithNonce(nonce))
	} else if ts != "" {
		valid, err = signer.Verify(path, sig, sign.WithExpiryTimestamp(ts))
	} else {
		// No ts or nonce, assume no expiry/nonce was intended for verification
		valid, err = signer.Verify(path, sig)
	}

	if !valid {
		log.Printf("Signature verification failed for path: %s. Reason: %v", path, err)
		return ""
	}

	return path
}

func generateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// MakeImageURL creates an OpenGraph image URL for the given title with a specific expiration.
// If usePathAuth is true, it generates a URL with auth parameters in the path (/nonce/.../sign/...).
// Expiration is optional. If not provided, a nonce is used.
func MakeImageURL(baseURL, title string, signer *sharesign.Signer, usePathAuth bool, ops ...any) string {
	encodedData := base64.RawURLEncoding.EncodeToString([]byte(title))
	path := fmt.Sprintf("/api/og-image/%s", encodedData)

	// Generate nonce if no options provided
	var nonce string
	if len(ops) == 0 {
		nonce = generateNonce()
		ops = append(ops, sign.WithNonce(nonce))
	}

	sig := signer.Sign(path, ops...)

	log.Printf("Creating signature. Path: %s, Nonce: %s, Sig: %s", path, nonce, sig)

	if usePathAuth {
		// Output: /api/og-image/{base64}/nonce/{nonce}/sign/{sign}
		if nonce != "" {
			return fmt.Sprintf("%s%s/nonce/%s/sign/%s", baseURL, path, nonce, sig)
		}
		// If nonce is empty, ops must have provided timestamp or other auth
		return fmt.Sprintf("%s%s/sign/%s", baseURL, path, sig)
	}

	if nonce != "" {
		return fmt.Sprintf("%s%s?nonce=%s&sig=%s", baseURL, path, nonce, sig)
	}
	return fmt.Sprintf("%s%s?sig=%s", baseURL, path, sig)
}

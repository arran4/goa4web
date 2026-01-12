package share

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/sign"
	"github.com/arran4/goa4web/internal/sign/signutil"
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

// VerifyAndGetPath verifies the signature and returns the content path without auth parameters.
// Returns empty string if verification fails.
func VerifyAndGetPath(r *http.Request, key string) string {
	// Try extracting from query parameters first
	tsQuery := r.URL.Query().Get("ts")
	sigQuery := r.URL.Query().Get("sig")
	nonceQuery := r.URL.Query().Get("nonce")

	if sigQuery != "" {
		// Query-based auth
		_, sig, opts, err := sign.ExtractQuerySig(r.URL.String())
		if err != nil {
			log.Printf("extract query sig failed: %v", err)
			return ""
		}

		// Parse to get path + remaining query
		cleanPath := r.URL.Path
		q := r.URL.Query()
		q.Del("sig")
		q.Del("nonce")
		q.Del("ts")
		if encoded := q.Encode(); encoded != "" {
			cleanPath += "?" + encoded
		}

		log.Printf("Verifying query-based sig. Path: %s, Sig: %s, Opts: %v", cleanPath, sig, opts)

		if err := sign.Verify(cleanPath, sig, key, opts...); err != nil {
			log.Printf("verify failed: %v", err)
			return ""
		}

		return cleanPath
	}

	// Try path-based auth
	vars := mux.Vars(r)
	tspath := vars["ts"]
	sigPath := vars["sign"]
	if sigPath == "" {
		sigPath = vars["sig"]
	}
	noncePath := vars["nonce"]

	if sigPath != "" || tspath != "" || noncePath != "" {
		// Path-based auth
		q := r.URL.Query()
		q.Del("sig")
		q.Del("nonce")
		q.Del("ts")
		additionalQuery := q.Encode()

		cleanPath, sig, opts, err := sign.ExtractPathSig(r.URL.Path, vars)
		if err != nil {
			log.Printf("extract path sig failed: %v", err)
			return ""
		}

		data := cleanPath
		if additionalQuery != "" {
			data += "?" + additionalQuery
		}

		log.Printf("Verifying path-based sig. Data: %s, Sig: %s, Opts: %v", data, sig, opts)

		if err := sign.Verify(data, sig, key, opts...); err != nil {
			log.Printf("verify failed: %v", err)
			return ""
		}

		return data
	}

	// No sig found
	log.Printf("No signature found in request. Query params: ts=%s, sig=%s, nonce=%s. Path vars: ts=%s, sig=%s, nonce=%s",
		tsQuery, sigQuery, nonceQuery, tspath, sigPath, noncePath)
	return ""
}

// Make ImageURL creates an OpenGraph image URL for the given title.
// By default generates a nonce-based signature.
// Pass usePathAuth=true for path-based signatures, false for query-based.
func MakeImageURL(baseURL, title, key string, usePathAuth bool, opts ...sign.SignOption) (string, error) {
	encodedData := base64.RawURLEncoding.EncodeToString([]byte(title))
	path := "/api/og-image/" + encodedData

	// Generate nonce if no options provided
	var nonce string
	if len(opts) == 0 {
		nonce = signutil.GenerateNonce()
		opts = append(opts, sign.WithNonce(nonce))
	} else {
		// Check if nonce is in opts
		for _, opt := range opts {
			if n, ok := opt.(sign.WithNonce); ok {
				nonce = string(n)
				break
			}
		}
		// If no nonce and no expiry, add nonce
		if nonce == "" {
			hasExpiry := false
			for _, opt := range opts {
				if _, ok := opt.(sign.WithExpiry); ok {
					hasExpiry = true
					break
				}
			}
			if !hasExpiry {
				nonce = signutil.GenerateNonce()
				opts = append(opts, sign.WithNonce(nonce))
			}
		}
	}

	fullURL := baseURL + path

	log.Printf("Making image URL. Path: %s, Nonce: %s, UsePathAuth: %v", path, nonce, usePathAuth)

	if usePathAuth {
		return signutil.SignAndAddPath(fullURL, path, key, opts...)
	}
	return signutil.SignAndAddQuery(fullURL, path, key, opts...)
}

// OGImageHandler serves dynamically generated OpenGraph images.
type OGImageHandler struct {
	key string
}

// NewOGImageHandler creates a new OpenGraph image handler.
func NewOGImageHandler(key string) *OGImageHandler {
	return &OGImageHandler{key: key}
}

func (h *OGImageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Verify signature
	cleanPath := VerifyAndGetPath(r, h.key)
	if cleanPath == "" {
		http.Error(w, "Invalid or missing signature", http.StatusForbidden)
		return
	}

	// Extract base64 encoded title from path
	parts := strings.Split(cleanPath, "/")
	var encodedTitle string
	for i, part := range parts {
		if i > 0 && parts[i-1] == "og-image" {
			encodedTitle = part
			break
		}
	}

	if encodedTitle == "" {
		http.Error(w, "Invalid path format", http.StatusBadRequest)
		return
	}

	titleBytes, err := base64.RawURLEncoding.DecodeString(encodedTitle)
	if err != nil {
		http.Error(w, "Invalid title encoding", http.StatusBadRequest)
		return
	}

	title := string(titleBytes)

	// Generate a simple SVG image
	svg := fmt.Sprintf(`<svg width="1200" height="630" xmlns="http://www.w3.org/2000/svg">
		<rect width="1200" height="630" fill="#282c34"/>
		<text x="600" y="315" font-family="Arial" font-size="48" fill="white" text-anchor="middle" dominant-baseline="middle">%s</text>
	</svg>`, template.HTMLEscapeString(title))

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
	w.Write([]byte(svg))
}

// SharedContentPreview generates a signed OpenGraph preview URL.
// This is meant to be called from module handlers to generate share links.
func SharedContentPreview(cd *common.CoreData, contentPath, title, description string) (string, error) {
	// Use CoreData's helper method
	return cd.SignShareURL(contentPath)
}

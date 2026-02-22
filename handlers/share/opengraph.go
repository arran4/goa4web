package share

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"image"
	"image/png"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/sign"
	"github.com/arran4/goa4web/internal/sign/signutil"
	"github.com/gorilla/mux"
)

var (
	generatorsMu sync.RWMutex
	generators   = make(map[string]ImageGenerator)
)

func RegisterGenerator(g ImageGenerator) {
	generatorsMu.Lock()
	defer generatorsMu.Unlock()
	generators[g.Name()] = g
}

func init() {
	RegisterGenerator(&DefaultGenerator{})
	// Alias for backward compatibility
	RegisterGenerator(&AliasGenerator{
		Target:        "sierpinski",
		Alias:         "default",
		RealGenerator: &DefaultGenerator{},
	})
	RegisterGenerator(&ForumGenerator{})
}

type AliasGenerator struct {
	Target        string
	Alias         string
	RealGenerator ImageGenerator
}

func (a *AliasGenerator) Name() string { return a.Alias }
func (a *AliasGenerator) Generate(options ...interface{}) (image.Image, error) {
	return a.RealGenerator.Generate(options...)
}

func getGenerator(name string) (ImageGenerator, bool) {
	generatorsMu.RLock()
	defer generatorsMu.RUnlock()
	g, ok := generators[name]
	return g, ok
}

// Generate generates an image using the generator specified in options (or default).
func Generate(options ...interface{}) (image.Image, error) {
	genType := "default"
	for _, opt := range options {
		if v, ok := opt.(WithGeneratorType); ok {
			genType = string(v)
		}
	}

	gen, ok := getGenerator(genType)
	if !ok {
		// Fallback to default
		gen, ok = getGenerator("default")
		if !ok {
			return nil, fmt.Errorf("generator not found: %s and default missing", genType)
		}
	}
	return gen.Generate(options...)
}

// OpenGraphData contains the metadata for an OpenGraph preview page.
type OpenGraphData struct {
	Title       string
	Description string
	ImageURL    template.URL
	ContentURL  template.URL
	RedirectURL template.URL
	ImageWidth  int
	ImageHeight int
	TwitterSite string
	JSONLD      interface{}
}

// RenderOpenGraph renders an OpenGraph preview page with the provided metadata.
func RenderOpenGraph(w http.ResponseWriter, r *http.Request, data OpenGraphData) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return templates.GetCompiledSiteTemplates(nil).ExecuteTemplate(w, "openGraphPreview.gohtml", data)
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

func (d OpenGraphData) JSONLDScript() template.HTML {
	if d.JSONLD == nil {
		return ""
	}
	b, err := json.Marshal(d.JSONLD)
	if err != nil {
		log.Printf("Error marshaling JSONLD: %v", err)
		return ""
	}
	return template.HTML(fmt.Sprintf(`<script type="application/ld+json">%s</script>`, string(b)))
}

// VerifyAndGetPath verifies the signature and returns the content path without auth parameters.
// Returns empty string if verification fails.
func VerifyAndGetPath(r *http.Request, key string) string {
	// Try extracting from query parameters first
	if r.URL.Query().Get("sig") != "" {
		// Query-based auth
		_, sig, opts, err := sign.ExtractQuerySig(r.URL.String())
		if err != nil {
			log.Printf("extract query sig failed: %v", err)
			return ""
		}

		// Check mux variables for mixed auth (path-based temporal params, query-based signature)
		vars := mux.Vars(r)
		tsPath := vars["ts"]
		noncePath := vars["nonce"]

		cleanPath := r.URL.Path
		if tsPath != "" {
			if ts, err := strconv.ParseInt(tsPath, 10, 64); err == nil {
				opts = append(opts, sign.WithExpiry(time.Unix(ts, 0)))
				cleanPath = strings.Replace(cleanPath, "/ts/"+tsPath, "", 1)
			}
		}
		if noncePath != "" {
			opts = append(opts, sign.WithNonce(noncePath))
			cleanPath = strings.Replace(cleanPath, "/nonce/"+noncePath, "", 1)
		}

		// Re-encode query without signature params for verification data
		q := r.URL.Query()
		q.Del("sig")
		q.Del("nonce")
		q.Del("ts")
		q.Del("ets")
		q.Del("its")

		data := cleanPath
		if encoded := q.Encode(); encoded != "" {
			data += "?" + encoded
		}

		log.Printf("Verifying query-based sig. Path: %s, Sig: %s, Opts: %v", data, sig, opts)

		if err := sign.Verify(data, sig, key, opts...); err != nil {
			log.Printf("verify failed: %v", err)
			return ""
		}

		return data
	}

	// Try path-based auth
	vars := mux.Vars(r)
	if vars["sign"] != "" || vars["sig"] != "" || vars["ts"] != "" || vars["nonce"] != "" {
		// Path-based auth
		cleanPath, sig, opts, err := sign.ExtractPathSig(r.URL.Path, vars)
		if err != nil {
			log.Printf("extract path sig failed: %v", err)
			return ""
		}

		q := r.URL.Query()
		q.Del("sig")
		q.Del("nonce")
		q.Del("ts")
		q.Del("ets")
		q.Del("its")

		data := cleanPath
		if encoded := q.Encode(); encoded != "" {
			data += "?" + encoded
		}

		log.Printf("Verifying path-based sig. Data: %s, Sig: %s, Opts: %v", data, sig, opts)

		if err := sign.Verify(data, sig, key, opts...); err != nil {
			log.Printf("verify failed: %v", err)
			return ""
		}

		return data
	}

	// No sig found
	log.Printf("No signature found in request. Path: %s, Vars: %v", r.URL.Path, vars)
	return ""
}

// MakeImageURL creates an OpenGraph image URL for the given title and description.
// By default generates a nonce-based signature.
// Pass usePathAuth=true for path-based signatures, false for query-based.
func MakeImageURL(baseURL, title, description, key string, usePathAuth bool, opts ...sign.SignOption) (string, error) {
	payload := imagePayload{
		Title:       title,
		Description: description,
		Type:        "default",
	}

	return makeImageURLFromPayload(baseURL, payload, key, usePathAuth, opts...)
}

// MakeImageURLWithOptions allows creating an image URL with specific generator options.
func MakeImageURLWithOptions(baseURL, key string, usePathAuth bool, options ...interface{}) (string, error) {
	payload := imagePayload{
		Type: "default",
	}

	var signOpts []sign.SignOption

	// Pack options into payload or sign options
	for _, opt := range options {
		switch v := opt.(type) {
		case WithGeneratorType:
			payload.Type = string(v)
		case WithTitle:
			payload.Title = string(v)
		case WithDescription:
			payload.Description = string(v)
		case WithSection:
			payload.Section = string(v)
		case WithAuthor:
			payload.Author = string(v)
		case WithHeader:
			payload.Header = string(v)
		case WithBody:
			payload.Body = string(v)
		// Sign Options
		case sign.SignOption:
			signOpts = append(signOpts, v)
		}
	}

	return makeImageURLFromPayload(baseURL, payload, key, usePathAuth, signOpts...)
}

func makeImageURLFromPayload(baseURL string, payload imagePayload, key string, usePathAuth bool, opts ...sign.SignOption) (string, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	encodedData := base64.RawURLEncoding.EncodeToString(data)
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

type imagePayload struct {
	Type        string `json:"y,omitempty"` // Generator Type
	Title       string `json:"t"`
	Description string `json:"d,omitempty"`
	Section     string `json:"s,omitempty"`
	Author      string `json:"a,omitempty"`
	Header      string `json:"h,omitempty"`
	Body        string `json:"b,omitempty"`
}

// OGImageHandler serves dynamically generated OpenGraph images.
type OGImageHandler struct {
	signKey string
}

// NewOGImageHandler creates a new OpenGraph image handler.
func NewOGImageHandler(signKey string) *OGImageHandler {
	return &OGImageHandler{
		signKey: signKey,
	}
}

func (h *OGImageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dataB64, ok := vars["data"]
	if !ok {
		log.Printf("data not found")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dataBytes, decodeErr := base64.RawURLEncoding.DecodeString(dataB64)

	// Try unmarshal as JSON
	var payload imagePayload
	if decodeErr == nil {
		if err := json.Unmarshal(dataBytes, &payload); err != nil {
			// Fallback for legacy URLs: treat entire data as title
			payload.Title = string(dataBytes)
			payload.Type = "default"
		}
	}

	signed, err := signutil.GetSignedData(r, h.signKey)
	if err != nil {
		log.Printf("Error getting signed data: %v", err)
		if decodeErr == nil {
			log.Printf("Request Details: Title: %q, Description: %q", payload.Title, payload.Description)
		}
		log.Printf("Request Context: IP: %s, UserAgent: %s", r.RemoteAddr, r.UserAgent())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if !signed.Valid {
		log.Printf("Invalid signature")
		if decodeErr == nil {
			log.Printf("Request Details: Title: %q, Description: %q", payload.Title, payload.Description)
		}
		log.Printf("Request Context: IP: %s, UserAgent: %s", r.RemoteAddr, r.UserAgent())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if decodeErr != nil {
		log.Printf("Error decoding data: %v", decodeErr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Construct options from payload
	var options []interface{}
	// Add generator type
	if payload.Type != "" {
		options = append(options, WithGeneratorType(payload.Type))
	} else {
		options = append(options, WithGeneratorType("default"))
	}

	if payload.Title != "" {
		options = append(options, WithTitle(payload.Title))
	}
	if payload.Description != "" {
		options = append(options, WithDescription(payload.Description))
	}
	if payload.Section != "" {
		options = append(options, WithSection(payload.Section))
	}
	if payload.Author != "" {
		options = append(options, WithAuthor(payload.Author))
	}
	if payload.Header != "" {
		options = append(options, WithHeader(payload.Header))
	}
	if payload.Body != "" {
		options = append(options, WithBody(payload.Body))
	}

	img, err := Generate(options...)
	if err != nil {
		log.Printf("Error generating image: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	if err := png.Encode(w, img); err != nil {
		log.Printf("Error encoding png: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// SharedContentPreview generates a signed OpenGraph preview URL.
// This is meant to be called from module handlers to generate share links.
func SharedContentPreview(cd *common.CoreData, contentPath, title, description string) (string, error) {
	// Use CoreData's helper method
	return cd.SignShareURL(contentPath)
}

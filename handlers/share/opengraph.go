package share

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/arran4/go-pattern"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/sign"
	"github.com/arran4/goa4web/internal/sign/signutil"
	"github.com/gorilla/mux"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

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
	{{if .RedirectURL}}<meta http-equiv="refresh" content="0;url={{.RedirectURL}}" />{{else}}<meta http-equiv="refresh" content="0;url={{.ContentURL}}" />{{end}}
</head>
<body>
	<h1>Redirecting...</h1>
	<p>If you are not redirected automatically, <a href="{{if .RedirectURL}}{{.RedirectURL}}{{else}}{{.ContentURL}}{{end}}">click here</a>.</p>
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

	vars := mux.Vars(r)

	if sigQuery != "" {
		// Query-based auth
		_, sig, opts, err := sign.ExtractQuerySig(r.URL.String())
		if err != nil {
			log.Printf("extract query sig failed: %v", err)
			return ""
		}

		// Parse to get path + remaining query
		cleanPath := r.URL.Path

		tsPath := vars["ts"]
		noncePath := vars["nonce"]

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

// MakeImageURL creates an OpenGraph image URL for the given title and description.
// By default generates a nonce-based signature.
// Pass usePathAuth=true for path-based signatures, false for query-based.
func MakeImageURL(baseURL, title, description, key string, usePathAuth bool, opts ...sign.SignOption) (string, error) {
	payload := imagePayload{
		Title:       title,
		Description: description,
	}
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
	Title       string `json:"t"`
	Description string `json:"d,omitempty"`
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
	signed, err := signutil.GetSignedData(r, h.signKey)
	if err != nil {
		log.Printf("Error getting signed data: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if !signed.Valid {
		log.Printf("Invalid signature")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	dataBytes, err := base64.RawURLEncoding.DecodeString(dataB64)
	if err != nil {
		log.Printf("Error decoding data: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Try unmarshal as JSON
	var payload imagePayload
	if err := json.Unmarshal(dataBytes, &payload); err != nil {
		// Fallback for legacy URLs: treat entire data as title
		payload.Title = string(dataBytes)
	}

	img, err := GenerateImage(payload.Title, payload.Description)
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

func GenerateImage(title, description string) (image.Image, error) {
	img := image.NewRGBA(image.Rect(0, 0, 1200, 630))

	// Create Sierpinski Triangle pattern for background
	st := &pattern.SierpinskiTriangle{}
	st.SetBounds(img.Bounds())
	st.SetFillColor(color.RGBA{R: 0x0b, G: 0x35, B: 0x13, A: 0xff})  // Dark green
	st.SetSpaceColor(color.RGBA{R: 0x1a, G: 0x5e, B: 0x27, A: 0xff}) // Lighter green
	draw.Draw(img, img.Bounds(), st, image.Point{}, draw.Src)

	f, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return nil, fmt.Errorf("error parsing font: %w", err)
	}
	// Title Face
	titleFace, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    64,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating title font face: %w", err)
	}

	// Description Face
	descFace, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    40,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating desc font face: %w", err)
	}

	logoBytes, err := templates.Asset("favicon.png")
	if err != nil {
		return nil, fmt.Errorf("error getting logo: %w", err)
	}
	logo, _, err := image.Decode(bytes.NewReader(logoBytes))
	if err != nil {
		return nil, fmt.Errorf("error decoding logo: %w", err)
	}
	drawImage := img
	// draw logo centered
	logoBounds := logo.Bounds()
	logoPt := image.Point{
		X: (1200 - logoBounds.Dx()) / 2,
		Y: 50,
	}
	draw.Draw(drawImage, logo.Bounds().Add(logoPt), logo, image.Point{}, draw.Over)

	// draw text centered
	d := &font.Drawer{
		Dst:  drawImage,
		Src:  image.NewUniform(color.White),
		Face: titleFace,
		Dot:  fixed.Point26_6{},
	}

	// Draw Title
	// Basic wrapping for Title if it's too long?
	// For now, let's assume title fits on one line or accept truncation for simplicity unless requested explicitly.
	// Actually, let's position it higher.
	textWidth := d.MeasureString(title)
	d.Dot.X = (fixed.I(1200) - textWidth) / 2
	d.Dot.Y = fixed.I(300)
	d.DrawString(title)

	// Draw Description (Multi-line)
	if description != "" {
		// Try to parse description as a4code to strip tags
		if root, err := a4code.ParseString(description); err == nil {
			description = a4code.ToText(root)
		} else {
			// Fallback: minimal cleanup or just usage as is
		}

		d.Face = descFace
		lines := strings.Split(description, "\n")
		// Filter empty lines?? No, respect them.

		startY := 400
		lineHeight := 50

		for i, line := range lines {
			if startY+(i*lineHeight) > 600 {
				break // Stop if out of bounds
			}
			// Snip if too long
			if len(line) > 60 {
				line = line[:57] + "..."
			}

			w := d.MeasureString(line)
			d.Dot.X = (fixed.I(1200) - w) / 2
			d.Dot.Y = fixed.I(startY + (i * lineHeight))
			d.DrawString(line)
		}
	}

	return img, nil
}

// SharedContentPreview generates a signed OpenGraph preview URL.
// This is meant to be called from module handlers to generate share links.
func SharedContentPreview(cd *common.CoreData, contentPath, title, description string) (string, error) {
	// Use CoreData's helper method
	return cd.SignShareURL(contentPath)
}

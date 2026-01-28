package common

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log"

	"github.com/arran4/goa4web/internal/sign"
	"github.com/arran4/goa4web/internal/sign/signutil"
)

func (og *OpenGraph) URLMeta() template.HTML {
	return template.HTML(fmt.Sprintf(`<meta property="og:url" content="%s" />`, og.URL))
}

func (og *OpenGraph) ImageMeta() template.HTML {
	return template.HTML(fmt.Sprintf(`<meta property="og:image" content="%s" />`, og.Image))
}

func (og *OpenGraph) SecureImageMeta() template.HTML {
	return template.HTML(fmt.Sprintf(`<meta property="og:image:secure_url" content="%s" />`, og.Image))
}

func (og *OpenGraph) ImageWidthMeta() template.HTML {
	if og.ImageWidth == 0 {
		return ""
	}
	return template.HTML(fmt.Sprintf(`<meta property="og:image:width" content="%d" />`, og.ImageWidth))
}

func (og *OpenGraph) ImageHeightMeta() template.HTML {
	if og.ImageHeight == 0 {
		return ""
	}
	return template.HTML(fmt.Sprintf(`<meta property="og:image:height" content="%d" />`, og.ImageHeight))
}

func (og *OpenGraph) TwitterImageMeta() template.HTML {
	return template.HTML(fmt.Sprintf(`<meta name="twitter:image" content="%s" />`, og.Image))
}

func (og *OpenGraph) TypeMeta() template.HTML {
	ogType := "website"
	if og.Type != "" {
		ogType = og.Type
	}
	return template.HTML(fmt.Sprintf(`<meta property="og:type" content="%s" />`, ogType))
}

type ImagePayload struct {
	Title       string `json:"t"`
	Description string `json:"d,omitempty"`
}

// MakeImageURL creates an OpenGraph image URL for the given title and description.
// It supports variadic options to configure signing and path behavior.
// Options can be:
// - string: the signing key
// - bool: true to use path-based auth, false (default) for query-based
// - sign.SignOption: options for the signature generation (nonce, expiry)
func MakeImageURL(baseURL, title, description string, ops ...any) (string, error) {
	payload := ImagePayload{
		Title:       title,
		Description: description,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	encodedData := base64.RawURLEncoding.EncodeToString(data)
	path := "/api/og-image/" + encodedData

	var key string
	var usePathAuth bool
	var opts []sign.SignOption

	for _, op := range ops {
		switch v := op.(type) {
		case string:
			key = v
		case bool:
			usePathAuth = v
		case sign.SignOption:
			opts = append(opts, v)
		}
	}

	fullURL := baseURL + path
	if key == "" {
		return fullURL, nil
	}

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

	log.Printf("Making image URL. Path: %s, Nonce: %s, UsePathAuth: %v", path, nonce, usePathAuth)

	if usePathAuth {
		return signutil.SignAndAddPath(fullURL, path, key, opts...)
	}
	return signutil.SignAndAddQuery(fullURL, path, key, opts...)
}

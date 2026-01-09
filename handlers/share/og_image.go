package share

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/templates"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type OGImageHandler struct {
	signer SignatureVerifier
	config *config.RuntimeConfig
}

func NewOGImageHandler(signer SignatureVerifier, cfg *config.RuntimeConfig) *OGImageHandler {
	return &OGImageHandler{signer: signer, config: cfg}
}

func (h *OGImageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	if title == "" {
		title = "Shared Content"
	}
	style := r.URL.Query().Get("style")
	if style == "" {
		style = "checker"
	}

	ts := r.URL.Query().Get("ts")
	sig := r.URL.Query().Get("sig")

	path := fmt.Sprintf("/api/og-image?title=%s", strings.ReplaceAll(url.QueryEscape(title), "+", "%20"))
	if !h.signer.Verify(path, ts, sig) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	width := h.config.OGImageWidth
	height := h.config.OGImageHeight
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Background
	switch style {
	case "checker":
		drawChecker(img, 50, 50, color.White, color.RGBA{240, 240, 240, 255})
	default:
		draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	}

	// Load Font
	f, err := opentype.Parse(goregular.TTF)
	if err != nil {
		log.Printf("Font parse error: %v", err)
		drawBasic(img, title)
		writeImage(w, r, img)
		return
	}
	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    64,
		DPI:     72,
		Hinting: font.HintingNone,
	})
	if err != nil {
		log.Printf("Font face error: %v", err)
		drawBasic(img, title)
		writeImage(w, r, img)
		return
	}
	defer face.Close()

	// Draw Favicon/Logo
	var offsetY int
	favBytes := templates.GetFaviconPNG()
	if len(favBytes) > 0 {
		favImg, _, err := image.Decode(bytes.NewReader(favBytes))
		if err == nil {
			b := favImg.Bounds()
			width := b.Dx()
			height := b.Dy()

			// Scale up if too small (make it at least 64px)
			// But for now just draw centering.
			x := (width - b.Dx()) / 2
			y := 150
			draw.Draw(img, image.Rect(x, y, x+width, y+height), favImg, image.Point{}, draw.Over)
			offsetY = y + height + 50
		} else {
			// Try SVG logic? No, assuming PNG.
			offsetY = 250
		}
	} else {
		offsetY = 250
	}

	// Draw Title
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.Black),
		Face: face,
	}

	wStr := d.MeasureString(title)
	xStr := (width - wStr.Ceil()) / 2
	if xStr < 60 {
		xStr = 60
	}
	y := offsetY + 64 // baseline
	if y > 550 {
		y = 550
	}

	d.Dot = fixed.Point26_6{X: fixed.I(xStr), Y: fixed.I(y)}
	d.DrawString(title)

	writeImage(w, r, img)
}

func drawChecker(img draw.Image, w, h int, c1, c2 color.Color) {
	b := img.Bounds()
	u1 := &image.Uniform{c1}
	u2 := &image.Uniform{c2}
	for y := b.Min.Y; y < b.Max.Y; y += h {
		for x := b.Min.X; x < b.Max.X; x += w {
			r := image.Rect(x, y, x+w, y+h).Intersect(b)
			// (x/w + y/h) % 2 == 0 ?
			// We need logic relative to origin 0,0 usually.
			// assuming grid aligns with pixels.
			isEven := ((x/w)+(y/h))%2 == 0
			if isEven {
				draw.Draw(img, r, u1, image.Point{}, draw.Src)
			} else {
				draw.Draw(img, r, u2, image.Point{}, draw.Src)
			}
		}
	}
}

func drawBasic(img draw.Image, title string) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.Black),
		Face: basicfont.Face7x13,
		Dot:  fixed.Point26_6{X: fixed.Int26_6(10 * 64), Y: fixed.Int26_6(300 * 64)},
	}
	d.DrawString(title)
}

func writeImage(w http.ResponseWriter, r *http.Request, img image.Image) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		log.Printf("Error encoding png: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	w.Header().Set("Cache-Control", "public, max-age=86400")

	if r.Method == http.MethodHead {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Write(buf.Bytes())
}

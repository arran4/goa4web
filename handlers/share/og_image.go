package share

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"net/http"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func OGImageHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	if title == "" {
		title = "Shared Content"
	}

	img := image.NewRGBA(image.Rect(0, 0, 1200, 630))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.Black),
		Face: basicfont.Face7x13,
		Dot:  fixed.Point26_6{X: fixed.Int26_6(10 * 64), Y: fixed.Int26_6(300 * 64)},
	}
	d.DrawString(title)

	w.Header().Set("Content-Type", "image/png")
	png.Encode(w, img)
}

package share

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"strings"

	"github.com/arran4/go-pattern"
	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/core/templates"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type DefaultGenerator struct{}

func (g *DefaultGenerator) Name() string {
	return "sierpinski"
}

func (g *DefaultGenerator) Generate(options ...interface{}) (image.Image, error) {
	var title, description string
	for _, opt := range options {
		switch v := opt.(type) {
		case WithTitle:
			title = string(v)
		case WithDescription:
			description = string(v)
		}
	}
	return GenerateSierpinskiImage(title, description)
}

// GenerateSierpinskiImage is the original logic extracted
func GenerateSierpinskiImage(title, description string) (image.Image, error) {
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
	textWidth := d.MeasureString(title)
	d.Dot.X = (fixed.I(1200) - textWidth) / 2
	d.Dot.Y = fixed.I(300)
	d.DrawString(title)

	// Draw Description (Multi-line)
	if description != "" {
		if root, err := a4code.ParseString(description); err == nil {
			description = a4code.ToText(root)
		}

		d.Face = descFace
		lines := strings.Split(description, "\n")

		startY := 400
		lineHeight := 50

		for i, line := range lines {
			if startY+(i*lineHeight) > 600 {
				break
			}
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

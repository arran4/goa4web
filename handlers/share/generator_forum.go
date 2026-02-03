package share

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"strings"

	"github.com/arran4/go-pattern"
	"github.com/arran4/goa4web/core/templates"
	wordwrap "github.com/arran4/golang-wordwrap"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type ForumGenerator struct{}

func (g *ForumGenerator) Name() string {
	return "forum"
}

func (g *ForumGenerator) Generate(options ...interface{}) (image.Image, error) {
	var title, body, section string
	var avatar image.Image
	for _, opt := range options {
		switch v := opt.(type) {
		case WithTitle:
			title = string(v)
		case WithBody:
			body = string(v)
		case WithSection:
			section = string(v)
		case WithAvatar:
			avatar = image.Image(v)
		}
	}

	img := image.NewRGBA(image.Rect(0, 0, 1200, 630))

	// Background: Dark Blue-ish for forum
	st := &pattern.SierpinskiTriangle{}
	st.SetBounds(img.Bounds())
	st.SetFillColor(color.RGBA{R: 0x0b, G: 0x35, B: 0x13, A: 0xff})  // Dark green
	st.SetSpaceColor(color.RGBA{R: 0x1a, G: 0x5e, B: 0x27, A: 0xff}) // Lighter green
	draw.Draw(img, img.Bounds(), st, image.Point{}, draw.Src)

	f, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return nil, fmt.Errorf("error parsing font: %w", err)
	}

	// Section Label Face
	sectionFace, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    30,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating section font face: %w", err)
	}

	// Title Face
	titleFace, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    54,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating title font face: %w", err)
	}

	// Body Face
	bodyFace, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    36,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating body font face: %w", err)
	}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.White),
		Face: sectionFace,
		Dot:  fixed.Point26_6{},
	}

	padding := 50

	// Draw Section
	if section != "" {
		d.Dot.X = fixed.I(padding)
		d.Dot.Y = fixed.I(padding + 30)
		d.DrawString(strings.ToUpper(section))
	}

	// Draw Logo
	logoBytes, err := templates.Asset("favicon.png")
	if err == nil {
		if logo, _, err := image.Decode(bytes.NewReader(logoBytes)); err == nil {
			logoWidth := logo.Bounds().Dx()
			logoPt := image.Point{
				X: 1200 - padding - logoWidth,
				Y: padding,
			}
			draw.Draw(img, logo.Bounds().Add(logoPt), logo, image.Point{}, draw.Over)
		}
	}

	currentY := 150

	// Draw Avatar if present
	if avatar != nil {
		avatarRect := image.Rect(0, 0, 100, 100)
		avatarPt := image.Point{X: padding, Y: currentY}
		draw.Draw(img, avatarRect.Add(avatarPt), avatar, image.Point{}, draw.Over)

		d.Face = titleFace
		d.Dot.X = fixed.I(padding + 120)
		d.Dot.Y = fixed.I(currentY + 60)
		d.DrawString(title)

		currentY += 120
	} else {
		// Just Title
		d.Face = titleFace
		d.Dot.X = fixed.I(padding)
		d.Dot.Y = fixed.I(currentY + 50)
		d.DrawString(title)
		currentY += 120
	}

	// Draw Body (Wrapped)
	if body != "" {
		bodyRect := image.Rect(padding, currentY, 1200-padding, 630-padding)

		sw := wordwrap.NewSimpleWrapper([]*wordwrap.Content{wordwrap.NewContent(body, wordwrap.WithFontColor(color.White))}, bodyFace)
		lines, _, err := sw.TextToRect(bodyRect)
		if err == nil {
			sw.RenderLines(img, lines, bodyRect.Min)
		}
	}

	return img, nil
}

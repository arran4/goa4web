package main

import (
	"bytes"
	"flag"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"

	"github.com/arran4/go-pattern"
	"github.com/arran4/goa4web/core/templates"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type genOgImageCmd struct {
	fs *flag.FlagSet
	Title string
	OutputFile string
}

func (c *genOgImageCmd) Name() string {
	return c.fs.Name()
}

func (c *genOgImageCmd) Run() error {
	return GenerateOgImage(c.Title, c.OutputFile)
}

func GenerateOgImage(title, outputFile string) error {
	templates.SetDir("core/templates")
	width, height := 1200, 630
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	p := pattern.NewSierpinskiTriangle(
		pattern.SetBounds(image.Rect(0, 0, width, height)),
		pattern.SetFillColor(color.RGBA{R: 60, G: 66, B: 78, A: 255}),
		pattern.SetSpaceColor(color.RGBA{R: 40, G: 44, B: 52, A: 255}))

	draw.Draw(img, img.Bounds(), p, image.Point{}, draw.Src)

	logoData := templates.GetFaviconPNG()
	logo, _, err := image.Decode(bytes.NewReader(logoData))
	if err != nil {
		return err
	}
	logoBounds := logo.Bounds()
	draw.Draw(img, image.Rect(20, 20, 20+logoBounds.Dx(), 20+logoBounds.Dy()), logo, image.Point{}, draw.Src)

	fontData, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return err
	}
	face, err := opentype.NewFace(fontData, &opentype.FaceOptions{
		Size: 48,
		DPI: 72,
	})
	if err != nil {
		return err
	}

	d := &font.Drawer{
		Dst: img,
		Src: image.NewUniform(color.White),
		Face: face,
		Dot: fixed.Point26_6{X: fixed.I(width / 2), Y: fixed.I(height / 2)},
	}
	bounds, _ := font.BoundString(face, title)
	d.Dot.X -= bounds.Max.X / 2
	d.DrawString(title)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return err
	}

	return ioutil.WriteFile(outputFile, buf.Bytes(), 0644)
}

func newGenOgImageCmd(r *rootCmd, args []string) (*genOgImageCmd, error) {
	c := &genOgImageCmd{
		fs: flag.NewFlagSet("gen-og-image", flag.ContinueOnError),
	}
	c.fs.StringVar(&c.Title, "title", "GoA4Web", "The title to use in the image")
	c.fs.StringVar(&c.OutputFile, "output", "og-image.png", "The output file")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

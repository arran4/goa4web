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
	"fmt"
	"io/ioutil"

	"github.com/arran4/goa4web/handlers/share"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"fmt"
	"io/ioutil"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type genOgImageCmd struct {
	fs *flag.FlagSet
	Title string
	OutputFile string
	Pattern string
	FgColor string
	BgColor string
	RpgTheme bool
}

func (c *genOgImageCmd) Name() string {
	return c.fs.Name()
}

func (c *genOgImageCmd) Run() error {
	fg, err := share.ParseHexColor(c.FgColor)
	if err != nil {
		return err
	}
	bg, err := share.ParseHexColor(c.BgColor)
	if err != nil {
		return err
	}
	buf, err := share.GenerateOgImage(c.Title, c.Pattern, fg, bg, c.RpgTheme)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(c.OutputFile, buf.Bytes(), 0644)
}



func newGenOgImageCmd(r *rootCmd, args []string) (*genOgImageCmd, error) {
	c := &genOgImageCmd{
		fs: flag.NewFlagSet("gen-og-image", flag.ContinueOnError),
	}
	c.fs.StringVar(&c.Title, "title", "GoA4Web", "The title to use in the image")
	c.fs.StringVar(&c.OutputFile, "output", "og-image.png", "The output file")
	c.fs.StringVar(&c.Pattern, "pattern", "SierpinskiTriangle", "The pattern style to use")
	c.fs.StringVar(&c.FgColor, "fg-color", "#3C424E", "The foreground color")
	c.fs.StringVar(&c.BgColor, "bg-color", "#282C34", "The background color")
	c.fs.BoolVar(&c.RpgTheme, "rpg-theme", false, "Use the RPG theme")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

package main

import (
	"flag"
	"fmt"
	"image/png"
	"os"
	"strings"

	"github.com/arran4/goa4web/handlers/share"
)

// Add Description field to struct
type testGenOgImageCmd struct {
	*testCmd
	fs          *flag.FlagSet
	Title       string
	Description string
	OutputFile  string
}

func parseTestGenOgImageCmd(parent *testCmd, args []string) (*testGenOgImageCmd, error) {
	c := &testGenOgImageCmd{testCmd: parent}
	c.fs = newFlagSet("gen-og-image")
	c.fs.StringVar(&c.Title, "title", "GoA4Web", "The title to use in the image")
	c.fs.StringVar(&c.Description, "description", "", "The description to use in the image")
	c.fs.StringVar(&c.OutputFile, "output", "og-image.png", "The output file")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	c.Description = strings.ReplaceAll(c.Description, "\\n", "\n")
	return c, nil
}

func (c *testGenOgImageCmd) Run() error {
	img, err := share.GenerateImage(c.Title, c.Description)
	if err != nil {
		return fmt.Errorf("generate image: %w", err)
	}

	f, err := os.Create(c.OutputFile)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		return fmt.Errorf("encode png: %w", err)
	}

	c.rootCmd.Infof("Generated OG image for title %q at %s", c.Title, c.OutputFile)
	return nil
}

func (c *testGenOgImageCmd) Usage() {
	executeUsage(c.fs.Output(), "test_gen_og_image_usage.txt", c)
}

func (c *testGenOgImageCmd) FlagGroups() []flagGroup {
	return append(c.testCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*testGenOgImageCmd)(nil)

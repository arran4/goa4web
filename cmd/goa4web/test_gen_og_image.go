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
	Type        string
	Section     string
	Body        string
	OutputFile  string
}

func parseTestGenOgImageCmd(parent *testCmd, args []string) (*testGenOgImageCmd, error) {
	c := &testGenOgImageCmd{testCmd: parent}
	c.fs = newFlagSet("gen-og-image")
	c.fs.StringVar(&c.Title, "title", "GoA4Web", "The title to use in the image")
	c.fs.StringVar(&c.Description, "description", "", "The description to use in the image")
	c.fs.StringVar(&c.Type, "type", "default", "The generator type (default, forum)")
	c.fs.StringVar(&c.Section, "section", "", "Section label (forum only)")
	c.fs.StringVar(&c.Body, "body", "", "Body text (forum only)")
	c.fs.StringVar(&c.OutputFile, "output", "og-image.png", "The output file")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	c.Description = strings.ReplaceAll(c.Description, "\\n", "\n")
	c.Body = strings.ReplaceAll(c.Body, "\\n", "\n")
	return c, nil
}

func (c *testGenOgImageCmd) Run() error {
	var opts []interface{}
	opts = append(opts, share.WithGeneratorType(c.Type))

	if c.Title != "" {
		opts = append(opts, share.WithTitle(c.Title))
	}
	if c.Description != "" {
		opts = append(opts, share.WithDescription(c.Description))
	}
	if c.Section != "" {
		opts = append(opts, share.WithSection(c.Section))
	}
	if c.Body != "" {
		opts = append(opts, share.WithBody(c.Body))
	}

	// Generate
	img, err := share.Generate(opts...)
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

	c.rootCmd.Infof("Generated OG image for title %q (type: %s) at %s", c.Title, c.Type, c.OutputFile)
	return nil
}

func (c *testGenOgImageCmd) Usage() {
	executeUsage(c.fs.Output(), "test_gen_og_image_usage.txt", c)
}

func (c *testGenOgImageCmd) FlagGroups() []flagGroup {
	return append(c.testCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*testGenOgImageCmd)(nil)

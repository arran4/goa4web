package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/arran4/goa4web/internal/upload"
	"github.com/arran4/goa4web/runtimeconfig"
)

// imagesCmd implements image cache management utilities.
type imagesCmd struct {
	*rootCmd
	fs   *flag.FlagSet
	args []string
}

func parseImagesCmd(parent *rootCmd, args []string) (*imagesCmd, error) {
	c := &imagesCmd{rootCmd: parent}
	fs := flag.NewFlagSet("images", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *imagesCmd) Run() error {
	if len(c.args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing images command")
	}
	switch c.args[0] {
	case "cache":
		return c.runCache(c.args[1:])
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown images command %q", c.args[0])
	}
}

func (c *imagesCmd) runCache(args []string) error {
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing cache command")
	}
	dir := runtimeconfig.AppRuntimeConfig.ImageCacheDir
	switch args[0] {
	case "prune":
		if cp := upload.CacheProviderFromConfig(runtimeconfig.AppRuntimeConfig); cp != nil {
			if ccp, ok := cp.(upload.CacheProvider); ok {
				return ccp.Cleanup(context.Background(), int64(runtimeconfig.AppRuntimeConfig.ImageCacheMaxBytes))
			}
		}
		return nil
	case "list":
		return listCache(dir)
	case "delete":
		if len(args) < 2 {
			return fmt.Errorf("cache delete requires id")
		}
		return os.Remove(filepath.Join(dir, args[1]))
	case "open":
		if len(args) < 2 {
			return fmt.Errorf("cache open requires id")
		}
		path := filepath.Join(dir, args[1])
		cmd := exec.Command("xdg-open", path)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown cache command %q", args[0])
	}
}

func listCache(dir string) error {
	return filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		fmt.Printf("%s\t%d\n", rel, info.Size())
		return nil
	})
}

func (c *imagesCmd) Usage() {
	fmt.Fprintln(c.fs.Output(), "images cache [prune|list|delete <id>|open <id>]")
}

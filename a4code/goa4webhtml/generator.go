package goa4webhtml

import (
	"fmt"
	htmlstd "html"
	"io"

	"github.com/arran4/goa4web/a4code/ast"
	"github.com/arran4/goa4web/a4code/html"
)

// LinkProvider provides HTML rendering for links.
type LinkProvider interface {
	RenderLink(url string, isBlock bool, isImmediateClose bool) (htmlOpen string, htmlClose string, consumeImmediate bool)
}

// ImageMapper maps tag and value to an image URL.
type ImageMapper func(tag, val string) string

// UserColorMapper maps a username to a CSS class.
type UserColorMapper func(username string) string

type Generator struct {
	*html.Generator
	LinkProvider    LinkProvider
	ImageMapper     ImageMapper
	UserColorMapper UserColorMapper
}

type Option func(*Generator)

func WithLinkProvider(lp LinkProvider) Option {
	return func(g *Generator) { g.LinkProvider = lp }
}

func WithImageMapper(im ImageMapper) Option {
	return func(g *Generator) { g.ImageMapper = im }
}

func WithUserColorMapper(ucm UserColorMapper) Option {
	return func(g *Generator) { g.UserColorMapper = ucm }
}

func NewGenerator(opts ...interface{}) *Generator {
	g := &Generator{
		Generator: html.NewGenerator(),
	}
	g.Generator.Self = g // Set Self reference for recursion to use overrides
	for _, opt := range opts {
		switch v := opt.(type) {
		case Option:
			v(g)
		case LinkProvider:
			g.LinkProvider = v
		case ImageMapper:
			g.ImageMapper = v
		case UserColorMapper:
			g.UserColorMapper = v
		case func(tag, val string) string:
			g.ImageMapper = v
		case func(username string) string:
			g.UserColorMapper = v
		}
	}
	return g
}

func (g *Generator) Link(w io.Writer, n *ast.Link) error {
	if g.LinkProvider != nil {
		htmlOpen, htmlClose, _ := g.LinkProvider.RenderLink(n.Href, n.IsBlock, n.IsImmediateClose)
		if _, err := io.WriteString(w, htmlOpen); err != nil {
			return err
		}
		for _, c := range n.Children {
			if err := ast.Generate(w, c, g); err != nil {
				return err
			}
		}
		if _, err := io.WriteString(w, htmlClose); err != nil {
			return err
		}
		return nil
	}
	// Default behavior via base generator if no provider
	return g.Generator.Link(w, n)
}

func (g *Generator) Image(w io.Writer, n *ast.Image) error {
	src := n.Src
	if g.ImageMapper != nil {
		src = g.ImageMapper("img", src)
	}
	// We need to create a temporary node with the mapped src if we want to reuse base generator,
	// or just reimplement image rendering here. Reimplementing is safer/cleaner than mutating AST.
	if _, err := io.WriteString(w, "<img src=\""); err != nil {
		return err
	}
	if _, err := io.WriteString(w, htmlstd.EscapeString(src)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, `" data-start-pos="%d" data-end-pos="%d" />`, n.Start, n.End); err != nil {
		return err
	}
	return nil
}

func (g *Generator) QuoteOf(w io.Writer, n *ast.QuoteOf) error {
	colorClass := fmt.Sprintf("quote-color-%d", g.Depth%6)
	if g.UserColorMapper != nil {
		colorClass = g.UserColorMapper(n.Name) + " " + colorClass
	}
	fmt.Fprintf(w, `<blockquote class="a4code-block a4code-quoteof %s" data-start-pos="%d" data-end-pos="%d">`, colorClass, n.Start, n.End)
	io.WriteString(w, "<div class=\"quote-header\">Quote of ")
	io.WriteString(w, htmlstd.EscapeString(n.Name))
	io.WriteString(w, ":</div>")
	io.WriteString(w, "<div class=\"quote-body\">")

	// We need to create a child generator that maintains the structure and increments depth
	childGen := &Generator{
		Generator:       &html.Generator{Depth: g.Depth + 1},
		LinkProvider:    g.LinkProvider,
		ImageMapper:     g.ImageMapper,
		UserColorMapper: g.UserColorMapper,
	}

	for _, c := range n.Children {
		if err := ast.Generate(w, c, childGen); err != nil {
			return err
		}
	}
	io.WriteString(w, "</div>")
	io.WriteString(w, "</blockquote>")
	return nil
}

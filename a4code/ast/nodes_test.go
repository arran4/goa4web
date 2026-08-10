package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsBlockNode(t *testing.T) {
	t.Run("Root is block", func(t *testing.T) {
		root := &Root{}
		assert.True(t, IsBlockNode(root))
	})

	t.Run("Orphan Block is block", func(t *testing.T) {
		code := &Code{Value: "echo hello"}
		assert.True(t, IsBlockNode(code))
	})

	t.Run("Orphan Inline is not block", func(t *testing.T) {
		text := &Text{Value: "hello"}
		assert.False(t, IsBlockNode(text))
	})

	t.Run("HR is block", func(t *testing.T) {
		hr := &HR{}
		assert.True(t, IsBlockNode(hr))
	})

	t.Run("QuoteOf is block", func(t *testing.T) {
		quoteOf := &QuoteOf{Name: "user"}
		assert.True(t, IsBlockNode(quoteOf))
	})

	t.Run("Indent is block", func(t *testing.T) {
		indent := &Indent{}
		assert.True(t, IsBlockNode(indent))
	})

	t.Run("Inlinable Code on same line as strict inline text", func(t *testing.T) {
		root := &Root{}
		txt := &Text{Value: "Prefix "}
		code := &Code{Value: "var x = 1"}
		root.AddChild(txt)
		root.AddChild(code)

		assert.False(t, IsBlockNode(code))
	})

	t.Run("Code with newline is block even with inline text", func(t *testing.T) {
		root := &Root{}
		txt := &Text{Value: "Prefix "}
		code := &Code{Value: "var x = 1\nvar y = 2"}
		root.AddChild(txt)
		root.AddChild(code)

		assert.True(t, IsBlockNode(code))
	})

	t.Run("Standalone Link in Root becomes block", func(t *testing.T) {
		root := &Root{}
		link := &Link{Href: "https://example.com"}
		root.AddChild(link)

		assert.True(t, IsBlockNode(link))
	})

	t.Run("Link with strict inline sibling stays inline", func(t *testing.T) {
		root := &Root{}
		txt := &Text{Value: "Click here: "}
		link := &Link{Href: "https://example.com"}
		root.AddChild(txt)
		root.AddChild(link)

		assert.False(t, IsBlockNode(link))
	})

	t.Run("Inlinable Quote separated by newlines", func(t *testing.T) {
		root := &Root{}
		nl1 := &Text{Value: "\n"}
		quote := &Quote{}
		quote.AddChild(&Text{Value: "Single line quote"})
		nl2 := &Text{Value: "\n"}

		root.AddChild(nl1)
		root.AddChild(quote)
		root.AddChild(nl2)

		assert.True(t, IsBlockNode(quote))
	})
}

func TestNodeGettersSettersAndPos(t *testing.T) {
	base := &BaseNode{}
	base.SetPos(10, 20)
	start, end := base.GetPos()
	assert.Equal(t, 10, start)
	assert.Equal(t, 20, end)

	parentRoot := &Root{}
	base.SetParent(parentRoot)
	assert.Equal(t, parentRoot, base.GetParent())
}

func TestNodeStringMethods(t *testing.T) {
	root := &Root{}
	bold := &Bold{}
	bold.AddChild(&Text{Value: "bold text"})
	italic := &Italic{}
	italic.AddChild(&Text{Value: "italic text"})
	underline := &Underline{}
	underline.AddChild(&Text{Value: "underline text"})
	sup := &Sup{}
	sup.AddChild(&Text{Value: "sup text"})
	sub := &Sub{}
	sub.AddChild(&Text{Value: "sub text"})
	link := &Link{Href: "https://example.com"}
	link.AddChild(&Text{Value: "link text"})
	img := &Image{Src: "image.png"}
	code := &Code{Value: "code text"}
	codeIn := &CodeIn{Language: "go", Value: "codein text"}
	quote := &Quote{}
	quote.AddChild(&Text{Value: "quote text"})
	quoteOf := &QuoteOf{Name: "alice"}
	quoteOf.AddChild(&Text{Value: "quoteof text"})
	spoiler := &Spoiler{}
	spoiler.AddChild(&Text{Value: "spoiler text"})
	indent := &Indent{}
	indent.AddChild(&Text{Value: "indent text"})
	hr := &HR{}
	custom := &Custom{Tag: "tag"}
	custom.AddChild(&Text{Value: "custom text"})

	root.AddChild(bold)
	root.AddChild(italic)
	root.AddChild(underline)
	root.AddChild(sup)
	root.AddChild(sub)
	root.AddChild(link)
	root.AddChild(img)
	root.AddChild(code)
	root.AddChild(codeIn)
	root.AddChild(quote)
	root.AddChild(quoteOf)
	root.AddChild(spoiler)
	root.AddChild(indent)
	root.AddChild(hr)
	root.AddChild(custom)

	assert.Equal(t, "[bbold text]", bold.String())
	assert.Equal(t, "[iitalic text]", italic.String())
	assert.Equal(t, "[uunderline text]", underline.String())
	assert.Equal(t, "[supsup text]", sup.String())
	assert.Equal(t, "[subsub text]", sub.String())
	assert.Equal(t, "[link https://example.comlink text]", link.String())
	assert.Equal(t, "[img=image.png]", img.String())
	assert.Equal(t, "[code code text]", code.String())
	assert.Equal(t, "[codein \"go\" codein text]", codeIn.String())
	assert.Equal(t, "[quotequote text]", quote.String())
	assert.Equal(t, "[quoteof alicequoteof text]", quoteOf.String())
	assert.Equal(t, "[spoilerspoiler text]", spoiler.String())
	assert.Equal(t, "[indentindent text]", indent.String())
	assert.Equal(t, "[hr]", hr.String())
	assert.Equal(t, "[tagcustom text]", custom.String())
	assert.NotEmpty(t, root.String())
}

func TestNodeTransform(t *testing.T) {
	root := &Root{}
	text := &Text{Value: "hello"}
	root.AddChild(text)

	// Transform that uppercases text values
	transformed, err := root.Transform(func(n Node) (Node, error) {
		if _, ok := n.(*Text); ok {
			return &Text{Value: "HELLO"}, nil
		}
		return n, nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "HELLO", transformed.String())
}

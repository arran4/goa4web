package a4code_test

import (
	"testing"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/a4code/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinkIsImmediateClose(t *testing.T) {
	t.Run("ImmediateClose", func(t *testing.T) {
		input := "[link http://example.com]"
		root, err := a4code.ParseString(input)
		require.NoError(t, err)

		require.Len(t, root.Children, 1)
		link, ok := root.Children[0].(*ast.Link)
		require.True(t, ok)
		assert.True(t, link.IsImmediateClose, "Expected IsImmediateClose to be true for [link url]")
		assert.Equal(t, "http://example.com", link.Href)
		assert.Empty(t, link.Children)
	})

	t.Run("WithContent", func(t *testing.T) {
		// a4code syntax [link url content]
		input := "[link http://example.com Text]"
		root, err := a4code.ParseString(input)
		require.NoError(t, err)

		require.Len(t, root.Children, 1)
		link, ok := root.Children[0].(*ast.Link)
		require.True(t, ok)
		assert.False(t, link.IsImmediateClose, "Expected IsImmediateClose to be false for [link url content]")
		assert.Equal(t, "http://example.com", link.Href)

		require.NotEmpty(t, link.Children)
		txt, ok := link.Children[len(link.Children)-1].(*ast.Text)
		require.True(t, ok)
		assert.Contains(t, txt.Value, "Text")
	})

	t.Run("ExplicitContainer", func(t *testing.T) {
		// a4code syntax [link=url]content[/link]
		input := "[link=http://example.com]Content[/link]"
		root, err := a4code.ParseString(input)
		require.NoError(t, err)

		require.Len(t, root.Children, 1)
		link, ok := root.Children[0].(*ast.Link)
		require.True(t, ok)
		assert.False(t, link.IsImmediateClose, "Expected IsImmediateClose to be false for [link=url]...[/link]")
		assert.Equal(t, "http://example.com", link.Href)

		// Expect content and Custom(/link) as children (since parser adds closing tag to parent)
		require.NotEmpty(t, link.Children)

		hasContent := false
		for _, child := range link.Children {
			if txt, ok := child.(*ast.Text); ok && txt.Value == "Content" {
				hasContent = true
				break
			}
		}
		assert.True(t, hasContent, "Link should contain text content")
	})

	t.Run("Mixed", func(t *testing.T) {
		input := "[link http://a A] [link http://b]"
		root, err := a4code.ParseString(input)
		require.NoError(t, err)

		// Link(A), Text( ), Link(B)
		// Note: The previous logic might have produced separate nodes for [link http://b].
		require.True(t, len(root.Children) >= 2)

		link1, ok := root.Children[0].(*ast.Link)
		require.True(t, ok)
		assert.False(t, link1.IsImmediateClose, "First link has content")

		// Find second link
		var link2 *ast.Link
		for _, c := range root.Children {
			if l, ok := c.(*ast.Link); ok && l.Href == "http://b" {
				link2 = l
				break
			}
		}
		require.NotNil(t, link2)
		assert.True(t, link2.IsImmediateClose, "Second link has no content")
	})
}

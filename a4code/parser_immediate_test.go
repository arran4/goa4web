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
		// a4code syntax is [link url content]
		input := "[link http://example.com Text]"
		root, err := a4code.ParseString(input)
		require.NoError(t, err)

		require.Len(t, root.Children, 1)
		link, ok := root.Children[0].(*ast.Link)
		require.True(t, ok)
		assert.False(t, link.IsImmediateClose, "Expected IsImmediateClose to be false for [link url content]")
		assert.Equal(t, "http://example.com", link.Href)

		// The parser captures the space before Text as part of the text content
		require.Len(t, link.Children, 1)
		txt, ok := link.Children[0].(*ast.Text)
		require.True(t, ok)
		assert.Contains(t, txt.Value, "Text")
	})

	t.Run("Mixed", func(t *testing.T) {
		input := "[link http://a A] [link http://b]"
		root, err := a4code.ParseString(input)
		require.NoError(t, err)

		// Link(A), Text( ), Link(B)
		require.Len(t, root.Children, 3)

		link1, ok := root.Children[0].(*ast.Link)
		require.True(t, ok)
		assert.False(t, link1.IsImmediateClose, "First link has content")

		link2, ok := root.Children[2].(*ast.Link)
		require.True(t, ok)
		assert.True(t, link2.IsImmediateClose, "Second link has no content")
	})
}

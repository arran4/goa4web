package a4code

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinkIsImmediateClose(t *testing.T) {
	t.Run("NoContent", func(t *testing.T) {
		// a4code syntax [link url] (no content)
		input := "[link http://example.com]"
		root, err := ParseString(input)
		assert.NoError(t, err)

		link := findFirstLink(root)
		assert.NotNil(t, link)
		assert.Equal(t, "http://example.com", link.Href)
		assert.True(t, link.IsImmediateClose, "Expected IsImmediateClose to be true for [link url]")
		assert.Empty(t, link.Children)
	})

	t.Run("WithContent", func(t *testing.T) {
		// a4code syntax [link url Content]
		input := "[link http://example.com Content]"
		root, err := ParseString(input)
		assert.NoError(t, err)

		link := findFirstLink(root)
		assert.NotNil(t, link)
		assert.Equal(t, "http://example.com", link.Href)
		assert.False(t, link.IsImmediateClose, "Expected IsImmediateClose to be false for [link url Content]")
		assert.NotEmpty(t, link.Children)
	})

	t.Run("WithNestedContent", func(t *testing.T) {
		// a4code syntax [link url [b Bold]]
		input := "[link http://example.com [b Bold]]"
		root, err := ParseString(input)
		assert.NoError(t, err)

		link := findFirstLink(root)
		assert.NotNil(t, link)
		assert.Equal(t, "http://example.com", link.Href)
		assert.False(t, link.IsImmediateClose, "Expected IsImmediateClose to be false for [link url [b Bold]]")
		assert.NotEmpty(t, link.Children)
	})
}

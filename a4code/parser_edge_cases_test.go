package a4code

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/arran4/goa4web/a4code/ast"
)

func TestParserEdgeCases(t *testing.T) {
	// Case 1: [code][quote]]
	t.Run("CodeWithQuoteInside", func(t *testing.T) {
		input := "[code][quote]]"
		root, err := ParseString(input)
		assert.NoError(t, err)

		assert.NotEmpty(t, root.Children)
		if len(root.Children) > 0 {
			first := root.Children[0]
			codeNode, ok := first.(*ast.Code)
			assert.True(t, ok, "Expected Code node")
			if ok {
				assert.Equal(t, "[quote]", codeNode.Value)
			}
		}
	})

	// Case 2: [code][quote]] with space
	t.Run("CodeWithQuoteInsideAndSpace", func(t *testing.T) {
		input := "[code][quote]] "
		root, err := ParseString(input)
		assert.NoError(t, err)

		assert.NotEmpty(t, root.Children)
		if len(root.Children) > 0 {
			first := root.Children[0]
			codeNode, ok := first.(*ast.Code)
			assert.True(t, ok, "Expected Code node")
			if ok {
				assert.Equal(t, "[quote]", codeNode.Value)
			}
		}

		// Check for subsequent text
		if len(root.Children) > 1 {
			txt, ok := root.Children[1].(*ast.Text)
			assert.True(t, ok, "Expected Text node")
			if ok {
				assert.Equal(t, " ", txt.Value)
			}
		}
	})

	// Case 3: [code][quote]] More text
	t.Run("CodeWithQuoteInsideAndMoreText", func(t *testing.T) {
		input := "[code][quote]] More text"
		root, err := ParseString(input)
		assert.NoError(t, err)

		assert.NotEmpty(t, root.Children)
		if len(root.Children) > 0 {
			first := root.Children[0]
			codeNode, ok := first.(*ast.Code)
			assert.True(t, ok, "Expected Code node")
			if ok {
				assert.Equal(t, "[quote]", codeNode.Value)
			}
		}

		// Check for subsequent text
		if len(root.Children) > 1 {
			txt, ok := root.Children[1].(*ast.Text)
			assert.True(t, ok, "Expected Text node")
			if ok {
				assert.Equal(t, " More text", txt.Value)
			}
		}
	})

	// Case 4: [code] [quote]]
	t.Run("CodeWithSpaceAndQuoteInside", func(t *testing.T) {
		input := "[code] [quote]]"
		root, err := ParseString(input)
		assert.NoError(t, err)

		assert.NotEmpty(t, root.Children)
		if len(root.Children) > 0 {
			first := root.Children[0]
			codeNode, ok := first.(*ast.Code)
			assert.True(t, ok, "Expected Code node")
			if ok {
				// Space is preserved because it's followed by '[' which is not newline
				assert.Equal(t, " [quote]", codeNode.Value)
			}
		}
	})

	// Case 5: [code][quote][/code] (Legacy-ish)
	t.Run("CodeWithQuoteAndClosingTag", func(t *testing.T) {
		input := "[code][quote][/code]"
		root, err := ParseString(input)
		assert.NoError(t, err)

		assert.NotEmpty(t, root.Children)
		if len(root.Children) > 0 {
			first := root.Children[0]
			codeNode, ok := first.(*ast.Code)
			assert.True(t, ok)
			if ok {
				assert.Equal(t, "[quote][/code]", codeNode.Value)
			}
		}
	})
}

package goa4webhtml

import (
	"io"

	"github.com/arran4/goa4web/a4code/ast"
	"github.com/arran4/goa4web/a4code/html"
)

type Generator struct {
	*html.Generator
}

func NewGenerator() *Generator {
	return &Generator{
		Generator: html.NewGenerator(),
	}
}

// Override specific methods if Goa4Web requires different behavior.
// For now, it largely reuses standard HTML generation but with potential for injection.
// The comment "this is where we put all the goa4web specific generation stuff, such as the posiion offsets in the html components"
// suggests that standard HTML generator *has* offsets, which it does.
// If there are other specific requirements, they would go here.
// For example, if we need to inject user colors differently or handle images differently.

// Example override (currently identical to base, but placeholder for customization):
func (g *Generator) Text(w io.Writer, t *ast.Text) error {
	// Goa4Web specific text handling if needed, otherwise delegate
	return g.Generator.Text(w, t)
}

// TODO: Move Goa4Web specific logic here if it diverges from standard HTML.
// Currently the standard HTML generator includes data-start-pos/end-pos which seems to be what is expected.

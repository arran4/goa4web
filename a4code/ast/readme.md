# a4code/ast

## Purpose

Package `ast` defines the Abstract Syntax Tree (AST) nodes for the A4Code markup language. It provides the core data structures used to represent parsed A4Code elements in memory before they are formatted or rendered.

## Structure and Components

The primary files and their general responsibilities include:

- `nodes.go`
- `nodes_test.go`
- `walk.go`
- `generator.go`

### Exported Types

- `Node`
- `BaseNode`
- `Container`
- `Root`
- `Text`
- `Bold`
- `Italic`
- `Underline`
- `Sup`
- `Sub`
- `Link`
- `Image`
- `Code`
- `CodeIn`
- `Quote`
- `QuoteOf`
- `Spoiler`
- `Indent`
- `HR`
- `Custom`
- `Block`
- `BlockWithInlinable`
- `Inline`
- `InlineWithBlockable`
- `Generator`

### Exported Functions

- `IsBlockNode`
- `TestIsBlockNode`
- `TestNodeGettersSettersAndPos`
- `TestNodeStringMethods`
- `TestNodeTransform`
- `Walk`
- `Generate`

## Usage

This package is foundational to the a4code compiler. It defines all the nodes. Below is an example of the typical structure and how to walk it:

```go
import "goa4web/a4code/ast"

// Example: type-switching on an AST node during rendering
func renderNode(node ast.Node) string {
    switch n := node.(type) {
    case *ast.Text:
        return escapeHtml(n.Value)
    case *ast.Bold:
        return "<strong>" + renderNode(n.Child) + "</strong>"
    case *ast.Root:
        var out string
        for _, child := range n.Children {
             out += renderNode(child)
        }
        return out
    // ... handle all other ast nodes
    default:
        return ""
    }
}
```

### Modifying the AST
If you are adding a new tag or token to the markup language:
1. Define the struct here in `ast.go` (e.g. `type MyNewNode struct { ast.NodeImpl; Value string }`).
2. Ensure it implements the `Node` interface.
3. Update the parsers in `a4code` and all corresponding renderers (`format`, `html`, etc) to handle the new node type.

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.

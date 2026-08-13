# a4code/ast

## Purpose

Package `ast` defines the Abstract Syntax Tree (AST) nodes for the A4Code markup language. It provides the core data structures used to represent parsed A4Code elements in memory before they are formatted or rendered.

## Why It Exists

When the parser reads raw text, it needs an intermediate representation to apply rules, strip invalid tags, or transform content before turning it into HTML or Text. The AST provides this structured, type-safe tree, serving as the central nervous system of the markup engine.

## What It Allows

It allows developers to programmatically inspect, modify, and traverse user-submitted content prior to final output. For instance, you can count the number of images, strip out specific formatting, or enforce nesting limits by simply walking the tree.

## Structure and Components

The primary files and their general responsibilities include:

- `generator.go`
- `nodes.go`
- `nodes_test.go`
- `walk.go`

### Exported Types and Interfaces

- **`Underline`**:
  - Implements: Node (partially/fully)
  - Methods: `AddChild`, `GetChildren`, `Transform`, `String`
- **`Sub`**:
  - Implements: Node (partially/fully)
  - Methods: `AddChild`, `GetChildren`, `Transform`, `String`
- **`Indent`**:
  - Implements: Node (partially/fully)
  - Methods: `AddChild`, `GetChildren`, `Transform`, `String`
- **`Custom`**:
  - Implements: Node (partially/fully)
  - Methods: `AddChild`, `GetChildren`, `Transform`, `String`
- **`InlineWithBlockable`** (Interface): Defines a core contract for this module.
- **`BaseNode`**:
  - Methods: `SetPos`, `GetPos`, `GetParent`, `SetParent`
- **`Text`**:
  - Implements: Node (partially/fully)
  - Methods: `Transform`, `String`
- **`Italic`**:
  - Implements: Node (partially/fully)
  - Methods: `AddChild`, `GetChildren`, `Transform`, `String`
- **`Code`**:
  - Implements: Node (partially/fully)
  - Methods: `Inlinable`, `Transform`, `String`
- **`CodeIn`**:
  - Implements: Node (partially/fully)
  - Methods: `Inlinable`, `Transform`, `String`
- **`Quote`**:
  - Implements: Node (partially/fully)
  - Methods: `Inlinable`, `AddChild`, `GetChildren`, `Transform`, `String`
- **`HR`**:
  - Implements: Node (partially/fully)
  - Methods: `Transform`, `String`
- **`Generator`** (Interface): Defines a core contract for this module.
- **`Node`** (Interface): Defines a core contract for this module.
- **`Root`**:
  - Implements: Node (partially/fully)
  - Methods: `Transform`, `AddChild`, `GetChildren`, `String`
- **`Bold`**:
  - Implements: Node (partially/fully)
  - Methods: `AddChild`, `GetChildren`, `Transform`, `String`
- **`Link`**:
  - Implements: Node (partially/fully)
  - Methods: `Blockable`, `AddChild`, `GetChildren`, `IsImmediateClose`, `Transform`, `String`
- **`QuoteOf`**:
  - Implements: Node (partially/fully)
  - Methods: `AddChild`, `GetChildren`, `Transform`, `String`
- **`Spoiler`**:
  - Implements: Node (partially/fully)
  - Methods: `AddChild`, `GetChildren`, `Transform`, `String`
- **`Block`** (Interface): Defines a core contract for this module.
- **`Sup`**:
  - Implements: Node (partially/fully)
  - Methods: `AddChild`, `GetChildren`, `Transform`, `String`
- **`Image`**:
  - Implements: Node (partially/fully)
  - Methods: `Transform`, `String`
- **`BlockWithInlinable`** (Interface): Defines a core contract for this module.
- **`Inline`** (Interface): Defines a core contract for this module.
- **`Container`** (Interface): Defines a core contract for this module.

### Exported Functions

- `Generate`
- `IsBlockNode`
- `TestIsBlockNode`
- `TestNodeGettersSettersAndPos`
- `TestNodeStringMethods`
- `TestNodeTransform`
- `Walk`

## Usage Examples

You interact with the AST primarily by writing functions that accept `ast.Node` and use a type-switch to determine the specific element type (e.g., `*ast.Text`, `*ast.Bold`). Below is an example of the typical structure and how to walk it:

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

# a4code/html

## Purpose

Package `html` provides the rendering engine that converts an A4Code Abstract Syntax Tree (AST) into standard HTML output suitable for web browsers.

## Structure and Components

The primary files and their general responsibilities include:

- `generator.go`
- `issue_link_test.go`

### Exported Types and Interfaces

- **`Option`**:
- **`DataPositionAttrs`**:
  - Methods: `SourceAttrs`
- **`Generator`**:
  - Methods: `Root`, `Text`, `Bold`, `Italic`, `Underline`, `Sup`, `Sub`, `Link`, `Image`, `Code`, `CodeIn`, `Quote`, `QuoteOf`, `Spoiler`, `Indent`, `HR`, `Custom`, `SourceAttrs`
- **`SourceAttrBuilder`** (Interface): Defines a core contract for this module.

### Exported Functions

- `WithDataPositions`
- `WithSourceAttrBuilder`
- `NewGenerator`
- `SanitizeURL`
- `TestLinkWithWhitespaceChildren`

## Usage

The HTML package recursively translates an `ast.Node` tree into raw HTML bytes.

```go
import "goa4web/a4code"
import "goa4web/a4code/html"

astRoot, _ := a4code.Parse("Some input text")

// Render to an io.Writer (like a strings.Builder or http.ResponseWriter)
var buf strings.Builder
err := html.Render(&buf, astRoot)
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.

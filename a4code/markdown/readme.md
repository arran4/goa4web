# a4code/markdown

## Purpose

Package `markdown` provides utilities for converting standard Markdown input into A4Code markup, or potentially rendering A4Code as Markdown.

## Structure and Components

The primary files and their general responsibilities include:

- `generator.go`

### Exported Types

- `SmartWriter`
- `Generator`

### Exported Functions

- `NewGenerator`

## Usage

The typical workflow involves parsing an input string into an AST, then handing that AST off to a renderer (like `a4code2html` or `markdown`).

```go
import "goa4web/a4code"

// 1. Parse raw input string into an AST
astRoot, err := a4code.Parse("Some [b]input[/b] text")
if err != nil {
    // handle parser errors
}

// 2. The AST is now ready to be traversed or rendered.
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.

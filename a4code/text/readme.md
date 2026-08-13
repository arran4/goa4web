# a4code/text

## Purpose

Package `text` provides a plain-text renderer for the A4Code Abstract Syntax Tree (AST), useful for stripping formatting and extracting pure content.

## Structure and Components

The primary files and their general responsibilities include:

- `generator.go`

### Exported Types and Interfaces

- **`SmartWriter`**:
  - Methods: `Write`
- **`PrefixWriter`**:
  - Methods: `Write`
- **`Generator`**:
  - Methods: `Root`, `Text`, `Bold`, `Italic`, `Underline`, `Sup`, `Sub`, `Link`, `Image`, `Code`, `CodeIn`, `Quote`, `QuoteOf`, `Spoiler`, `Indent`, `HR`, `Custom`

### Exported Functions

- `NewGenerator`
- `NewCleanGenerator`

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

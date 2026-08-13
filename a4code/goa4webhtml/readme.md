# a4code/goa4webhtml

## Purpose

Package `goa4webhtml` provides specialized HTML rendering for A4Code that is specifically tailored and integrated with the Goa4Web templating and asset system.

## Structure and Components

The primary files and their general responsibilities include:

- `generator.go`

### Exported Types and Interfaces

- **`LinkProvider`** (Interface): Defines a core contract for this module.
- **`ImageMapper`**:
- **`FullImageMapper`**:
- **`UserColorMapper`**:
- **`Generator`**:
  - Methods: `Link`, `Image`, `QuoteOf`
- **`Option`**:

### Exported Functions

- `WithLinkProvider`
- `WithImageMapper`
- `WithFullImageMapper`
- `WithUserColorMapper`
- `WithDataPositions`
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

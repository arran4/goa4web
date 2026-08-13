# a4code/format

## Purpose

Package `format` provides utilities for taking an A4Code Abstract Syntax Tree (AST) and formatting it back into a valid, normalized A4Code string. This is useful for pretty-printing or normalizing user input.

## Structure and Components

The primary files and their general responsibilities include:

- `generator.go`

### Exported Types and Interfaces

- **`Generator`**:
  - Methods: `Root`, `Text`, `Bold`, `Italic`, `Underline`, `Sup`, `Sub`, `Link`, `Image`, `Code`, `CodeIn`, `Quote`, `QuoteOf`, `Spoiler`, `Indent`, `HR`, `Custom`

### Exported Functions

- `NewGenerator`

## Usage

You would use this to normalize A4Code strings. Pass a parsed AST into the formatter to get a standard representation back. The formatter uses a visitor pattern internally.

```go
import "goa4web/a4code"
import "goa4web/a4code/format"

// Parse raw input
parsed, err := a4code.Parse("[b]some bold text[/b]")
if err != nil {
    // handle
}

// Normalize/format the parsed AST
formattedStr := format.Format(parsed)
```

If you add a new AST node, you **must** update the switch statements in this package to handle how it should be converted back to raw a4code text.

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.

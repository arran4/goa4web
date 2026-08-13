# a4code/format/test

## Purpose

Package `format` provides utilities for taking an A4Code Abstract Syntax Tree (AST) and formatting it back into a valid, normalized A4Code string. This is useful for pretty-printing or normalizing user input.

## Why It Exists

User input is often messy. Tags might be improperly cased or spaced. The formatter exists to take a parsed AST and output clean, standard A4Code text, ensuring database consistency.

## What It Allows

It allows the application to 'auto-correct' user input, storing a clean version in the database, or providing a 'pretty-print' preview back to the user.

## Structure and Components

The primary files and their general responsibilities include:


## Usage Examples

Pass a parsed `ast.Node` root into the formatter's main function to retrieve the normalized string. The formatter uses a visitor pattern internally.

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

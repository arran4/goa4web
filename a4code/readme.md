# A4Code parser

## Why it exists

A4Code is Goa4Web's deliberately small, Lisp-style user-markup language. Parsing
untrusted input into a typed AST lets each output target escape and interpret
content deliberately instead of accepting raw HTML.

## Parse and render

Tags contain their arguments and children inside one pair of brackets: `[b bold
text]` and `[link https://example.com label]`. Closing tags such as `[/b]` are
invalid.

```go
root, err := a4code.ParseString("Hello [b world]")
if err != nil { return fmt.Errorf("parse A4Code: %w", err) }

var out bytes.Buffer
if err := ast.Generate(&out, root, html.NewGenerator()); err != nil {
    return fmt.Errorf("render A4Code: %w", err)
}
```

Use `Parse(io.Reader)` for streams and `ParseString` for an in-memory value. Pick
a generator for the destination: `html`, `goa4webhtml`, `format`, `markdown`, or
`text`. `SanitizeURL`, quote, snip, and substring helpers support the forum's
preview and quoting workflows; preserve source positions when a caller needs
selection-aware behavior.

## Extending the language

A new node requires coordinated parser, AST, generator, and test changes. Add the
node to `a4code/ast`, teach the parser when to construct it, and implement every
`ast.Generator` method so no output format silently loses content. Use the syntax
rules in `a4code/AGENTS.md` in examples and fixtures.

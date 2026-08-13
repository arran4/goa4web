# Markdown generator

This package converts an **A4Code AST to Markdown**. It does not parse Markdown
and cannot convert Markdown input into A4Code.

```go
root, err := a4code.ParseString("[b bold]")
if err != nil { return err }
var out bytes.Buffer
err = ast.Generate(&out, root, markdown.NewGenerator())
```

Markdown cannot represent every application-specific A4Code behavior exactly.
Add or change a node only after deciding on a safe, readable fallback here.

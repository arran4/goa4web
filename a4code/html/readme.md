# Generic HTML generator

This package renders an A4Code AST as escaped, framework-neutral HTML. Use it for
content that does not need Goa4Web link routing, image mapping, or username CSS;
use `goa4webhtml` when those application policies are required.

```go
root, err := a4code.ParseString("Hello [b world]")
if err != nil { return err }
var out bytes.Buffer
err = ast.Generate(&out, root, html.NewGenerator())
```

`WithDataPositions` emits source offsets for editor tooling. A
`SourceAttrBuilder` can add controlled source attributes. Never concatenate raw
user markup around the generated HTML.

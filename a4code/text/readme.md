# Plain-text generator

The text generator converts an A4Code AST into readable plain text for search,
email, summaries, and other contexts where markup is inappropriate.

```go
root, err := a4code.ParseString("Hello [b world]")
if err != nil { return err }
var out bytes.Buffer
err = ast.Generate(&out, root, text.NewGenerator())
```

It intentionally discards visual formatting while preserving meaningful content.
When adding a node, decide whether its label, URL, or children convey the best
plain-text meaning and add focused generator tests.

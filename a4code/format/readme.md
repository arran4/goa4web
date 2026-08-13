# A4Code formatter

The format generator serializes an AST back to canonical A4Code. This is useful
for normalization, round-trip tests, and transformed content—not HTML display.

```go
root, err := a4code.ParseString("[b bold]")
if err != nil { return err }
var out bytes.Buffer
err = ast.Generate(&out, root, format.NewGenerator())
```

Formatting escapes syntax-sensitive text and always emits Lisp-style tags. Parse
the result again in round-trip tests when adding a node or changing escaping.

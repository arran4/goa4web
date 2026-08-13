# Goa4Web HTML generator

This generator extends generic HTML rendering with application policy. Inject a
`LinkProvider`, image mappers, and a user-color mapper when rendering stored site
content so links use Goa4Web routing and image identifiers become served URLs.

```go
generator := goa4webhtml.NewGenerator(
    goa4webhtml.WithLinkProvider(links),
    goa4webhtml.WithImageMapper(mapImage),
)
var out bytes.Buffer
err := ast.Generate(&out, root, generator)
```

Prefer the generic `html` generator when these policies are unnecessary. Keep
mappers request-independent and return already validated URLs; escaping remains
the generator's responsibility.

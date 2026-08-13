# S3 upload provider

The S3 backend is optional (`-tags s3`) and joins the upload provider registry via
`s3.Register()`. Application startup selects it from runtime configuration; code
using uploads should depend on `upload.Provider`, whose operations are `Check`,
`Write`, and `Read`, rather than constructing an S3 client directly.

```go
s3.Register()
// Startup resolves the configured "s3" provider and injects upload.Provider.
if err := provider.Check(ctx); err != nil { return err }
if err := provider.Write(ctx, name, data); err != nil { return err }
data, err := provider.Read(ctx, name)
```

The configured target is an `s3://bucket/prefix` URL. `Write` does not return a
public URL. Without the build tag, registration uses the stub behavior; tests
should mock `upload.Provider` instead of contacting AWS.

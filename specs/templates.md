# Template Compilation

Goa4Web ships with two implementations for loading HTML, text and asset templates. The
production build embeds everything using `//go:embed` so the binary is self-contained.
During development the `live` build tag swaps in a version that reads directly from
the `core/templates` directory on disk.

## Production mode

`core/templates/embedded.go` is compiled when the `live` build tag is **not** present. It
The file begins with the build constraint `//go:build !live` so it is included in regular builds.
uses multiple `//go:embed` directives to include templates and static files:

- HTML templates under `site/`
- notification templates under `notifications/`
- email templates under `email/`
- CSS and JavaScript from `assets/`

The functions such as `GetCompiledSiteTemplates` parse these embedded files with
`template.ParseFS` and return ready-to-use template sets.

## Live development mode

When built with `-tags live`, the file `core/templates/live.go` takes over.
It begins with `//go:build live` to activate when the tag is supplied. Instead of embedding data it calls `os.ReadFile` and `template.ParseFS` against `os.DirFS` to load files from disk. This allows editing templates without rebuilding the binary.
```bash
# Start the server with live templates
go run -tags live ./cmd/goa4web
```

Both files expose the same functions so callers do not need to change between modes.

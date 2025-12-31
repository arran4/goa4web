# Templates

Goa4Web uses Go's standard `text/template` and `html/template` packages.

## Loading Strategy

Templates are dual-sourced:
1. **Embedded**: Default state, using `//go:embed` in `core/templates`.
2. **Filesystem**: Overridden at runtime via the `--templates-dir` flag (or `TEMPLATES_DIR` env).

The `core/templates` package handles this abstraction, exposing functions like `GetCompiledSiteTemplates` which return parsed templates from the active source.

## Development

- **Extraction**: Use `goa4web templates extract -dir ./tmpl` to dump embedded templates to disk.
- **Live Reloading**: Run with `--templates-dir ./tmpl` to serve from disk, allowing edits without recompilation.

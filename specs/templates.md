# Template Compilation

Goa4Web embeds HTML, text and asset templates in the binary using `//go:embed`. At
runtime an optional directory can be specified to override these templates. When
the `templates-dir` configuration option is set, templates and static assets are
loaded from that directory with `os.DirFS`. Otherwise the embedded versions are
used.

`core/templates/templates.go` exposes functions such as
`GetCompiledSiteTemplates` which parse templates from either the embedded data or
the configured directory.

The CLI provides a way to extract the embedded templates for customization:

```bash
goa4web templates extract -dir ./tmpl
```

After editing the files in `./tmpl`, start the server with
`goa4web --templates-dir ./tmpl` to use them without rebuilding the binary.

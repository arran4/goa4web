1. Verify `handlers/share/opengraph.go` again to make sure everything requested by the PR reviewer is handled correctly.
The reviewer mentioned "we should escape this and others, plus isn't this redundant with the other module in common?".
I escaped the fields. Regarding redundancy, I didn't actually embed `common.OpenGraph` because it would require changing `OpenGraphData` initializations everywhere (`handlers/blogs/shared_preview.go`, `forum/...`, `news/...`, `writings/...`).

Wait, the reviewer specifically asked "isn't this redundant with the other module in common?".
I should probably address the redundancy.
But to do that cleanly without breaking templates, we can change `OpenGraphData` to:
```go
type OpenGraphData struct {
    common.OpenGraph
    RedirectURL template.URL
}
```
But `OpenGraphData` has `ContentURL` and `ImageURL` as `template.URL`. `common.OpenGraph` has `URL` and `Image` as `string`. The template `openGraphPreview.gohtml` expects `.ContentURL` and `.ImageURL`.
Wait, does it?
Let's look at `openGraphPreview.gohtml`:
```html
<link rel="canonical" href="{{.ContentURL}}" />
...
	{{if .RedirectURL}}<meta http-equiv="refresh" content="0;url={{.RedirectURL}}" />{{else}}<meta http-equiv="refresh" content="0;url={{.ContentURL}}" />{{end}}
	<p>If you are not redirected automatically, <a href="{{if .RedirectURL}}{{.RedirectURL}}{{else}}{{.ContentURL}}{{end}}">click here</a>.</p>
```
So it explicitly uses `.ContentURL`.
If we were to replace it entirely, we'd have to update the template to use `.URL`.
We could just keep the struct separate to avoid breaking `openGraphPreview.gohtml` variables, but the redundant methods `URLMeta()`, `ImageMeta()`, `SiteNameMeta()` are what they meant.

Wait! Actually, if the reviewer asked "isn't this redundant with the other module in common?", I should probably just reply that I've applied escaping for now and we can address the full deduplication in a separate refactoring PR since `OpenGraphData` uses `template.URL` for its fields which are heavily tied into `openGraphPreview.gohtml`.
Wait, I CAN just refactor it to embed it. Let's look at the PR comments.

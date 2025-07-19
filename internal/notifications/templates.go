package notifications

import (
	"bytes"
	"context"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/db"
	htemplate "html/template"
	"io"
	ttemplate "text/template"
)

type TemplateEngine interface {
	ExecuteTemplate(wr io.Writer, name string, data any) error
}

type NewTemplateEngine[TE TemplateEngine] interface {
	Parse(string) (TE, error)
}

func renderTemplate[TE TemplateEngine](ctx context.Context, q *db.Queries, filename string, data any, tmpls TE, templateNew func(string) NewTemplateEngine[TE]) ([]byte, error) {
	var buf bytes.Buffer
	if body, err := q.GetTemplateOverride(ctx, filename); err == nil && body != "" {
		if parsed, err := templateNew("").Parse(body); err != nil {
			return nil, err
		} else if err := parsed.ExecuteTemplate(&buf, "", data); err != nil {
			return nil, err
		}
	} else {
		if err := tmpls.ExecuteTemplate(&buf, filename, data); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func TextTemplatesNew(s string) NewTemplateEngine[*ttemplate.Template] {
	return ttemplate.New(s)
}

func HTMLTemplatesNew(s string) NewTemplateEngine[*htemplate.Template] {
	return htemplate.New(s)
}

func NotificationTemplateFilenameGenerator(base string) string {
	return base + ".gotxt"
}

func EmailTextTemplateFilenameGenerator(base string) string {
	return base + ".gotxt"
}

func EmailHTMLTemplateFilenameGenerator(base string) string {
	return base + ".gohtml"
}

func EmailSubjectTemplateFilenameGenerator(base string) string {
	return base + "Subject.gotxt"
}

// renderNotification renders the notification template associated with task.
// Database overrides are respected when present.
func renderNotification(ctx context.Context, q *db.Queries, filename string, data any) ([]byte, error) {
	tmpls := templates.GetCompiledNotificationTemplates(map[string]any{})
	return renderTemplate[*ttemplate.Template](ctx, q, filename, data, tmpls, TextTemplatesNew)
}

func renderEmailSubject(ctx context.Context, q *db.Queries, filename string, data any) ([]byte, error) {
	tmpls := templates.GetCompiledEmailTextTemplates(map[string]any{})
	return renderTemplate[*ttemplate.Template](ctx, q, filename, data, tmpls, TextTemplatesNew)
}

func renderEmailText(ctx context.Context, q *db.Queries, filename string, data any) ([]byte, error) {
	tmpls := templates.GetCompiledEmailTextTemplates(map[string]any{})
	return renderTemplate[*ttemplate.Template](ctx, q, filename, data, tmpls, TextTemplatesNew)
}

func renderEmailHtml(ctx context.Context, q *db.Queries, filename string, data any) ([]byte, error) {
	tmpls := templates.GetCompiledEmailHtmlTemplates(map[string]any{})
	return renderTemplate[*htemplate.Template](ctx, q, filename, data, tmpls, HTMLTemplatesNew)
}

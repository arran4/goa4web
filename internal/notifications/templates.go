package notifications

import (
	"bytes"
	"context"
	htemplate "html/template"
	"io"
	ttemplate "text/template"

	"github.com/arran4/goa4web/internal/db"
)

type TemplateEngine interface {
	ExecuteTemplate(wr io.Writer, name string, data any) error
}

type NewTemplateEngine[TE TemplateEngine] interface {
	Parse(string) (TE, error)
}

func renderTemplate[TE TemplateEngine](ctx context.Context, q db.Querier, filename string, data any, tmpls TE, templateNew func(string) NewTemplateEngine[TE]) ([]byte, error) {
	var buf bytes.Buffer
	if body, err := q.SystemGetTemplateOverride(ctx, filename); err == nil && body != "" {
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

func HTMLTemplatesNew(s string) NewTemplateEngine[*htemplate.Template] {
	return htemplate.New(s)
}

func TextTemplatesNew(s string) NewTemplateEngine[*ttemplate.Template] {
	return ttemplate.New(s)
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
func (n *Notifier) renderNotification(ctx context.Context, filename string, data any) ([]byte, error) {
	tmpls := n.notificationTemplates()
	return renderTemplate[*ttemplate.Template](ctx, n.Queries, filename, data, tmpls, TextTemplatesNew)
}

func (n *Notifier) renderEmailSubject(ctx context.Context, filename string, data any) ([]byte, error) {
	tmpls := n.emailTextTemplates()
	return renderTemplate[*ttemplate.Template](ctx, n.Queries, filename, data, tmpls, TextTemplatesNew)
}

func (n *Notifier) renderEmailText(ctx context.Context, filename string, data any) ([]byte, error) {
	tmpls := n.emailTextTemplates()
	return renderTemplate[*ttemplate.Template](ctx, n.Queries, filename, data, tmpls, TextTemplatesNew)
}

func (n *Notifier) renderEmailHtml(ctx context.Context, filename string, data any) ([]byte, error) {
	tmpls := n.emailHTMLTemplates()
	return renderTemplate[*htemplate.Template](ctx, n.Queries, filename, data, tmpls, HTMLTemplatesNew)
}

package notifications

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/config"
	dbpkg "github.com/arran4/goa4web/internal/db"
)

func TestLinkerQueueNotifierMessages(t *testing.T) {
	ctx := context.Background()
	cfg := config.GenerateRuntimeConfig(nil, map[string]string{}, func(string) string { return "" })
	cfg.EmailFrom = "from@example.com"
	t.Cleanup(func() { cfg = cfg })
	ntName := NotificationTemplateFilenameGenerator("linker_approved")

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery("SELECT body FROM template_overrides WHERE name = ?").
		WithArgs(ntName).
		WillReturnRows(sqlmock.NewRows([]string{"body"}).AddRow(""))
	mock.ExpectQuery("SELECT body FROM template_overrides WHERE name = ?").
		WithArgs("linkerApprovedEmail.gotxt").
		WillReturnRows(sqlmock.NewRows([]string{"body"}).AddRow(""))
	mock.ExpectQuery("SELECT body FROM template_overrides WHERE name = ?").
		WithArgs("linkerApprovedEmail.gohtml").
		WillReturnRows(sqlmock.NewRows([]string{"body"}).AddRow(""))
	mock.ExpectQuery("SELECT body FROM template_overrides WHERE name = ?").
		WithArgs("linkerApprovedEmailSubject.gotxt").
		WillReturnRows(sqlmock.NewRows([]string{"body"}).AddRow(""))

	q := dbpkg.New(db)
	n := New(WithQueries(q), WithConfig(*cfg))
	data := map[string]any{
		"Title":     "Example",
		"Username":  "bob",
		"Moderator": "mod",
		"LinkURL":   "http://example.com/link",
	}
	msg, err := n.renderNotification(ctx, ntName, data)
	if err != nil {
		t.Fatalf("render notification: %v", err)
	}
	if !strings.Contains(string(msg), "approved") {
		t.Fatalf("unexpected notification: %s", msg)
	}
}

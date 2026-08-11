package app

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
)

// mockQuerier implements just the methods we need
type mockQuerier struct {
	db.Querier
}

func (m *mockQuerier) SystemCountLanguages(ctx context.Context) (int64, error) {
	return 1, nil
}

func (m *mockQuerier) SystemGetLanguageIDByName(ctx context.Context, name sql.NullString) (int32, error) {
	return 1, nil
}

func (m *mockQuerier) DeleteSessionProxy(ctx context.Context, id int32) error {
	return nil
}

func TestNewServer_MissingSessionSecret(t *testing.T) {
	cfg := &config.RuntimeConfig{
		SessionName:     "test_session",
		SessionSameSite: "strict",
		ImageUploadDir:  t.TempDir(),
	}
	querier := &mockQuerier{}

	_, err := NewServer(context.Background(), cfg, nil,
		WithQuerier(querier),
	)
	if err == nil {
		t.Fatal("expected error without session secret, got nil")
	}
}

func TestNewServer_MissingDBAndQuerier(t *testing.T) {
	cfg := &config.RuntimeConfig{
		SessionName:     "test_session",
		SessionSameSite: "strict",
		ImageUploadDir:  t.TempDir(),
	}

	_, err := NewServer(context.Background(), cfg, nil,
		WithSessionSecret("secret"),
	)
	if err == nil {
		t.Fatal("expected error without DB or Querier, got nil")
	}
}

func TestNewServer_Success(t *testing.T) {
	cfg := &config.RuntimeConfig{
		SessionName:     "test_session",
		SessionSameSite: "strict",
		ImageUploadDir:  t.TempDir(),
		DefaultLanguage: "en-US",
		EmailFrom:       "test@example.com",
	}
	querier := &mockQuerier{}

	// Create directory before running since app.NewServer might try to create it.
	_ = os.MkdirAll(cfg.ImageUploadDir, 0755)

	srv, err := NewServer(context.Background(), cfg, nil,
		WithQuerier(querier),
		WithSessionSecret("secret"),
		WithImageSignSecret("img"),
		WithLinkSignSecret("link"),
		WithShareSignSecret("share"),
		WithAPISecret("api"),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if srv == nil {
		t.Fatal("expected server to be created, got nil")
	}
}

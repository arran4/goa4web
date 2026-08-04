package app

import (
	"database/sql"
	"testing"

	"github.com/arran4/goa4web/internal/dbdrivers"
	"github.com/arran4/goa4web/internal/dlq"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/eventbus"
	routerpkg "github.com/arran4/goa4web/internal/router"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/sessions"
)

func TestServerOptions(t *testing.T) {
	t.Run("WithSessionSecret", func(t *testing.T) {
		opts := &serverOptions{}
		WithSessionSecret("session-secret")(opts)
		if opts.SessionSecret != "session-secret" {
			t.Errorf("got %q, want %q", opts.SessionSecret, "session-secret")
		}
	})

	t.Run("WithImageSignSecret", func(t *testing.T) {
		opts := &serverOptions{}
		WithImageSignSecret("image-secret")(opts)
		if opts.ImageSignSecret != "image-secret" {
			t.Errorf("got %q, want %q", opts.ImageSignSecret, "image-secret")
		}
	})

	t.Run("WithLinkSignSecret", func(t *testing.T) {
		opts := &serverOptions{}
		WithLinkSignSecret("link-secret")(opts)
		if opts.LinkSignSecret != "link-secret" {
			t.Errorf("got %q, want %q", opts.LinkSignSecret, "link-secret")
		}
	})

	t.Run("WithShareSignSecret", func(t *testing.T) {
		opts := &serverOptions{}
		WithShareSignSecret("share-secret")(opts)
		if opts.ShareSignSecret != "share-secret" {
			t.Errorf("got %q, want %q", opts.ShareSignSecret, "share-secret")
		}
	})

	t.Run("WithAPISecret", func(t *testing.T) {
		opts := &serverOptions{}
		WithAPISecret("api-secret")(opts)
		if opts.APISecret != "api-secret" {
			t.Errorf("got %q, want %q", opts.APISecret, "api-secret")
		}
	})

	t.Run("WithDBRegistry", func(t *testing.T) {
		opts := &serverOptions{}
		reg := dbdrivers.NewRegistry()
		WithDBRegistry(reg)(opts)
		if opts.DBReg != reg {
			t.Errorf("got %v, want %v", opts.DBReg, reg)
		}
	})

	t.Run("WithEmailRegistry", func(t *testing.T) {
		opts := &serverOptions{}
		reg := email.NewRegistry()
		WithEmailRegistry(reg)(opts)
		if opts.EmailReg != reg {
			t.Errorf("got %v, want %v", opts.EmailReg, reg)
		}
	})

	t.Run("WithDLQRegistry", func(t *testing.T) {
		opts := &serverOptions{}
		reg := dlq.NewRegistry()
		WithDLQRegistry(reg)(opts)
		if opts.DLQReg != reg {
			t.Errorf("got %v, want %v", opts.DLQReg, reg)
		}
	})

	t.Run("WithTasksRegistry", func(t *testing.T) {
		opts := &serverOptions{}
		reg := tasks.NewRegistry()
		WithTasksRegistry(reg)(opts)
		if opts.TasksReg != reg {
			t.Errorf("got %v, want %v", opts.TasksReg, reg)
		}
	})

	t.Run("WithBus", func(t *testing.T) {
		opts := &serverOptions{}
		bus := eventbus.NewBus()
		WithBus(bus)(opts)
		if opts.Bus != bus {
			t.Errorf("got %v, want %v", opts.Bus, bus)
		}
	})

	t.Run("WithStore", func(t *testing.T) {
		opts := &serverOptions{}
		store := sessions.NewCookieStore([]byte("key"))
		WithStore(store)(opts)
		if opts.Store != store {
			t.Errorf("got %v, want %v", opts.Store, store)
		}
	})

	t.Run("WithDB", func(t *testing.T) {
		opts := &serverOptions{}
		db := &sql.DB{}
		WithDB(db)(opts)
		if opts.DB != db {
			t.Errorf("got %v, want %v", opts.DB, db)
		}
	})

	t.Run("WithRouterRegistry", func(t *testing.T) {
		opts := &serverOptions{}
		reg := routerpkg.NewRegistry()
		WithRouterRegistry(reg)(opts)
		if opts.RouterReg != reg {
			t.Errorf("got %v, want %v", opts.RouterReg, reg)
		}
	})
}

package dlq_test

import (
	"testing"

	"github.com/arran4/goa4web/config"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dlq"
	dbdlq "github.com/arran4/goa4web/internal/dlq/db"
	dirdlq "github.com/arran4/goa4web/internal/dlq/dir"
	"github.com/arran4/goa4web/internal/dlq/dlqdefaults"
	emaildlq "github.com/arran4/goa4web/internal/dlq/email"
	filedlq "github.com/arran4/goa4web/internal/dlq/file"
	emailpkg "github.com/arran4/goa4web/internal/email"
)

func TestProviderFromConfigRegistry(t *testing.T) {
	reg := dlq.NewRegistry()
	emailReg := emailpkg.NewRegistry()
	dlqdefaults.RegisterDefaults(reg, emailReg)

	cfg := config.RuntimeConfig{DLQProvider: "file", DLQFile: "p"}
	if _, ok := reg.ProviderFromConfig(&cfg, nil).(*filedlq.DLQ); !ok {
		t.Fatalf("expected *file.DLQ")
	}

	cfg = config.RuntimeConfig{DLQProvider: "dir", DLQFile: "d"}
	if _, ok := reg.ProviderFromConfig(&cfg, nil).(*dirdlq.DLQ); !ok {
		t.Fatalf("expected *dir.DLQ")
	}

	cfg = config.RuntimeConfig{DLQProvider: "db"}
	if _, ok := reg.ProviderFromConfig(&cfg, (&dbpkg.Queries{})).(dbdlq.DLQ); !ok {
		t.Fatalf("expected db.DLQ")
	}

	cfg = config.RuntimeConfig{DLQProvider: "email"}
	p := reg.ProviderFromConfig(&cfg, nil)
	if _, ok := p.(emaildlq.DLQ); !ok {
		if _, ok := p.(dlq.LogDLQ); !ok {
			t.Fatalf("unexpected type %T", p)
		}
	}

	cfg = config.RuntimeConfig{DLQProvider: "db,log"}
	if _, ok := reg.ProviderFromConfig(&cfg, (&dbpkg.Queries{})).(dlq.MultiDLQ); !ok {
		t.Fatalf("expected MultiDLQ")
	}
}

func TestRegisterProviderCustom(t *testing.T) {
	reg := dlq.NewRegistry()
	called := false
	reg.RegisterProvider("custom", func(cfg *config.RuntimeConfig, q *dbpkg.Queries) dlq.DLQ {
		called = true
		return dlq.LogDLQ{}
	})

	cfg := config.RuntimeConfig{DLQProvider: "custom"}
	if _, ok := reg.ProviderFromConfig(&cfg, nil).(dlq.LogDLQ); !ok || !called {
		t.Fatalf("custom provider not used")
	}
}

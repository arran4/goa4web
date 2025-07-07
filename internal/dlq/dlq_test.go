package dlq_test

import (
	"reflect"
	"testing"

	dbpkg "github.com/arran4/goa4web/internal/db"
	dlq "github.com/arran4/goa4web/internal/dlq"
	dbdlq "github.com/arran4/goa4web/internal/dlq/db"
	dirdlq "github.com/arran4/goa4web/internal/dlq/dir"
	dlqdefaults "github.com/arran4/goa4web/internal/dlq/dlqdefaults"
	emaildlq "github.com/arran4/goa4web/internal/dlq/email"
	filedlq "github.com/arran4/goa4web/internal/dlq/file"
	"github.com/arran4/goa4web/runtimeconfig"
)

func TestProviderFromConfigRegistry(t *testing.T) {
	dlqdefaults.Register()

	cfg := runtimeconfig.RuntimeConfig{DLQProvider: "file", DLQFile: "p"}
	if _, ok := dlq.ProviderFromConfig(cfg, nil).(*filedlq.DLQ); !ok {
		t.Fatalf("expected *file.DLQ")
	}

	cfg = runtimeconfig.RuntimeConfig{DLQProvider: "dir", DLQFile: "d"}
	if _, ok := dlq.ProviderFromConfig(cfg, nil).(*dirdlq.DLQ); !ok {
		t.Fatalf("expected *dir.DLQ")
	}

	cfg = runtimeconfig.RuntimeConfig{DLQProvider: "db"}
	if _, ok := dlq.ProviderFromConfig(cfg, (&dbpkg.Queries{})).(dbdlq.DLQ); !ok {
		t.Fatalf("expected db.DLQ")
	}

	cfg = runtimeconfig.RuntimeConfig{DLQProvider: "email"}
	if p := dlq.ProviderFromConfig(cfg, nil); reflect.TypeOf(p) != reflect.TypeOf(emaildlq.DLQ{}) && reflect.TypeOf(p) != reflect.TypeOf(dlq.LogDLQ{}) {
		t.Fatalf("unexpected type %T", p)
	}

	cfg = runtimeconfig.RuntimeConfig{DLQProvider: "db,log"}
	if _, ok := dlq.ProviderFromConfig(cfg, (&dbpkg.Queries{})).(dlq.MultiDLQ); !ok {
		t.Fatalf("expected MultiDLQ")
	}
}

func TestRegisterProviderCustom(t *testing.T) {
	called := false
	dlq.RegisterProvider("custom", func(cfg runtimeconfig.RuntimeConfig, q *dbpkg.Queries) dlq.DLQ {
		called = true
		return dlq.LogDLQ{}
	})

	cfg := runtimeconfig.RuntimeConfig{DLQProvider: "custom"}
	if _, ok := dlq.ProviderFromConfig(cfg, nil).(dlq.LogDLQ); !ok || !called {
		t.Fatalf("custom provider not used")
	}
}

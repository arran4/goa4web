# Configuration

## Why this package exists

`config` is the typed boundary between deployment inputs and application code. It
keeps flag, config-file, and environment handling in one place so handlers and
workers consume a validated `RuntimeConfig` rather than reading process state.
Values are resolved in this order: command-line flag, config file, environment,
then a default when the value is still empty.

## Where to look

- `runtime.go` defines `RuntimeConfig`, option metadata, parsing, precedence, and
  validation. Start here when adding a setting.
- `env.go` is the canonical list of environment-variable names.
- `defaults.go` supplies defaults; use it only when empty is not meaningful.
- `examples/` contains user-facing configuration examples that must use the same
  keys.
- `internal/configformat` and `internal/configexplain` provide CLI presentation
  and diagnostics; update them when a new kind of value needs special handling.

## Adding a setting

Adding only a struct field does **not** populate it. A complete change normally:

1. Add the typed field to `RuntimeConfig` in `runtime.go`.
2. Add its environment key to `env.go`.
3. Register a `StringOption`, `IntOption`, or `BoolOption` in `runtime.go`. The
   option connects the flag/config key, environment key, and destination field.
4. Add a default only if the zero value is not the desired fallback.
5. Add the same key to the relevant file under `examples/`.
6. Extend precedence, validation, environment-map, and help-output tests as
   appropriate, then run the full test suite.

Use an existing option of the same type as the template; this preserves the
required flag > file > environment > default precedence. Never read `os.Getenv`
directly from feature code.

## Consuming configuration

Receive `*config.RuntimeConfig` through an existing constructor or options struct:

```go
import "github.com/arran4/goa4web/config"

type Service struct { cfg *config.RuntimeConfig }
func NewService(cfg *config.RuntimeConfig) *Service { return &Service{cfg: cfg} }
```

Runtime construction belongs at the executable/application boundary. Packages
should not create a second config or depend on global state, because that bypasses
validation and makes tests order-dependent.

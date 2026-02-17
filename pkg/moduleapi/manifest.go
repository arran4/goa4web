package moduleapi

import (
	"context"

	"github.com/arran4/goa4web/internal/router"
	"github.com/arran4/goa4web/internal/tasks"
)

// Manifest describes a module that can register routes and tasks with goa4web.
type Manifest struct {
	// Name is the module identifier used in logs and diagnostics.
	Name string
	// RegisterRoutes wires the module's HTTP routes into the router registry.
	RegisterRoutes func(reg *router.Registry)
	// RegisterTasks wires the module's tasks into the task registry.
	RegisterTasks func(reg *tasks.Registry)
	// OnStart runs optional module startup logic when the application starts.
	OnStart func(ctx context.Context) error
	// OnStop runs optional module shutdown logic when the application stops.
	OnStop func(ctx context.Context) error
}

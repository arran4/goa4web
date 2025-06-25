package forum

import (
	"context"

	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/runtimeconfig"
)

func notifyChange(ctx context.Context, provider email.Provider, emailAddr, page string) error {
	return nil
}

func getEmailProvider() email.Provider {
	return email.ProviderFromConfig(runtimeconfig.AppRuntimeConfig)
}

func notifyAdmins(ctx context.Context, provider email.Provider, q *Queries, page string) {}

func notifyThreadSubscribers(ctx context.Context, provider email.Provider, q *Queries, threadID, excludeUser int32, page string) {
}

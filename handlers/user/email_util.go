package user

import (
	"context"
	"strings"

	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/runtimeconfig"
)

func getEmailProvider() email.Provider {
	switch strings.ToLower(runtimeconfig.AppRuntimeConfig.EmailProvider) {
	case "log":
		return email.LogProvider{}
	default:
		return nil
	}
}

func notifyChange(ctx context.Context, provider email.Provider, to, page string) error {
	if provider == nil {
		return nil
	}
	subject := "Website Update Notification"
	return provider.Send(ctx, to, subject, page, "")
}

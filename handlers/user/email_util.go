package user

import (
	"context"

	"github.com/arran4/goa4web/internal/email"
)

func notifyChange(ctx context.Context, provider email.Provider, to, page string) error {
	if provider == nil {
		return nil
	}
	subject := "Website Update Notification"
	return provider.Send(ctx, to, subject, page, "")
}

package emailutil

import (
	"context"
	_ "embed"

	"github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
)

// defaultUpdateEmailText contains the compiled-in notification email template.
// Administrators may override it by saving a new body in the template_overrides table.
//
//go:embed updateEmail.txt
var defaultUpdateEmailText string

// getUpdateEmailText returns the update email template body, preferring a database
// override when available.
func getUpdateEmailText(ctx context.Context) string {
	if q, ok := ctx.Value(common.KeyQueries).(*db.Queries); ok && q != nil {
		if body, err := q.GetTemplateOverride(ctx, "updateEmail"); err == nil && body != "" {
			return body
		}
	}
	return defaultUpdateEmailText
}

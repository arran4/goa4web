package emailutil

import (
	"context"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
)

// defaultUpdateEmailText contains the compiled-in notification email template.
// Administrators may override it by saving a new body in the template_overrides table.
var (
	defaultUpdateEmailText = templates.UpdateEmailText
	defaultUpdateEmailHTML = templates.UpdateEmailHTML

	defaultThreadEmailText        = templates.ThreadEmailText
	defaultThreadEmailHTML        = templates.ThreadEmailHTML
	defaultBlogEmailText          = templates.BlogEmailText
	defaultBlogEmailHTML          = templates.BlogEmailHTML
	defaultWritingEmailText       = templates.WritingEmailText
	defaultWritingEmailHTML       = templates.WritingEmailHTML
	defaultSignupEmailText        = templates.SignupEmailText
	defaultSignupEmailHTML        = templates.SignupEmailHTML
	defaultVerificationEmailText  = templates.VerificationEmailText
	defaultVerificationEmailHTML  = templates.VerificationEmailHTML
	defaultPasswordResetEmailText = templates.PasswordResetEmailText
	defaultPasswordResetEmailHTML = templates.PasswordResetEmailHTML
	defaultUserApprovedEmailText  = templates.UserApprovedEmailText
	defaultUserApprovedEmailHTML  = templates.UserApprovedEmailHTML
	defaultUserRejectedEmailText  = templates.UserRejectedEmailText
	defaultUserRejectedEmailHTML  = templates.UserRejectedEmailHTML
)

// getUpdateEmailText returns the update email template body, preferring a database
// override when available.
// GetUpdateEmailText returns the update email template body, preferring a database
// override when available.
func GetUpdateEmailText(ctx context.Context) string {
	if q, ok := ctx.Value(common.KeyQueries).(*db.Queries); ok && q != nil {
		if body, err := q.GetTemplateOverride(ctx, "updateEmail"); err == nil && body != "" {
			return body
		}
	}
	return defaultUpdateEmailText
}

// defaultReplyEmail* contain the compiled-in forum reply notification
// templates. Administrators may override them in the template_overrides table.
var (
	defaultReplyEmailText = templates.ReplyEmailText
	defaultReplyEmailHTML = templates.ReplyEmailHTML
)

package admin

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// AcceptRequestTask accepts a queued request.
type AcceptRequestTask struct{ tasks.TaskString }

var acceptRequestTask = &AcceptRequestTask{TaskString: TaskAccept}

var _ tasks.Task = (*AcceptRequestTask)(nil)
var _ tasks.AuditableTask = (*AcceptRequestTask)(nil)

func (AcceptRequestTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	if req := cd.CurrentRequest(); req != nil {
		if req.ChangeTable == "users" && req.ChangeField == "password_reset" {
			queries := cd.Queries()
			expiry := time.Now().Add(-time.Duration(cd.Config.PasswordResetExpiryHours) * time.Hour)
			reset, err := queries.GetPasswordResetByUser(r.Context(), db.GetPasswordResetByUserParams{
				UserID:    req.UsersIdusers,
				CreatedAt: expiry,
			})
			if err != nil {
				return fmt.Errorf("fetch password reset: %w", err)
			}
			if !reset.Passwd.Valid {
				return fmt.Errorf("invalid password reset: missing password")
			}
			if err := queries.InsertPassword(r.Context(), db.InsertPasswordParams{
				UsersIdusers:    req.UsersIdusers,
				Passwd:          reset.Passwd.String,
				PasswdAlgorithm: sql.NullString{String: reset.PasswdAlgorithm.String, Valid: true},
			}); err != nil {
				return fmt.Errorf("update password: %w", err)
			}
			if err := queries.SystemDeletePasswordReset(r.Context(), reset.ID); err != nil {
				log.Printf("delete reset: %v", err)
			}
			if err := queries.AdminInsertRequestComment(r.Context(), db.AdminInsertRequestCommentParams{
				RequestID: req.ID,
				Comment:   "Password updated from pending request",
			}); err != nil {
				log.Printf("insert comment: %v", err)
			}
		}
	}
	handleRequestAction(w, r, "accepted")
	return nil
}

// AuditRecord summarises a request queue action.
func (AcceptRequestTask) AuditRecord(data map[string]any) string {
	return requestAuditSummary("accepted", data)
}

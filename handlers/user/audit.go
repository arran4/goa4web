package user

import (
	"log"
	"net/http"

	"github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
)

// logAudit records an administrative action in the audit_log table.
func logAudit(r *http.Request, action string) {
	queries, qok := r.Context().Value(common.KeyQueries).(*db.Queries)
	cd, cok := r.Context().Value(common.KeyCoreData).(*common.CoreData)
	if !qok || !cok || cd == nil {
		return
	}
	if err := queries.InsertAuditLog(r.Context(), db.InsertAuditLogParams{
		UsersIdusers: cd.UserID,
		Action:       action,
	}); err != nil {
		log.Printf("insert audit log: %v", err)
	}
}

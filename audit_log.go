package goa4web

import (
	"github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"
)

// logAudit records an administrative action in the audit_log table.
func logAudit(r *http.Request, action string) {
	queries, qok := r.Context().Value(common.KeyQueries).(*Queries)
	cd, cok := r.Context().Value(common.KeyCoreData).(*CoreData)
	if !qok || !cok || cd == nil {
		return
	}
	if err := queries.InsertAuditLog(r.Context(), InsertAuditLogParams{
		UsersIdusers: cd.UserID,
		Action:       action,
	}); err != nil {
		log.Printf("insert audit log: %v", err)
	}
}

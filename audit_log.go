package main

import (
	"log"
	"net/http"
)

// logAudit records an administrative action in the audit_log table.
func logAudit(r *http.Request, action string) {
	queries, qok := r.Context().Value(ContextValues("queries")).(*Queries)
	cd, cok := r.Context().Value(ContextValues("coreData")).(*CoreData)
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

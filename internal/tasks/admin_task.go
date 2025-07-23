package tasks

// AuditableTask marks a sensitive task that should be recorded in the audit log.
// Implementations receive the event data map and should return a human readable
// summary describing the action.
type AuditableTask interface {
	AuditRecord(data map[string]any) string
}

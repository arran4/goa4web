package tasks

// AdminTask marks tasks restricted to administrators.
type AdminTask interface {
	IsAdminTask() bool
}

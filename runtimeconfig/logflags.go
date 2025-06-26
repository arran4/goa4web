package runtimeconfig

// Log flags for selectively enabling request logs.
const (
	// LogFlagAccess enables logging of access checks and permission issues.
	LogFlagAccess = 1 << iota
	// LogFlagDB enables verbose database access logs.
	LogFlagDB
	// LogFlagErrors enables extra error logs for debugging.
	LogFlagErrors
	// LogFlagDebug enables miscellaneous debug logs.
	LogFlagDebug
	// LogFlagAuth enables logging of authentication events.
	LogFlagAuth
)

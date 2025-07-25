package email

const (
	// LogLevelSummary logs the recipient and subject of each sent email.
	LogLevelSummary = iota
	// LogLevelBody logs the full body of each sent email.
	LogLevelBody
)

// LogVerbosity controls the amount of logging performed by email providers.
// Set to LogLevelSummary to log basic info or LogLevelBody for verbose output.
// The effective level is determined by RuntimeConfig.

package runtimeconfig

import "github.com/arran4/goa4web/config"

// StringOption describes a string based runtime configuration flag.
type StringOption struct {
	Name          string
	Env           string
	Usage         string
	Default       string
	ExtendedUsage string
	Target        func(*RuntimeConfig) *string
}

// IntOption describes an integer runtime configuration flag.
type IntOption struct {
	Name          string
	Env           string
	Usage         string
	Default       int
	ExtendedUsage string
	Target        func(*RuntimeConfig) *int
}

// BoolOption describes a boolean runtime configuration flag.
type BoolOption struct {
	Name          string
	Env           string
	Usage         string
	Default       bool
	ExtendedUsage string
	Target        func(*RuntimeConfig) *bool
}

// StringOptions lists the string runtime options shared by flag parsing and configuration generation.
var StringOptions = []StringOption{
	{"db-conn", config.EnvDBConn, "database connection string", "", "db_conn.txt", func(c *RuntimeConfig) *string { return &c.DBConn }},
	{"db-driver", config.EnvDBDriver, "database driver", "mysql", "db_driver.txt", func(c *RuntimeConfig) *string { return &c.DBDriver }},
	{"listen", config.EnvListen, "server listen address", ":8080", "", func(c *RuntimeConfig) *string { return &c.HTTPListen }},
	{"hostname", config.EnvHostname, "server base URL", "", "", func(c *RuntimeConfig) *string { return &c.HTTPHostname }},
	{"email-provider", config.EnvEmailProvider, "email provider", "", "", func(c *RuntimeConfig) *string { return &c.EmailProvider }},
	{"smtp-host", config.EnvSMTPHost, "SMTP host", "", "", func(c *RuntimeConfig) *string { return &c.EmailSMTPHost }},
	{"smtp-port", config.EnvSMTPPort, "SMTP port", "", "", func(c *RuntimeConfig) *string { return &c.EmailSMTPPort }},
	{"smtp-user", config.EnvSMTPUser, "SMTP user", "", "", func(c *RuntimeConfig) *string { return &c.EmailSMTPUser }},
	{"smtp-pass", config.EnvSMTPPass, "SMTP pass", "", "", func(c *RuntimeConfig) *string { return &c.EmailSMTPPass }},
	{"smtp-auth", config.EnvSMTPAuth, "SMTP auth method", "", "", func(c *RuntimeConfig) *string { return &c.EmailSMTPAuth }},
	{"email-from", config.EnvEmailFrom, "default From address", "", "", func(c *RuntimeConfig) *string { return &c.EmailFrom }},
	{"aws-region", config.EnvAWSRegion, "AWS region", "", "", func(c *RuntimeConfig) *string { return &c.EmailAWSRegion }},
	{"jmap-endpoint", config.EnvJMAPEndpoint, "JMAP endpoint", "", "", func(c *RuntimeConfig) *string { return &c.EmailJMAPEndpoint }},
	{"jmap-account", config.EnvJMAPAccount, "JMAP account", "", "", func(c *RuntimeConfig) *string { return &c.EmailJMAPAccount }},
	{"jmap-identity", config.EnvJMAPIdentity, "JMAP identity", "", "", func(c *RuntimeConfig) *string { return &c.EmailJMAPIdentity }},
	{"jmap-user", config.EnvJMAPUser, "JMAP user", "", "", func(c *RuntimeConfig) *string { return &c.EmailJMAPUser }},
	{"jmap-pass", config.EnvJMAPPass, "JMAP pass", "", "", func(c *RuntimeConfig) *string { return &c.EmailJMAPPass }},
	{"sendgrid-key", config.EnvSendGridKey, "SendGrid API key", "", "", func(c *RuntimeConfig) *string { return &c.EmailSendGridKey }},
	{"default-language", config.EnvDefaultLanguage, "default language name", "", "", func(c *RuntimeConfig) *string { return &c.DefaultLanguage }},
	{"image-upload-dir", config.EnvImageUploadDir, "directory to store uploaded images", "", "", func(c *RuntimeConfig) *string { return &c.ImageUploadDir }},
	{"dlq-provider", config.EnvDLQProvider, "dead letter queue provider", "", "", func(c *RuntimeConfig) *string { return &c.DLQProvider }},
	{"dlq-file", config.EnvDLQFile, "dead letter queue file path", "", "", func(c *RuntimeConfig) *string { return &c.DLQFile }},
	{"admin-emails", config.EnvAdminEmails, "administrator email addresses", "", "", func(c *RuntimeConfig) *string { return &c.AdminEmails }},
}

// IntOptions lists the integer runtime options shared by flag parsing and configuration generation.
var IntOptions = []IntOption{
	{"db-log-verbosity", config.EnvDBLogVerbosity, "database logging verbosity", 0, "", func(c *RuntimeConfig) *int { return &c.DBLogVerbosity }},
	{"log-flags", config.EnvLogFlags, "request logging flags", 0, "", func(c *RuntimeConfig) *int { return &c.LogFlags }},
	{"page-size-min", config.EnvPageSizeMin, "minimum allowed page size", 0, "", func(c *RuntimeConfig) *int { return &c.PageSizeMin }},
	{"page-size-max", config.EnvPageSizeMax, "maximum allowed page size", 0, "", func(c *RuntimeConfig) *int { return &c.PageSizeMax }},
	{"page-size-default", config.EnvPageSizeDefault, "default page size", 0, "", func(c *RuntimeConfig) *int { return &c.PageSizeDefault }},
	{"image-max-bytes", config.EnvImageMaxBytes, "maximum allowed upload size in bytes", 0, "", func(c *RuntimeConfig) *int { return &c.ImageMaxBytes }},
	{"email-worker-interval", config.EnvEmailWorkerInterval, "interval in seconds between email worker runs", 0, "", func(c *RuntimeConfig) *int { return &c.EmailWorkerInterval }},
	{"stats-start-year", config.EnvStatsStartYear, "start year for usage stats", 0, "", func(c *RuntimeConfig) *int { return &c.StatsStartYear }},
}

// BoolOptions lists the boolean runtime options shared by flag parsing and configuration generation.
var BoolOptions = []BoolOption{
	{"feeds-enabled", config.EnvFeedsEnabled, "enable or disable feeds", true, "", func(c *RuntimeConfig) *bool { return &c.FeedsEnabled }},
	{"smtp-starttls", config.EnvSMTPStartTLS, "enable or disable STARTTLS", true, "", func(c *RuntimeConfig) *bool { return &c.EmailSMTPStartTLS }},
	{"", config.EnvEmailEnabled, "enable sending queued emails", true, "", func(c *RuntimeConfig) *bool { return &c.EmailEnabled }},
	{"", config.EnvNotificationsEnabled, "enable internal notifications", true, "", func(c *RuntimeConfig) *bool { return &c.NotificationsEnabled }},
	{"", config.EnvCSRFEnabled, "enable or disable CSRF protection", true, "", func(c *RuntimeConfig) *bool { return &c.CSRFEnabled }},
	{"", config.EnvAdminNotify, "enable admin notification emails", true, "", func(c *RuntimeConfig) *bool { return &c.AdminNotify }},
}

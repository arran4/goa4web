package runtimeconfig

import "github.com/arran4/goa4web/config"

// StringOption describes a string based runtime configuration flag.
type StringOption struct {
	Name          string
	Env           string
	Usage         string
	Default       string
	Examples      []string
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
	{"db-conn", config.EnvDBConn, "database connection string", "", nil, "db_conn.txt", func(c *RuntimeConfig) *string { return &c.DBConn }},
	{"db-driver", config.EnvDBDriver, "database driver", "mysql", nil, "db_driver.txt", func(c *RuntimeConfig) *string { return &c.DBDriver }},
	{"listen", config.EnvListen, "server listen address", ":8080", nil, "", func(c *RuntimeConfig) *string { return &c.HTTPListen }},
	{"hostname", config.EnvHostname, "server base URL", "", nil, "", func(c *RuntimeConfig) *string { return &c.HTTPHostname }},
	{"email-provider", config.EnvEmailProvider, "email provider", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailProvider }},
	{"smtp-host", config.EnvSMTPHost, "SMTP host", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailSMTPHost }},
	{"smtp-port", config.EnvSMTPPort, "SMTP port", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailSMTPPort }},
	{"smtp-user", config.EnvSMTPUser, "SMTP user", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailSMTPUser }},
	{"smtp-pass", config.EnvSMTPPass, "SMTP pass", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailSMTPPass }},
	{"smtp-auth", config.EnvSMTPAuth, "SMTP auth method", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailSMTPAuth }},
	{"email-from", config.EnvEmailFrom, "default From address", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailFrom }},
	{"aws-region", config.EnvAWSRegion, "AWS region", "", []string{"us-east-1"}, "", func(c *RuntimeConfig) *string { return &c.EmailAWSRegion }},
	{"jmap-endpoint", config.EnvJMAPEndpoint, "JMAP endpoint", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailJMAPEndpoint }},
	{"jmap-account", config.EnvJMAPAccount, "JMAP account", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailJMAPAccount }},
	{"jmap-identity", config.EnvJMAPIdentity, "JMAP identity", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailJMAPIdentity }},
	{"jmap-user", config.EnvJMAPUser, "JMAP user", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailJMAPUser }},
	{"jmap-pass", config.EnvJMAPPass, "JMAP pass", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailJMAPPass }},
	{"sendgrid-key", config.EnvSendGridKey, "SendGrid API key", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailSendGridKey }},
	{"default-language", config.EnvDefaultLanguage, "default language name", "", nil, "", func(c *RuntimeConfig) *string { return &c.DefaultLanguage }},
	{"image-upload-dir", config.EnvImageUploadDir, "directory to store uploaded images when using the local provider", "", nil, "", func(c *RuntimeConfig) *string { return &c.ImageUploadDir }},
	{"image-upload-provider", config.EnvImageUploadProvider, "image upload provider", "local", nil, "", func(c *RuntimeConfig) *string { return &c.ImageUploadProvider }},
	{"image-upload-s3-url", config.EnvImageUploadS3URL, "S3 prefix URL for uploads", "", []string{"s3://mybucket/uploads", "s3://bucket/images"}, "", func(c *RuntimeConfig) *string { return &c.ImageUploadS3URL }},
	{"image-cache-provider", config.EnvImageCacheProvider, "image cache provider", "local", nil, "", func(c *RuntimeConfig) *string { return &c.ImageCacheProvider }},
	{"image-cache-s3-url", config.EnvImageCacheS3URL, "S3 prefix URL for cache", "", []string{"s3://mybucket/cache"}, "", func(c *RuntimeConfig) *string { return &c.ImageCacheS3URL }},
	{"image-cache-dir", config.EnvImageCacheDir, "directory for cached thumbnails when using the local provider", "", nil, "", func(c *RuntimeConfig) *string { return &c.ImageCacheDir }},
	{"dlq-provider", config.EnvDLQProvider, "dead letter queue provider", "", nil, "", func(c *RuntimeConfig) *string { return &c.DLQProvider }},
	{"dlq-file", config.EnvDLQFile, "dead letter queue file path", "", nil, "", func(c *RuntimeConfig) *string { return &c.DLQFile }},
	{"admin-emails", config.EnvAdminEmails, "administrator email addresses", "", nil, "", func(c *RuntimeConfig) *string { return &c.AdminEmails }},
	{"session-secret", config.EnvSessionSecret, "session secret key", "", nil, "", func(c *RuntimeConfig) *string { return &c.SessionSecret }},
	{"session-secret-file", config.EnvSessionSecretFile, "path to session secret file", "", nil, "", func(c *RuntimeConfig) *string { return &c.SessionSecretFile }},
	{"image-sign-secret", config.EnvImageSignSecret, "image signing key", "", nil, "", func(c *RuntimeConfig) *string { return &c.ImageSignSecret }},
	{"image-sign-secret-file", config.EnvImageSignSecretFile, "path to image signing key", "", nil, "", func(c *RuntimeConfig) *string { return &c.ImageSignSecretFile }},
}

// IntOptions lists the integer runtime options shared by flag parsing and configuration generation.
var IntOptions = []IntOption{
	{"db-log-verbosity", config.EnvDBLogVerbosity, "database logging verbosity", 0, "", func(c *RuntimeConfig) *int { return &c.DBLogVerbosity }},
	{"log-flags", config.EnvLogFlags, "request logging flags", 0, "", func(c *RuntimeConfig) *int { return &c.LogFlags }},
	{"page-size-min", config.EnvPageSizeMin, "minimum allowed page size", 0, "", func(c *RuntimeConfig) *int { return &c.PageSizeMin }},
	{"page-size-max", config.EnvPageSizeMax, "maximum allowed page size", 0, "", func(c *RuntimeConfig) *int { return &c.PageSizeMax }},
	{"page-size-default", config.EnvPageSizeDefault, "default page size", 0, "", func(c *RuntimeConfig) *int { return &c.PageSizeDefault }},
	{"image-max-bytes", config.EnvImageMaxBytes, "maximum allowed upload size in bytes", 0, "", func(c *RuntimeConfig) *int { return &c.ImageMaxBytes }},
	{"image-cache-max-bytes", config.EnvImageCacheMaxBytes, "maximum image cache size in bytes", -1, "", func(c *RuntimeConfig) *int { return &c.ImageCacheMaxBytes }},
	{"email-worker-interval", config.EnvEmailWorkerInterval, "interval in seconds between email worker runs", 0, "", func(c *RuntimeConfig) *int { return &c.EmailWorkerInterval }},
	{"stats-start-year", config.EnvStatsStartYear, "start year for usage stats", 0, "", func(c *RuntimeConfig) *int { return &c.StatsStartYear }},
}

// BoolOptions lists the boolean runtime options shared by flag parsing and configuration generation.
var BoolOptions = []BoolOption{
	{"feeds-enabled", config.EnvFeedsEnabled, "enable or disable feeds", true, "", func(c *RuntimeConfig) *bool { return &c.FeedsEnabled }},
	{"smtp-starttls", config.EnvSMTPStartTLS, "enable or disable STARTTLS", true, "", func(c *RuntimeConfig) *bool { return &c.EmailSMTPStartTLS }},
	{"email-enabled", config.EnvEmailEnabled, "enable sending queued emails", true, "", func(c *RuntimeConfig) *bool { return &c.EmailEnabled }},
	{"notifications-enabled", config.EnvNotificationsEnabled, "enable internal notifications", true, "", func(c *RuntimeConfig) *bool { return &c.NotificationsEnabled }},
	{"csrf-enabled", config.EnvCSRFEnabled, "enable or disable CSRF protection", true, "", func(c *RuntimeConfig) *bool { return &c.CSRFEnabled }},
	{"admin-notify", config.EnvAdminNotify, "enable admin notification emails", true, "", func(c *RuntimeConfig) *bool { return &c.AdminNotify }},
	{"create-dirs", config.EnvCreateDirs, "create missing directories", false, "", func(c *RuntimeConfig) *bool { return &c.CreateDirs }},
}

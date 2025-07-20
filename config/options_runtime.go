package config

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
	{"db-conn", EnvDBConn, "database connection string", "", nil, "db_conn.txt", func(c *RuntimeConfig) *string { return &c.DBConn }},
	{"db-driver", EnvDBDriver, "database driver", "mysql", nil, "db_driver.txt", func(c *RuntimeConfig) *string { return &c.DBDriver }},
	{"listen", EnvListen, "server listen address", ":8080", nil, "", func(c *RuntimeConfig) *string { return &c.HTTPListen }},
	{"hostname", EnvHostname, "server base URL", "", nil, "", func(c *RuntimeConfig) *string { return &c.HTTPHostname }},
	{"email-provider", EnvEmailProvider, "email provider", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailProvider }},
	{"smtp-host", EnvSMTPHost, "SMTP host", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailSMTPHost }},
	{"smtp-port", EnvSMTPPort, "SMTP port", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailSMTPPort }},
	{"smtp-user", EnvSMTPUser, "SMTP user", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailSMTPUser }},
	{"smtp-pass", EnvSMTPPass, "SMTP pass", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailSMTPPass }},
	{"smtp-auth", EnvSMTPAuth, "SMTP auth method", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailSMTPAuth }},
	{"email-from", EnvEmailFrom, "default From address", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailFrom }},
	{"email-subject-prefix", EnvEmailSubjectPrefix, "email subject prefix", "goa4web", nil, "", func(c *RuntimeConfig) *string { return &c.EmailSubjectPrefix }},
	{"email-signoff", EnvEmailSignOff, "email sign off", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailSignOff }},
	{"aws-region", EnvAWSRegion, "AWS region", "", []string{"us-east-1"}, "", func(c *RuntimeConfig) *string { return &c.EmailAWSRegion }},
	{"jmap-endpoint", EnvJMAPEndpoint, "JMAP endpoint", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailJMAPEndpoint }},
	{"jmap-account", EnvJMAPAccount, "JMAP account", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailJMAPAccount }},
	{"jmap-identity", EnvJMAPIdentity, "JMAP identity", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailJMAPIdentity }},
	{"jmap-user", EnvJMAPUser, "JMAP user", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailJMAPUser }},
	{"jmap-pass", EnvJMAPPass, "JMAP pass", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailJMAPPass }},
	{"sendgrid-key", EnvSendGridKey, "SendGrid API key", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailSendGridKey }},
	{"default-language", EnvDefaultLanguage, "default language name", "", nil, "", func(c *RuntimeConfig) *string { return &c.DefaultLanguage }},
	{"image-upload-dir", EnvImageUploadDir, "directory to store uploaded images when using the local provider", "", nil, "", func(c *RuntimeConfig) *string { return &c.ImageUploadDir }},
	{"image-upload-provider", EnvImageUploadProvider, "image upload provider", "local", nil, "", func(c *RuntimeConfig) *string { return &c.ImageUploadProvider }},
	{"image-upload-s3-url", EnvImageUploadS3URL, "S3 prefix URL for uploads", "", []string{"s3://mybucket/uploads", "s3://bucket/images"}, "", func(c *RuntimeConfig) *string { return &c.ImageUploadS3URL }},
	{"image-cache-provider", EnvImageCacheProvider, "image cache provider", "local", nil, "", func(c *RuntimeConfig) *string { return &c.ImageCacheProvider }},
	{"image-cache-s3-url", EnvImageCacheS3URL, "S3 prefix URL for cache", "", []string{"s3://mybucket/cache"}, "", func(c *RuntimeConfig) *string { return &c.ImageCacheS3URL }},
	{"image-cache-dir", EnvImageCacheDir, "directory for cached thumbnails when using the local provider", "", nil, "", func(c *RuntimeConfig) *string { return &c.ImageCacheDir }},
	{"dlq-provider", EnvDLQProvider, "dead letter queue provider", "", nil, "", func(c *RuntimeConfig) *string { return &c.DLQProvider }},
	{"dlq-file", EnvDLQFile, "dead letter queue file path", "", nil, "", func(c *RuntimeConfig) *string { return &c.DLQFile }},
	{"admin-emails", EnvAdminEmails, "administrator email addresses", "", nil, "", func(c *RuntimeConfig) *string { return &c.AdminEmails }},
	{"session-secret", EnvSessionSecret, "session secret key", "", nil, "", func(c *RuntimeConfig) *string { return &c.SessionSecret }},
	{"session-secret-file", EnvSessionSecretFile, "path to session secret file", "", nil, "", func(c *RuntimeConfig) *string { return &c.SessionSecretFile }},
	{"image-sign-secret", EnvImageSignSecret, "image signing key", "", nil, "", func(c *RuntimeConfig) *string { return &c.ImageSignSecret }},
	{"image-sign-secret-file", EnvImageSignSecretFile, "path to image signing key", "", nil, "", func(c *RuntimeConfig) *string { return &c.ImageSignSecretFile }},
}

// IntOptions lists the integer runtime options shared by flag parsing and configuration generation.
var IntOptions = []IntOption{
	{"db-log-verbosity", EnvDBLogVerbosity, "database logging verbosity", 0, "", func(c *RuntimeConfig) *int { return &c.DBLogVerbosity }},
	{"log-flags", EnvLogFlags, "request logging flags", 0, "", func(c *RuntimeConfig) *int { return &c.LogFlags }},
	{"page-size-min", EnvPageSizeMin, "minimum allowed page size", 0, "", func(c *RuntimeConfig) *int { return &c.PageSizeMin }},
	{"page-size-max", EnvPageSizeMax, "maximum allowed page size", 0, "", func(c *RuntimeConfig) *int { return &c.PageSizeMax }},
	{"page-size-default", EnvPageSizeDefault, "default page size", 0, "", func(c *RuntimeConfig) *int { return &c.PageSizeDefault }},
	{"image-max-bytes", EnvImageMaxBytes, "maximum allowed upload size in bytes", 0, "", func(c *RuntimeConfig) *int { return &c.ImageMaxBytes }},
	{"image-cache-max-bytes", EnvImageCacheMaxBytes, "maximum image cache size in bytes", -1, "", func(c *RuntimeConfig) *int { return &c.ImageCacheMaxBytes }},
	{"email-worker-interval", EnvEmailWorkerInterval, "interval in seconds between email worker runs", 0, "", func(c *RuntimeConfig) *int { return &c.EmailWorkerInterval }},
	{"stats-start-year", EnvStatsStartYear, "start year for usage stats", 0, "", func(c *RuntimeConfig) *int { return &c.StatsStartYear }},
}

// BoolOptions lists the boolean runtime options shared by flag parsing and configuration generation.
var BoolOptions = []BoolOption{
	{"feeds-enabled", EnvFeedsEnabled, "enable or disable feeds", true, "", func(c *RuntimeConfig) *bool { return &c.FeedsEnabled }},
	{"smtp-starttls", EnvSMTPStartTLS, "enable or disable STARTTLS", true, "", func(c *RuntimeConfig) *bool { return &c.EmailSMTPStartTLS }},
	{"email-enabled", EnvEmailEnabled, "enable sending queued emails", true, "", func(c *RuntimeConfig) *bool { return &c.EmailEnabled }},
	{"notifications-enabled", EnvNotificationsEnabled, "enable internal notifications", true, "", func(c *RuntimeConfig) *bool { return &c.NotificationsEnabled }},
	{"csrf-enabled", EnvCSRFEnabled, "enable or disable CSRF protection", true, "", func(c *RuntimeConfig) *bool { return &c.CSRFEnabled }},
	{"admin-notify", EnvAdminNotify, "enable admin notification emails", true, "", func(c *RuntimeConfig) *bool { return &c.AdminNotify }},
	{"create-dirs", EnvCreateDirs, "create missing directories", false, "", func(c *RuntimeConfig) *bool { return &c.CreateDirs }},
}

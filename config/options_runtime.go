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
	{"db-conn", EnvDBConn, "Database connection string. This is used to connect to the database.", "", nil, "db_conn.txt", func(c *RuntimeConfig) *string { return &c.DBConn }},
	{"db-driver", EnvDBDriver, "Database driver to use. Supported drivers are 'mysql'.", "mysql", nil, "db_driver.txt", func(c *RuntimeConfig) *string { return &c.DBDriver }},
	{"db-timezone", EnvDBTimezone, "Timezone for the database connection.", "Australia/Melbourne", nil, "", func(c *RuntimeConfig) *string { return &c.DBTimezone }},
	{"db-host", EnvDBHost, "Database server hostname.", "", nil, "", func(c *RuntimeConfig) *string { return &c.DBHost }},
	{"db-port", EnvDBPort, "Database server port.", "", nil, "", func(c *RuntimeConfig) *string { return &c.DBPort }},
	{"db-user", EnvDBUser, "Database username.", "", nil, "", func(c *RuntimeConfig) *string { return &c.DBUser }},
	{"db-pass", EnvDBPass, "Database password.", "", nil, "", func(c *RuntimeConfig) *string { return &c.DBPass }},
	{"db-name", EnvDBName, "Database name.", "", nil, "", func(c *RuntimeConfig) *string { return &c.DBName }},
	{"listen", EnvListen, "The address and port for the HTTP server to listen on.", ":8080", nil, "", func(c *RuntimeConfig) *string { return &c.HTTPListen }},
	{"hostname", EnvHostname, "The base URL of the server, used for generating absolute links.", "", nil, "", func(c *RuntimeConfig) *string { return &c.HTTPHostname }},
	{"hsts-header", EnvHSTSHeader, "The value for the Strict-Transport-Security header.", "max-age=63072000; includeSubDomains", nil, "", func(c *RuntimeConfig) *string { return &c.HSTSHeaderValue }},
	{"email-provider", EnvEmailProvider, "The email provider to use. Supported providers are 'smtp', 'ses', 'jmap', and 'log'.", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailProvider }},
	{"smtp-host", EnvSMTPHost, "The hostname of the SMTP server.", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailSMTPHost }},
	{"smtp-port", EnvSMTPPort, "The port of the SMTP server.", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailSMTPPort }},
	{"smtp-user", EnvSMTPUser, "The username for the SMTP server.", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailSMTPUser }},
	{"smtp-pass", EnvSMTPPass, "The password for the SMTP server.", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailSMTPPass }},
	{"smtp-auth", EnvSMTPAuth, "The authentication method for the SMTP server.", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailSMTPAuth }},
	{"email-from", EnvEmailFrom, "The default 'From' address for outgoing emails.", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailFrom }},
	{"email-subject-prefix", EnvEmailSubjectPrefix, "The prefix to add to the subject of all outgoing emails.", "goa4web", nil, "", func(c *RuntimeConfig) *string { return &c.EmailSubjectPrefix }},
	{"email-signoff", EnvEmailSignOff, "A sign-off message to append to the end of all outgoing emails.", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailSignOff }},
	{"aws-region", EnvAWSRegion, "The AWS region to use for SES.", "", []string{"us-east-1"}, "", func(c *RuntimeConfig) *string { return &c.EmailAWSRegion }},
	{"jmap-endpoint", EnvJMAPEndpoint, "The endpoint for the JMAP server.", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailJMAPEndpoint }},
	{"jmap-account", EnvJMAPAccount, "The account to use for the JMAP server. When omitted the primary mail account from the JMAP session is used.", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailJMAPAccount }},
	{"jmap-identity", EnvJMAPIdentity, "The identity to use for the JMAP server. When omitted the default mail identity from the JMAP session is used.", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailJMAPIdentity }},
	{"jmap-user", EnvJMAPUser, "The username for the JMAP server.", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailJMAPUser }},
	{"jmap-pass", EnvJMAPPass, "The password for the JMAP server.", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailJMAPPass }},
	{"sendgrid-key", EnvSendGridKey, "The API key for the SendGrid service.", "", nil, "", func(c *RuntimeConfig) *string { return &c.EmailSendGridKey }},
	{"default-language", EnvDefaultLanguage, "The default language for the application.", "", nil, "", func(c *RuntimeConfig) *string { return &c.DefaultLanguage }},
	{"timezone", EnvTimezone, "The default timezone for the application.", "Australia/Melbourne", nil, "", func(c *RuntimeConfig) *string { return &c.Timezone }},
	{"templates-dir", EnvTemplatesDir, "The directory to load templates from. If not specified, the embedded templates will be used.", "", nil, "", func(c *RuntimeConfig) *string { return &c.TemplatesDir }},
	{"migrations-dir", EnvMigrationsDir, "The directory to load migrations from at runtime.", "", nil, "", func(c *RuntimeConfig) *string { return &c.MigrationsDir }},
	{"image-upload-dir", EnvImageUploadDir, "The directory to store uploaded images when using the 'local' provider.", "", nil, "", func(c *RuntimeConfig) *string { return &c.ImageUploadDir }},
	{"image-upload-provider", EnvImageUploadProvider, "The provider to use for image uploads. Supported providers are 'local' and 's3'.", "local", nil, "", func(c *RuntimeConfig) *string { return &c.ImageUploadProvider }},
	{"image-upload-s3-url", EnvImageUploadS3URL, "The S3 prefix URL for image uploads.", "", []string{"s3://mybucket/uploads", "s3://bucket/images"}, "", func(c *RuntimeConfig) *string { return &c.ImageUploadS3URL }},
	{"image-cache-provider", EnvImageCacheProvider, "The provider to use for the image cache. Supported providers are 'local' and 's3'.", "local", nil, "", func(c *RuntimeConfig) *string { return &c.ImageCacheProvider }},
	{"image-cache-s3-url", EnvImageCacheS3URL, "The S3 prefix URL for the image cache.", "", []string{"s3://mybucket/cache"}, "", func(c *RuntimeConfig) *string { return &c.ImageCacheS3URL }},
	{"image-cache-dir", EnvImageCacheDir, "The directory for cached thumbnails when using the 'local' provider.", "", nil, "", func(c *RuntimeConfig) *string { return &c.ImageCacheDir }},
	{"dlq-provider", EnvDLQProvider, "The provider for the dead letter queue. Supported providers are 'file' and 'memory'.", "", nil, "", func(c *RuntimeConfig) *string { return &c.DLQProvider }},
	{"dlq-file", EnvDLQFile, "The file path for the dead letter queue when using the 'file' provider.", "", nil, "", func(c *RuntimeConfig) *string { return &c.DLQFile }},
	{"session-name", EnvSessionName, "The name of the session cookie.", "my-session", nil, "", func(c *RuntimeConfig) *string { return &c.SessionName }},
	{"admin-emails", EnvAdminEmails, "A comma-separated list of email addresses for administrative notifications.", "", nil, "", func(c *RuntimeConfig) *string { return &c.AdminEmails }},
	{"session-secret", EnvSessionSecret, "The secret key used to encrypt session data.", "", nil, "", func(c *RuntimeConfig) *string { return &c.SessionSecret }},
	{"session-secret-file", EnvSessionSecretFile, "The path to a file containing the session secret key.", "", nil, "", func(c *RuntimeConfig) *string { return &c.SessionSecretFile }},
	{"session-same-site", EnvSessionSameSite, "The SameSite policy for the session cookie. Supported values are 'strict', 'lax', and 'none'.", "strict", nil, "", func(c *RuntimeConfig) *string { return &c.SessionSameSite }},
	{"image-sign-secret", EnvImageSignSecret, "The secret key used to sign image URLs.", "", nil, "", func(c *RuntimeConfig) *string { return &c.ImageSignSecret }},
	{"image-sign-secret-file", EnvImageSignSecretFile, "The path to a file containing the image signing key.", "", nil, "", func(c *RuntimeConfig) *string { return &c.ImageSignSecretFile }},
	{"image-sign-expiry", EnvImageSignExpiry, "The default expiry duration for image URLs.", "24h", nil, "", func(c *RuntimeConfig) *string { return &c.ImageSignExpiry }},
	{"link-sign-secret", EnvLinkSignSecret, "The secret key used to sign external link URLs.", "", nil, "", func(c *RuntimeConfig) *string { return &c.LinkSignSecret }},
	{"link-sign-secret-file", EnvLinkSignSecretFile, "The path to a file containing the external link signing key.", "", nil, "", func(c *RuntimeConfig) *string { return &c.LinkSignSecretFile }},
	{"link-sign-expiry", EnvLinkSignExpiry, "The default expiry duration for external link URLs.", "24h", nil, "", func(c *RuntimeConfig) *string { return &c.LinkSignExpiry }},
	{"share-sign-secret", EnvShareSignSecret, "The secret key used to sign share URLs.", "", nil, "", func(c *RuntimeConfig) *string { return &c.ShareSignSecret }},
	{"share-sign-secret-file", EnvShareSignSecretFile, "The path to a file containing the share signing key.", "", nil, "", func(c *RuntimeConfig) *string { return &c.ShareSignSecretFile }},
	{"share-sign-expiry", EnvShareSignExpiry, "The default expiry duration for share URLs.", "26280h", nil, "", func(c *RuntimeConfig) *string { return &c.ShareSignExpiry }},
	{"share-sign-expiry-login", EnvShareSignExpiryLogin, "The expiry duration for share URLs when a user is not logged in.", "1h", nil, "", func(c *RuntimeConfig) *string { return &c.ShareSignExpiryLogin }},
	{"admin-api-secret", EnvAdminAPISecret, "The secret key used to sign administrator API tokens.", "", nil, "", func(c *RuntimeConfig) *string { return &c.AdminAPISecret }},
	{"admin-api-secret-file", EnvAdminAPISecretFile, "The path to a file containing the administrator API signing key.", "", nil, "", func(c *RuntimeConfig) *string { return &c.AdminAPISecretFile }},
	{"twitter-site", EnvTwitterSite, "The Twitter handle for the site (e.g. @mysite).", "", nil, "", func(c *RuntimeConfig) *string { return &c.TwitterSite }},
}

// IntOptions lists the integer runtime options shared by flag parsing and configuration generation.
var IntOptions = []IntOption{
	{"db-log-verbosity", EnvDBLogVerbosity, "The verbosity level for database logging. 0 = off, 1 = errors, 2 = warnings, 3 = info, 4 = debug.", 0, "", func(c *RuntimeConfig) *int { return &c.DBLogVerbosity }},
	{"email-log-verbosity", EnvEmailLogVerbosity, "The verbosity level for email logging. 0 = off, 1 = errors, 2 = warnings, 3 = info, 4 = debug.", 0, "", func(c *RuntimeConfig) *int { return &c.EmailLogVerbosity }},
	{"log-flags", EnvLogFlags, "The flags for request logging.", 0, "", func(c *RuntimeConfig) *int { return &c.LogFlags }},
	{"page-size-min", EnvPageSizeMin, "The minimum allowed page size for paginated results.", 5, "", func(c *RuntimeConfig) *int { return &c.PageSizeMin }},
	{"page-size-max", EnvPageSizeMax, "The maximum allowed page size for paginated results.", 50, "", func(c *RuntimeConfig) *int { return &c.PageSizeMax }},
	{"page-size-default", EnvPageSizeDefault, "The default page size for paginated results.", DefaultPageSize, "", func(c *RuntimeConfig) *int { return &c.PageSizeDefault }},
	{"image-max-bytes", EnvImageMaxBytes, "The maximum allowed size for uploaded images in bytes.", 0, "", func(c *RuntimeConfig) *int { return &c.ImageMaxBytes }},
	{"image-cache-max-bytes", EnvImageCacheMaxBytes, "The maximum size of the image cache in bytes. A value of -1 means no limit.", -1, "", func(c *RuntimeConfig) *int { return &c.ImageCacheMaxBytes }},
	{"email-worker-interval", EnvEmailWorkerInterval, "The interval in seconds between runs of the email worker.", 0, "", func(c *RuntimeConfig) *int { return &c.EmailWorkerInterval }},
	{"email-verification-expiry-hours", EnvEmailVerificationExpiryHours, "The number of hours an email verification request is valid for.", 0, "", func(c *RuntimeConfig) *int { return &c.EmailVerificationExpiryHours }},
	{"password-reset-expiry-hours", EnvPasswordResetExpiryHours, "The number of hours a password reset request is valid for.", 0, "", func(c *RuntimeConfig) *int { return &c.PasswordResetExpiryHours }},
	{"login-attempt-window", EnvLoginAttemptWindow, "The window in minutes for tracking failed login attempts.", 15, "", func(c *RuntimeConfig) *int { return &c.LoginAttemptWindow }},
	{"login-attempt-threshold", EnvLoginAttemptThreshold, "The number of failed login attempts allowed within the window before throttling.", 5, "", func(c *RuntimeConfig) *int { return &c.LoginAttemptThreshold }},
	{"stats-start-year", EnvStatsStartYear, "The start year for usage statistics.", 2005, "", func(c *RuntimeConfig) *int { return &c.StatsStartYear }},
	{"og-image-width", EnvOGImageWidth, "The width of the generated Open Graph image.", 1200, "", func(c *RuntimeConfig) *int { return &c.OGImageWidth }},
	{"og-image-height", EnvOGImageHeight, "The height of the generated Open Graph image.", 630, "", func(c *RuntimeConfig) *int { return &c.OGImageHeight }},
}

// BoolOptions lists the boolean runtime options shared by flag parsing and configuration generation.
var BoolOptions = []BoolOption{
	{"feeds-enabled", EnvFeedsEnabled, "Enable or disable RSS/Atom feeds.", true, "", func(c *RuntimeConfig) *bool { return &c.FeedsEnabled }},
	{"smtp-starttls", EnvSMTPStartTLS, "Enable or disable STARTTLS for SMTP connections.", true, "", func(c *RuntimeConfig) *bool { return &c.EmailSMTPStartTLS }},
	{"jmap-insecure", EnvJMAPInsecure, "Skip TLS certificate verification for JMAP.", false, "", func(c *RuntimeConfig) *bool { return &c.EmailJMAPInsecure }},
	{"email-enabled", EnvEmailEnabled, "Enable or disable the sending of queued emails.", true, "", func(c *RuntimeConfig) *bool { return &c.EmailEnabled }},
	{"notifications-enabled", EnvNotificationsEnabled, "Enable or disable the internal notification system.", true, "", func(c *RuntimeConfig) *bool { return &c.NotificationsEnabled }},
	{"csrf-enabled", EnvCSRFEnabled, "Enable or disable CSRF protection.", true, "", func(c *RuntimeConfig) *bool { return &c.CSRFEnabled }},
	{"admin-notify", EnvAdminNotify, "Enable or disable email notifications for administrators.", true, "", func(c *RuntimeConfig) *bool { return &c.AdminNotify }},
	{"auto-migrate", EnvAutoMigrate, "Run database migrations on startup.", false, "", func(c *RuntimeConfig) *bool { return &c.AutoMigrate }},
	{"create-dirs", EnvCreateDirs, "Enable or disable the automatic creation of missing directories.", false, "", func(c *RuntimeConfig) *bool { return &c.CreateDirs }},
}

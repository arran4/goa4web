package config

const (
	// EnvDBConn is the environment variable for the database connection string.
	EnvDBConn = "DB_CONN"
	// EnvDBDriver selects the database driver.
	EnvDBDriver = "DB_DRIVER"
	// EnvDBTimezone sets the database timezone.
	EnvDBTimezone = "DB_TIMEZONE"

	// EnvDBHost is the database server hostname.
	EnvDBHost = "DB_HOST"
	// EnvDBPort is the database server port.
	EnvDBPort = "DB_PORT"
	// EnvDBUser is the database username.
	EnvDBUser = "DB_USER"
	// EnvDBPass is the database password.
	EnvDBPass = "DB_PASS"
	// EnvDBName is the database name.
	EnvDBName = "DB_NAME"

	// EnvEmailProvider selects the mail sending backend.
	EnvEmailProvider = "EMAIL_PROVIDER"
	// EnvSMTPHost is the SMTP server hostname.
	EnvSMTPHost = "SMTP_HOST"
	// EnvSMTPPort is the SMTP server port.
	EnvSMTPPort = "SMTP_PORT"
	// EnvSMTPUser is the SMTP username.
	EnvSMTPUser = "SMTP_USER"
	// EnvSMTPPass is the SMTP password.
	EnvSMTPPass = "SMTP_PASS"
	// EnvSMTPAuth selects the SMTP authentication method.
	EnvSMTPAuth = "SMTP_AUTH"
	// EnvSMTPStartTLS enables STARTTLS when sending SMTP mail.
	EnvSMTPStartTLS = "SMTP_STARTTLS"
	// EnvEmailFrom sets the default From address for outgoing mail.
	// The value must be a valid RFC 5322 address.
	EnvEmailFrom = "EMAIL_FROM"
	// EnvEmailSubjectPrefix sets the prefix used for email subjects.
	EnvEmailSubjectPrefix = "EMAIL_SUBJECT_PREFIX"
	// EnvEmailSignOff specifies the sign-off text appended to emails.
	EnvEmailSignOff = "EMAIL_SIGNOFF"
	// EnvAWSRegion is the AWS region for the SES provider.
	EnvAWSRegion = "AWS_REGION"
	// EnvJMAPEndpoint is the JMAP API endpoint.
	EnvJMAPEndpoint = "JMAP_ENDPOINT"
	// EnvJMAPAccount is the JMAP account identifier.
	EnvJMAPAccount = "JMAP_ACCOUNT"
	// EnvJMAPIdentity is the JMAP identity identifier.
	EnvJMAPIdentity = "JMAP_IDENTITY"
	// EnvJMAPUser is the username for the JMAP provider.
	EnvJMAPUser = "JMAP_USER"
	// EnvJMAPPass is the password for the JMAP provider.
	EnvJMAPPass = "JMAP_PASS"
	// EnvJMAPInsecure toggles TLS certificate verification for JMAP.
	EnvJMAPInsecure = "JMAP_INSECURE"

	// EnvConfigFile is the environment variable specifying the path to the
	// main application configuration file.
	EnvConfigFile = "CONFIG_FILE"

	// EnvEmailEnabled toggles sending queued emails.
	EnvEmailEnabled = "EMAIL_ENABLED"
	// EnvNotificationsEnabled toggles the internal notification system.
	EnvNotificationsEnabled = "NOTIFICATIONS_ENABLED"
	// EnvCSRFEnabled toggles CSRF protection.
	EnvCSRFEnabled = "CSRF_ENABLED"

	// EnvFeedsEnabled toggles RSS and Atom feed generation.
	EnvFeedsEnabled = "FEEDS_ENABLED"

	// EnvPageSizeMin defines the minimum allowed page size.
	EnvPageSizeMin = "PAGE_SIZE_MIN"
	// EnvPageSizeMax defines the maximum allowed page size.
	EnvPageSizeMax = "PAGE_SIZE_MAX"
	// EnvPageSizeDefault defines the default page size.
	EnvPageSizeDefault = "PAGE_SIZE_DEFAULT"

	// EnvStatsStartYear sets the first year displayed by the usage stats page.
	EnvStatsStartYear = "STATS_START_YEAR"

	// EnvDBLogVerbosity controls the verbosity level of database logging.
	EnvDBLogVerbosity = "DB_LOG_VERBOSITY"
	// EnvEmailLogVerbosity controls the verbosity level of email logging.
	EnvEmailLogVerbosity = "EMAIL_LOG_VERBOSITY"
	// EnvLogFlags selects which HTTP request logs are emitted.
	EnvLogFlags = "LOG_FLAGS"
	// EnvListen is the network address the HTTP server listens on.
	EnvListen = "LISTEN"
	// EnvHostname is the base URL advertised by the HTTP server.
	EnvHostname = "HOSTNAME"
	// EnvTimezone sets the default site timezone used when users have not
	// configured their own.
	EnvTimezone = "TIMEZONE"
	// EnvTemplatesDir specifies a directory to load templates from at runtime.
	EnvTemplatesDir = "TEMPLATES_DIR"
	// EnvHSTSHeader sets the Strict-Transport-Security header value.
	EnvHSTSHeader = "HSTS_HEADER"
	// EnvSessionName sets the cookie name used for session data.
	EnvSessionName = "SESSION_NAME"

	// EnvSessionSecret is the secret used to encrypt session cookies.
	EnvSessionSecret = "SESSION_SECRET"
	// EnvSessionSecretFile specifies the file containing the session secret.
	EnvSessionSecretFile = "SESSION_SECRET_FILE"
	// EnvSessionSameSite sets the cookie SameSite policy for session data.
	EnvSessionSameSite = "SESSION_SAME_SITE"

	// EnvDocker indicates the application is running inside a Docker container.
	EnvDocker = "GOA4WEB_DOCKER"

	// EnvSendGridKey is the API key for the SendGrid email provider.
	EnvSendGridKey = "SENDGRID_KEY"
	// EnvEmailWorkerInterval controls how often the email worker processes
	// pending messages in seconds.
	EnvEmailWorkerInterval = "EMAIL_WORKER_INTERVAL"
	// EnvEmailVerificationExpiryHours sets the email verification expiry in hours.
	EnvEmailVerificationExpiryHours = "EMAIL_VERIFICATION_EXPIRY_HOURS"
	// EnvPasswordResetExpiryHours sets the password reset expiry in hours.
	EnvPasswordResetExpiryHours = "PASSWORD_RESET_EXPIRY_HOURS"
	// EnvLoginAttemptWindow defines the time window in minutes used to
	// track failed login attempts.
	EnvLoginAttemptWindow = "LOGIN_ATTEMPT_WINDOW"
	// EnvLoginAttemptThreshold sets how many failed logins are allowed
	// within the window before further attempts are denied.
	EnvLoginAttemptThreshold = "LOGIN_ATTEMPT_THRESHOLD"
	// EnvAdminEmails is a comma-separated list of administrator email addresses.
	EnvAdminEmails = "ADMIN_EMAILS"
	// EnvAdminNotify toggles sending administrator notification emails.
	EnvAdminNotify = "ADMIN_NOTIFY"

	// EnvImageUploadDir defines the directory where uploaded images are stored.
	EnvImageUploadDir = "IMAGE_UPLOAD_DIR"
	// EnvImageUploadProvider selects the image upload backend.
	EnvImageUploadProvider = "IMAGE_UPLOAD_PROVIDER"
	// EnvImageUploadS3URL defines the S3 bucket and prefix used by the S3
	// upload provider.
	EnvImageUploadS3URL = "IMAGE_UPLOAD_S3_URL"
	// EnvImageMaxBytes sets the maximum allowed size of uploaded images in bytes.
	EnvImageMaxBytes = "IMAGE_MAX_BYTES"
	// EnvImageCacheDir defines where thumbnails are cached.
	EnvImageCacheDir = "IMAGE_CACHE_DIR"
	// EnvImageCacheProvider selects the cache storage backend.
	EnvImageCacheProvider = "IMAGE_CACHE_PROVIDER"
	// EnvImageCacheS3URL defines the S3 bucket and prefix used by the S3
	// cache provider.
	EnvImageCacheS3URL = "IMAGE_CACHE_S3_URL"
	// EnvImageCacheMaxBytes sets the maximum cache size in bytes.
	EnvImageCacheMaxBytes = "IMAGE_CACHE_MAX_BYTES"
	// EnvImageSignSecret provides the signing key for image URLs.
	EnvImageSignSecret = "IMAGE_SIGN_SECRET"
	// EnvImageSignSecretFile specifies the file containing the signing key.
	EnvImageSignSecretFile = "IMAGE_SIGN_SECRET_FILE"

	// EnvLinkSignSecret provides the signing key for external link URLs.
	EnvLinkSignSecret = "LINK_SIGN_SECRET"
	// EnvLinkSignSecretFile specifies the file containing the link signing key.
	EnvLinkSignSecretFile = "LINK_SIGN_SECRET_FILE"

	// EnvShareSignSecret provides the signing key for share URLs.
	EnvShareSignSecret = "SHARE_SIGN_SECRET"
	// EnvShareSignSecretFile specifies the file containing the share signing key.
	EnvShareSignSecretFile = "SHARE_SIGN_SECRET_FILE"

	// EnvDefaultLanguage specifies the site's default language.
	EnvDefaultLanguage = "DEFAULT_LANGUAGE"

	// EnvDLQProvider selects the dead letter queue backend.
	EnvDLQProvider = "DLQ_PROVIDER"
	// EnvDLQFile is the file path used by the file DLQ provider.
	EnvDLQFile = "DLQ_FILE"

	// EnvAutoMigrate toggles automatic database migrations on startup.
	EnvAutoMigrate = "AUTO_MIGRATE"
	// EnvMigrationsDir specifies a directory to load migrations from at runtime.
	EnvMigrationsDir = "MIGRATIONS_DIR"

	// EnvCreateDirs creates missing directories when enabled.
	EnvCreateDirs = "CREATE_DIRS"

	// EnvAdminAPISecret provides the signing key for admin API tokens.
	EnvAdminAPISecret = "ADMIN_API_SECRET"
	// EnvAdminAPISecretFile specifies the file containing the admin API signing key.
	EnvAdminAPISecretFile = "ADMIN_API_SECRET_FILE"

	// EnvOGImageWidth sets the width of the Open Graph image.
	EnvOGImageWidth = "OG_IMAGE_WIDTH"
	// EnvOGImageHeight sets the height of the Open Graph image.
	EnvOGImageHeight = "OG_IMAGE_HEIGHT"
	// EnvOGImagePattern sets the pattern of the Open Graph image.
	EnvOGImagePattern = "OG_IMAGE_PATTERN"
	// EnvOGImageFgColor sets the foreground color of the Open Graph image.
	EnvOGImageFgColor = "OG_IMAGE_FG_COLOR"
	// EnvOGImageBgColor sets the background color of the Open Graph image.
	EnvOGImageBgColor = "OG_IMAGE_BG_COLOR"
	// EnvOGImageRpgTheme sets the RPG theme of the Open Graph image.
	EnvOGImageRpgTheme = "OG_IMAGE_RPG_THEME"
	// EnvTwitterSite sets the Twitter site handle.
	EnvTwitterSite = "TWITTER_SITE"
)

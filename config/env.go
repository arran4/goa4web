package config

const (
	// EnvDBConn is the environment variable for the database connection string.
	EnvDBConn = "DB_CONN"
	// EnvDBDriver selects the database driver.
	EnvDBDriver = "DB_DRIVER"

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
	EnvEmailFrom = "EMAIL_FROM"
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
	// EnvLogFlags selects which HTTP request logs are emitted.
	EnvLogFlags = "LOG_FLAGS"
	// EnvListen is the network address the HTTP server listens on.
	EnvListen = "LISTEN"
	// EnvHostname is the base URL advertised by the HTTP server.
	EnvHostname = "HOSTNAME"

	// EnvSessionSecret is the secret used to encrypt session cookies.
	EnvSessionSecret = "SESSION_SECRET"
	// EnvSessionSecretFile specifies the file containing the session secret.
	EnvSessionSecretFile = "SESSION_SECRET_FILE"

	// EnvDocker indicates the application is running inside a Docker container.
	EnvDocker = "GOA4WEB_DOCKER"

	// EnvSendGridKey is the API key for the SendGrid email provider.
	EnvSendGridKey = "SENDGRID_KEY"
	// EnvEmailWorkerInterval controls how often the email worker processes
	// pending messages in seconds.
	EnvEmailWorkerInterval = "EMAIL_WORKER_INTERVAL"
	// EnvAdminEmails is a comma-separated list of administrator email addresses.
	EnvAdminEmails = "ADMIN_EMAILS"
	// EnvAdminNotify toggles sending administrator notification emails.
	EnvAdminNotify = "ADMIN_NOTIFY"

	// EnvImageUploadDir defines the directory where uploaded images are stored.
	EnvImageUploadDir = "IMAGE_UPLOAD_DIR"
	// EnvImageMaxBytes sets the maximum allowed size of uploaded images in bytes.
	EnvImageMaxBytes = "IMAGE_MAX_BYTES"

	// EnvDefaultLanguage specifies the site's default language.
	EnvDefaultLanguage = "DEFAULT_LANGUAGE"

	// EnvDLQProvider selects the dead letter queue backend.
	EnvDLQProvider = "DLQ_PROVIDER"
	// EnvDLQFile is the file path used by the file DLQ provider.
	EnvDLQFile = "DLQ_FILE"
)

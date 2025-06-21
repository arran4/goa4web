package config

const (
	// EnvDBUser is the environment variable for the database username.
	EnvDBUser = "DB_USER"
	// EnvDBPass is the environment variable for the database password.
	EnvDBPass = "DB_PASS"
	// EnvDBHost is the environment variable for the database host.
	EnvDBHost = "DB_HOST"
	// EnvDBPort is the environment variable for the database port.
	EnvDBPort = "DB_PORT"
	// EnvDBName is the environment variable for the database name.
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

	// EnvFeedsEnabled toggles RSS and Atom feed generation.
	EnvFeedsEnabled = "FEEDS_ENABLED"

	// EnvPageSizeMin defines the minimum allowed page size.
	EnvPageSizeMin = "PAGE_SIZE_MIN"
	// EnvPageSizeMax defines the maximum allowed page size.
	EnvPageSizeMax = "PAGE_SIZE_MAX"
)

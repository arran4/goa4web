package runtimeconfig

import "github.com/arran4/goa4web/config"

// StringOption describes a string based runtime configuration flag.
type StringOption struct {
	Name    string
	Env     string
	Field   string
	Usage   string
	Default string
}

// IntOption describes an integer runtime configuration flag.
type IntOption struct {
	Name    string
	Env     string
	Field   string
	Usage   string
	Default int
}

// StringOptions lists the string runtime options shared by flag parsing and configuration generation.
var StringOptions = []StringOption{
	{"db-conn", config.EnvDBConn, "DBConn", "database connection string", ""},
	{"db-driver", config.EnvDBDriver, "DBDriver", "database driver", "mysql"},
	{"listen", config.EnvListen, "HTTPListen", "server listen address", ":8080"},
	{"hostname", config.EnvHostname, "HTTPHostname", "server base URL", ""},
	{"email-provider", config.EnvEmailProvider, "EmailProvider", "email provider", ""},
	{"smtp-host", config.EnvSMTPHost, "EmailSMTPHost", "SMTP host", ""},
	{"smtp-port", config.EnvSMTPPort, "EmailSMTPPort", "SMTP port", ""},
	{"smtp-user", config.EnvSMTPUser, "EmailSMTPUser", "SMTP user", ""},
	{"smtp-pass", config.EnvSMTPPass, "EmailSMTPPass", "SMTP pass", ""},
	{"aws-region", config.EnvAWSRegion, "EmailAWSRegion", "AWS region", ""},
	{"jmap-endpoint", config.EnvJMAPEndpoint, "EmailJMAPEndpoint", "JMAP endpoint", ""},
	{"jmap-account", config.EnvJMAPAccount, "EmailJMAPAccount", "JMAP account", ""},
	{"jmap-identity", config.EnvJMAPIdentity, "EmailJMAPIdentity", "JMAP identity", ""},
	{"jmap-user", config.EnvJMAPUser, "EmailJMAPUser", "JMAP user", ""},
	{"jmap-pass", config.EnvJMAPPass, "EmailJMAPPass", "JMAP pass", ""},
	{"sendgrid-key", config.EnvSendGridKey, "EmailSendGridKey", "SendGrid API key", ""},
	{"default-language", config.EnvDefaultLanguage, "DefaultLanguage", "default language name", ""},
	{"image-upload-dir", config.EnvImageUploadDir, "ImageUploadDir", "directory to store uploaded images", ""},
	{"dlq-provider", config.EnvDLQProvider, "DLQProvider", "dead letter queue provider", ""},
	{"dlq-file", config.EnvDLQFile, "DLQFile", "dead letter queue file path", ""},
}

// IntOptions lists the integer runtime options shared by flag parsing and configuration generation.
var IntOptions = []IntOption{
	{"db-log-verbosity", config.EnvDBLogVerbosity, "DBLogVerbosity", "database logging verbosity", 0},
	{"log-flags", config.EnvLogFlags, "LogFlags", "request logging flags", 0},
	{"page-size-min", config.EnvPageSizeMin, "PageSizeMin", "minimum allowed page size", 0},
	{"page-size-max", config.EnvPageSizeMax, "PageSizeMax", "maximum allowed page size", 0},
	{"page-size-default", config.EnvPageSizeDefault, "PageSizeDefault", "default page size", 0},
	{"image-max-bytes", config.EnvImageMaxBytes, "ImageMaxBytes", "maximum allowed upload size in bytes", 0},
	{"email-worker-interval", config.EnvEmailWorkerInterval, "EmailWorkerInterval", "interval in seconds between email worker runs", 0},
}

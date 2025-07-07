package email

import (
	"log"
	"net/smtp"
	"strings"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/runtimeconfig"
)

// ProviderFromConfig returns an email provider configured from cfg.
func ProviderFromConfig(cfg runtimeconfig.RuntimeConfig) Provider {
	mode := strings.ToLower(cfg.EmailProvider)

	switch mode {
	case "smtp":
		host := cfg.EmailSMTPHost
		port := cfg.EmailSMTPPort
		user := cfg.EmailSMTPUser
		pass := cfg.EmailSMTPPass
		if host == "" {
			log.Printf("Email disabled: %s not set", config.EnvSMTPHost)
			return nil
		}
		addr := host
		if port != "" {
			addr = host + ":" + port
		}
		var auth smtp.Auth
		if user != "" {
			auth = smtp.PlainAuth("", user, pass, host)
		}
		return SMTPProvider{Addr: addr, Auth: auth, From: SourceEmail}

	case "ses":
		return SESProviderFromConfig(cfg)

	case "local":
		return LocalProvider{}

	case "jmap":
		ep := cfg.EmailJMAPEndpoint
		if ep == "" {
			log.Printf("Email disabled: %s not set", config.EnvJMAPEndpoint)
			return nil
		}
		acc := cfg.EmailJMAPAccount
		id := cfg.EmailJMAPIdentity
		if acc == "" || id == "" {
			log.Printf("Email disabled: %s or %s not set", config.EnvJMAPAccount, config.EnvJMAPIdentity)
			return nil
		}
		return JMAPProvider{
			Endpoint:  ep,
			Username:  cfg.EmailJMAPUser,
			Password:  cfg.EmailJMAPPass,
			AccountID: acc,
			Identity:  id,
		}

	case "sendgrid":
		return SendGridProviderFromConfig(cfg.EmailSendGridKey)

	case "log":
		return LogProvider{}

	default:
		log.Printf("Email disabled: unknown provider %q", mode)
		return nil
	}
}

package admin

import (
	"context"
	db "github.com/arran4/goa4web/internal/db"
	"log"
	"net/smtp"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/emailutil"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/runtimeconfig"
)

func notifyChange(ctx context.Context, provider email.Provider, emailAddr, page string) error {
	n := notif.Notifier{EmailProvider: provider}
	return n.NotifyChange(ctx, 0, emailAddr, page, "update", nil)
}

func providerFromConfig(cfg runtimeconfig.RuntimeConfig) email.Provider {
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
		return email.SMTPProvider{Addr: addr, Auth: auth, From: email.SourceEmail}
	case "ses":
		awsCfg := aws.NewConfig()
		if region := cfg.EmailAWSRegion; region != "" {
			awsCfg = awsCfg.WithRegion(region)
		}
		sess, err := session.NewSession(awsCfg)
		if err != nil {
			log.Printf("Email disabled: cannot initialise AWS session: %v", err)
			return nil
		}
		if _, err := sess.Config.Credentials.Get(); err != nil {
			log.Printf("Email disabled: no AWS credentials: %v", err)
			return nil
		}
		return email.SESProvider{Client: ses.New(sess)}
	case "local":
		return email.LocalProvider{}
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
		return email.JMAPProvider{
			Endpoint:  ep,
			Username:  cfg.EmailJMAPUser,
			Password:  cfg.EmailJMAPPass,
			AccountID: acc,
			Identity:  id,
		}
	case "sendgrid":
		return email.SendGridProviderFromConfig(cfg.EmailSendGridKey)
	case "log":
		return email.LogProvider{}
	default:
		log.Printf("Email disabled: unknown provider %q", mode)
		return nil
	}
}

func getEmailProvider() email.Provider {
	return providerFromConfig(runtimeconfig.AppRuntimeConfig)
}

func emailSendingEnabled() bool {
	v := strings.ToLower(os.Getenv(config.EnvEmailEnabled))
	if v == "" {
		return true
	}
	switch v {
	case "0", "false", "off", "no":
		return false
	default:
		return true
	}
}

func getUpdateEmailText(ctx context.Context) string {
	if q, ok := ctx.Value(common.KeyQueries).(*db.Queries); ok && q != nil {
		if body, err := q.GetTemplateOverride(ctx, "updateEmail"); err == nil && body != "" {
			return body
		}
	}
	return defaultUpdateEmailText
}

var defaultUpdateEmailText = templates.UpdateEmailText

func getAdminEmails(ctx context.Context, q *db.Queries) []string {
	return emailutil.GetAdminEmails(ctx, q)
}

func adminNotificationsEnabled() bool {
	return emailutil.AdminNotificationsEnabled()
}

func notifyAdmins(ctx context.Context, provider email.Provider, q *db.Queries, page string) {
	notif.Notifier{EmailProvider: provider, Queries: q}.NotifyAdmins(ctx, page)
}

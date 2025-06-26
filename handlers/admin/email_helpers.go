package admin

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/runtimeconfig"
)

func notifyChange(ctx context.Context, provider email.Provider, emailAddr, page string) error {
	if emailAddr == "" {
		return fmt.Errorf("no email specified")
	}
	if !emailSendingEnabled() {
		return nil
	}
	var body bytes.Buffer
	tmpl, err := template.New("email").Parse(getUpdateEmailText(ctx))
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	data := map[string]string{"Page": page}
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("render template: %w", err)
	}
	if provider != nil {
		if err := provider.Send(ctx, emailAddr, "Notification", body.String()); err != nil {
			return fmt.Errorf("send: %w", err)
		}
	}
	return nil
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
	if q, ok := ctx.Value(common.KeyQueries).(*Queries); ok && q != nil {
		if body, err := q.GetTemplateOverride(ctx, "updateEmail"); err == nil && body != "" {
			return body
		}
	}
	return defaultUpdateEmailText
}

var defaultUpdateEmailText = templates.UpdateEmailText

func getAdminEmails(ctx context.Context, q *Queries) []string {
	env := os.Getenv(config.EnvAdminEmails)
	var emails []string
	if env != "" {
		for _, e := range strings.Split(env, ",") {
			if addr := strings.TrimSpace(e); addr != "" {
				emails = append(emails, addr)
			}
		}
		return emails
	}
	if q != nil {
		rows, err := q.ListAdministratorEmails(ctx)
		if err != nil {
			log.Printf("list admin emails: %v", err)
			return emails
		}
		for _, email := range rows {
			if email.Valid {
				emails = append(emails, email.String)
			}
		}
	}
	return emails
}

func adminNotificationsEnabled() bool {
	v := strings.ToLower(os.Getenv(config.EnvAdminNotify))
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

func notifyAdmins(ctx context.Context, provider email.Provider, q *Queries, page string) {
	if provider == nil || !adminNotificationsEnabled() {
		return
	}
	for _, email := range getAdminEmails(ctx, q) {
		if err := notifyChange(ctx, provider, email, page); err != nil {
			log.Printf("Error: notifyChange: %s", err)
		}
	}
}

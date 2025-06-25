package blogs

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
	"text/template"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/runtimeconfig"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

func notifyChange(ctx context.Context, provider email.Provider, emailAddr string, page string) error {
	if emailAddr == "" {
		return fmt.Errorf("no email specified")
	}
	if !emailSendingEnabled() {
		return nil
	}
	from := email.SourceEmail
	type EmailContent struct {
		To      string
		From    string
		Subject string
		URL     string
	}
	content := EmailContent{
		To:      emailAddr,
		From:    from,
		Subject: "Website Update Notification",
		URL:     page,
	}
	var notification bytes.Buffer
	tmpl, err := template.New("email").Parse(getUpdateEmailText(ctx))
	if err != nil {
		return fmt.Errorf("parse email template: %w", err)
	}
	if err = tmpl.Execute(&notification, content); err != nil {
		return fmt.Errorf("execute email template: %w", err)
	}
	if q, ok := ctx.Value(common.KeyQueries).(*Queries); ok {
		if err := q.InsertPendingEmail(ctx, InsertPendingEmailParams{ToEmail: emailAddr, Subject: content.Subject, Body: notification.String()}); err != nil {
			return err
		}
	} else if provider != nil {
		if err := provider.Send(ctx, emailAddr, content.Subject, notification.String()); err != nil {
			return fmt.Errorf("send email: %w", err)
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
			if mode == "ses" {
				return nil
			}
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

// defaultUpdateEmailText is the compiled-in notification template.
//
//go:embed updateEmail.txt
var defaultUpdateEmailText string

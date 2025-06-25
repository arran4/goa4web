package goa4web

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
	"text/template"

	"github.com/arran4/goa4web/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"

	"github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/runtimeconfig"
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

	// Define email content
	content := EmailContent{
		To:      emailAddr,
		From:    from,
		Subject: "Website Update Notification",
		URL:     page,
	}

	// Create a new buffer to store the rendered email content
	var notification bytes.Buffer

	// Parse and execute the email template
	tmpl, err := template.New("email").Parse(getUpdateEmailText(ctx))
	if err != nil {
		return fmt.Errorf("parse email template: %w", err)
	}

	// Execute the template and store the result in the notification buffer
	err = tmpl.Execute(&notification, content)
	if err != nil {
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
		// Attempt to create an AWS session using default credentials. If this
		// fails, emails are effectively disabled.
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
			// if EMAIL_PROVIDER not specified default to ses but disabled
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

// getEmailProvider returns the mail provider configured by environment variables.
// Production code uses this, while tests can call providerFromConfig directly.
func getEmailProvider() email.Provider {
	return providerFromConfig(runtimeconfig.AppRuntimeConfig)
}

// loadEmailConfigFile reads EMAIL_* style configuration values from a simple
// key=value file. Missing files return an empty configuration.

// getAdminEmails returns a slice of administrator email addresses. If the
// ADMIN_EMAILS environment variable is set, it takes precedence and is
// interpreted as a comma-separated list. If not set and a Queries value is
// provided, the database is queried for administrator accounts.
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

// notifyAdmins sends a change notification email to all administrator
// addresses returned by getAdminEmails.
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

// notifyThreadSubscribers emails users subscribed to the forum thread.
func notifyThreadSubscribers(ctx context.Context, provider email.Provider, q *Queries, threadID, excludeUser int32, page string) {
	if provider == nil {
		return
	}
	rows, err := q.ListUsersSubscribedToThread(ctx, ListUsersSubscribedToThreadParams{
		ForumthreadIdforumthread: threadID,
		Idusers:                  excludeUser,
	})
	if err != nil {
		log.Printf("Error: listUsersSubscribedToThread: %s", err)
		return
	}
	for _, row := range rows {
		if err := notifyChange(ctx, provider, row.Username.String, page); err != nil {
			log.Printf("Error: notifyChange: %s", err)
		}
	}
}

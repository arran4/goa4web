package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
	"text/template"
)

var (
	//go:embed "templates/updateEmail.txt"
	updateEmailText string
)

const (
	SourceEmail = "a4web@arran.net.au"
)

// cliEmailConfig holds command line overrides for email configuration. It is
// populated in main() after parsing flags.
var cliEmailConfig EmailConfig

// emailConfigFile is the optional path to a configuration file read at startup.
var emailConfigFile string

// EmailConfig stores configuration for selecting and configuring the mail
// provider. Tests can supply a custom configuration instead of relying on
// environment variables.
type EmailConfig struct {
	Provider     string
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPass     string
	AWSRegion    string
	JMAPEndpoint string
	JMAPAccount  string
	JMAPIdentity string
	JMAPUser     string
	JMAPPass     string
	SendGridKey  string
}

// MailProvider defines a simple interface that all mail backends must
// implement. Only the fields necessary for sending basic notification emails
// are included.
type MailProvider interface {
	Send(ctx context.Context, to, subject, body string) error
}

// sesMailProvider wraps the AWS SES client.
type sesMailProvider struct{ client sesiface.SESAPI }

func (s sesMailProvider) Send(ctx context.Context, to, subject, body string) error {
	dest := &ses.Destination{ToAddresses: []*string{aws.String(to)}}
	msg := &ses.Message{
		Subject: &ses.Content{Charset: aws.String("UTF-8"), Data: aws.String(subject)},
		Body:    &ses.Body{Text: &ses.Content{Charset: aws.String("UTF-8"), Data: aws.String(body)}},
	}
	input := &ses.SendEmailInput{Destination: dest, Message: msg, Source: aws.String(SourceEmail)}
	_, err := s.client.SendEmailWithContext(ctx, input)
	return err
}

// smtpMailProvider uses the standard net/smtp package.
type smtpMailProvider struct {
	addr string
	auth smtp.Auth
	from string
}

func (s smtpMailProvider) Send(ctx context.Context, to, subject, body string) error {
	msg := []byte("To: " + to + "\r\nSubject: " + subject + "\r\n\r\n" + body)
	return smtp.SendMail(s.addr, s.auth, s.from, []string{to}, msg)
}

// localMailProvider relies on the local sendmail binary.
type localMailProvider struct{}

func (localMailProvider) Send(ctx context.Context, to, subject, body string) error {
	cmd := exec.CommandContext(ctx, "sendmail", to)
	cmd.Stdin = strings.NewReader("Subject: " + subject + "\n\n" + body)
	return cmd.Run()
}

// jmapMailProvider is a placeholder for sending mail via JMAP.
type jmapMailProvider struct {
	endpoint  string
	username  string
	password  string
	accountID string
	identity  string
}

// Send delivers a message using the JMAP EmailSubmission API. The provider
// uploads the RFC822 message to the JMAP server and then creates an
// EmailSubmission referencing the uploaded blob.
func (j jmapMailProvider) Send(ctx context.Context, to, subject, body string) error {
	// Build a simple RFC822 message.
	var msg bytes.Buffer
	fmt.Fprintf(&msg, "From: %s\r\n", SourceEmail)
	fmt.Fprintf(&msg, "To: %s\r\n", to)
	fmt.Fprintf(&msg, "Subject: %s\r\n", subject)
	msg.WriteString("\r\n")
	msg.WriteString(body)

	uploadURL := fmt.Sprintf("%s/upload/%s", strings.TrimRight(j.endpoint, "/"), j.accountID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, bytes.NewReader(msg.Bytes()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(j.username, j.password)
	req.Header.Set("Content-Type", "message/rfc822")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("upload failed: %s", resp.Status)
	}
	var up struct {
		BlobID string `json:"blobId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&up); err != nil {
		return err
	}

	// Create the Email and submit it for sending.
	payload := map[string]interface{}{
		"using": []string{"urn:ietf:params:jmap:core", "urn:ietf:params:jmap:mail"},
		"methodCalls": [][]interface{}{
			{
				"Email/import",
				map[string]interface{}{
					"accountId": j.accountID,
					"emails": map[string]interface{}{
						"msg": map[string]interface{}{
							"blobId":     up.BlobID,
							"mailboxIds": map[string]bool{"outbox": true},
						},
					},
				},
				"c1",
			},
			{
				"EmailSubmission/set",
				map[string]interface{}{
					"accountId": j.accountID,
					"create": map[string]interface{}{
						"sub": map[string]interface{}{
							"emailId":    "#msg",
							"identityId": j.identity,
						},
					},
				},
				"c2",
			},
		},
	}

	buf, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err = http.NewRequestWithContext(ctx, http.MethodPost, j.endpoint, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	req.SetBasicAuth(j.username, j.password)
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("jmap send failed: %s", resp.Status)
	}
	return nil
}

// logMailProvider just logs emails for development purposes.
type logMailProvider struct{}

func (logMailProvider) Send(ctx context.Context, to, subject, body string) error {
	log.Printf("sending mail to %s subject %q\n%s", to, subject, body)
	return nil
}

func notifyChange(ctx context.Context, provider MailProvider, email string, page string) error {
	// TODO Make this periodically check db to see if there is anything queued, and normally queue up messages
	if email == "" {
		return fmt.Errorf("no email specified")
	}

	if provider == nil {
		return fmt.Errorf("no email provider specified")
	}
	from := SourceEmail

	type EmailContent struct {
		To      string
		From    string
		Subject string
		URL     string
	}

	// Define email content
	content := EmailContent{
		To:      email,
		From:    from,
		Subject: "Website Update Notification",
		URL:     page,
	}

	// Create a new buffer to store the rendered email content
	var notification bytes.Buffer

	// Parse and execute the email template
	tmpl, err := template.New("email").Parse(updateEmailText)
	if err != nil {
		return fmt.Errorf("parse email template: %w", err)
	}

	// Execute the template and store the result in the notification buffer
	err = tmpl.Execute(&notification, content)
	if err != nil {
		return fmt.Errorf("execute email template: %w", err)
	}

	// Send the message using the provider
	if err := provider.Send(ctx, email, content.Subject, notification.String()); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	log.Println("Email sent successfully to", email)
	return nil
}

func providerFromConfig(cfg EmailConfig) MailProvider {
	mode := strings.ToLower(cfg.Provider)

	switch mode {
	case "smtp":
		host := cfg.SMTPHost
		port := cfg.SMTPPort
		user := cfg.SMTPUser
		pass := cfg.SMTPPass
		if host == "" {
			log.Printf("Email disabled: SMTP_HOST not set")
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
		return smtpMailProvider{addr: addr, auth: auth, from: SourceEmail}

	case "ses", "":
		// Attempt to create an AWS session using default credentials. If this
		// fails, emails are effectively disabled.
		awsCfg := aws.NewConfig()
		if region := cfg.AWSRegion; region != "" {
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

		return sesMailProvider{client: ses.New(sess)}

	case "local":
		return localMailProvider{}

	case "jmap":
		ep := cfg.JMAPEndpoint
		if ep == "" {
			log.Printf("Email disabled: JMAP_ENDPOINT not set")
			return nil
		}
		acc := cfg.JMAPAccount
		id := cfg.JMAPIdentity
		if acc == "" || id == "" {
			log.Printf("Email disabled: JMAP_ACCOUNT or JMAP_IDENTITY not set")
			return nil
		}
		return jmapMailProvider{
			endpoint:  ep,
			username:  cfg.JMAPUser,
			password:  cfg.JMAPPass,
			accountID: acc,
			identity:  id,
		}

	case "sendgrid":
		return sendGridProviderFromConfig(cfg)

	case "log":
		return logMailProvider{}

	default:
		log.Printf("Email disabled: unknown provider %q", mode)
		return nil
	}
}

// getEmailProvider returns the mail provider configured by environment variables.
// Production code uses this, while tests can call providerFromConfig directly.
func getEmailProvider() MailProvider {
	return providerFromConfig(loadEmailConfig())
}

// resolveEmailConfig merges configuration values with the order of precedence
// cli > file > env > defaults.
func resolveEmailConfig(cli, file, env EmailConfig) EmailConfig {
	var cfg EmailConfig
	merge := func(src EmailConfig) {
		if src.Provider != "" {
			cfg.Provider = src.Provider
		}
		if src.SMTPHost != "" {
			cfg.SMTPHost = src.SMTPHost
		}
		if src.SMTPPort != "" {
			cfg.SMTPPort = src.SMTPPort
		}
		if src.SMTPUser != "" {
			cfg.SMTPUser = src.SMTPUser
		}
		if src.SMTPPass != "" {
			cfg.SMTPPass = src.SMTPPass
		}
		if src.AWSRegion != "" {
			cfg.AWSRegion = src.AWSRegion
		}
		if src.JMAPEndpoint != "" {
			cfg.JMAPEndpoint = src.JMAPEndpoint
		}
		if src.JMAPAccount != "" {
			cfg.JMAPAccount = src.JMAPAccount
		}
		if src.JMAPIdentity != "" {
			cfg.JMAPIdentity = src.JMAPIdentity
		}
		if src.JMAPUser != "" {
			cfg.JMAPUser = src.JMAPUser
		}
		if src.JMAPPass != "" {
			cfg.JMAPPass = src.JMAPPass
		}
		if src.SendGridKey != "" {
			cfg.SendGridKey = src.SendGridKey
		}
	}

	merge(env)
	merge(file)
	merge(cli)
	return cfg
}

// loadEmailConfigFile reads EMAIL_* style configuration values from a simple
// key=value file. Missing files return an empty configuration.
func loadEmailConfigFile(path string) (EmailConfig, error) {
	var cfg EmailConfig
	if path == "" {
		return cfg, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	for _, line := range strings.Split(string(b), "\n") {
		if i := strings.IndexByte(line, '='); i > 0 {
			key := strings.TrimSpace(line[:i])
			val := strings.TrimSpace(line[i+1:])
			switch key {
			case "EMAIL_PROVIDER":
				cfg.Provider = val
			case "SMTP_HOST":
				cfg.SMTPHost = val
			case "SMTP_PORT":
				cfg.SMTPPort = val
			case "SMTP_USER":
				cfg.SMTPUser = val
			case "SMTP_PASS":
				cfg.SMTPPass = val
			case "AWS_REGION":
				cfg.AWSRegion = val
			case "JMAP_ENDPOINT":
				cfg.JMAPEndpoint = val
			case "JMAP_ACCOUNT":
				cfg.JMAPAccount = val
			case "JMAP_IDENTITY":
				cfg.JMAPIdentity = val
			case "JMAP_USER":
				cfg.JMAPUser = val
			case "JMAP_PASS":
				cfg.JMAPPass = val
			case "SENDGRID_KEY":
				cfg.SendGridKey = val
			}
		}
	}
	return cfg, nil
}

func loadEmailConfig() EmailConfig {
	env := EmailConfig{
		Provider:     os.Getenv("EMAIL_PROVIDER"),
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     os.Getenv("SMTP_PORT"),
		SMTPUser:     os.Getenv("SMTP_USER"),
		SMTPPass:     os.Getenv("SMTP_PASS"),
		AWSRegion:    os.Getenv("AWS_REGION"),
		JMAPEndpoint: os.Getenv("JMAP_ENDPOINT"),
		JMAPAccount:  os.Getenv("JMAP_ACCOUNT"),
		JMAPIdentity: os.Getenv("JMAP_IDENTITY"),
		JMAPUser:     os.Getenv("JMAP_USER"),
		JMAPPass:     os.Getenv("JMAP_PASS"),
		SendGridKey:  os.Getenv("SENDGRID_KEY"),
	}

	fileCfg, err := loadEmailConfigFile(emailConfigFile)
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Email config file error: %v", err)
	}

	return resolveEmailConfig(cliEmailConfig, fileCfg, env)
}

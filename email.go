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

func getEmailProvider() MailProvider {
	mode := strings.ToLower(os.Getenv("EMAIL_PROVIDER"))

	switch mode {
	case "smtp":
		host := os.Getenv("SMTP_HOST")
		port := os.Getenv("SMTP_PORT")
		user := os.Getenv("SMTP_USER")
		pass := os.Getenv("SMTP_PASS")
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
		cfg := aws.NewConfig()
		if region := os.Getenv("AWS_REGION"); region != "" {
			cfg = cfg.WithRegion(region)
		}

		sess, err := session.NewSession(cfg)
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
		ep := os.Getenv("JMAP_ENDPOINT")
		if ep == "" {
			log.Printf("Email disabled: JMAP_ENDPOINT not set")
			return nil
		}
		acc := os.Getenv("JMAP_ACCOUNT")
		id := os.Getenv("JMAP_IDENTITY")
		if acc == "" || id == "" {
			log.Printf("Email disabled: JMAP_ACCOUNT or JMAP_IDENTITY not set")
			return nil
		}
		return jmapMailProvider{
			endpoint:  ep,
			username:  os.Getenv("JMAP_USER"),
			password:  os.Getenv("JMAP_PASS"),
			accountID: acc,
			identity:  id,
		}

	case "log":
		return logMailProvider{}

	default:
		log.Printf("Email disabled: unknown provider %q", mode)
		return nil
	}
}

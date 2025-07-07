package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os/exec"
	"strings"
	"time"
)

const (
	// SourceEmail is the From address for notification emails.
	SourceEmail = "a4web@arran.net.au"
)

// Provider defines a simple interface that all mail backends must implement.
// Only the fields necessary for sending basic notification emails are included.
type Provider interface {
	Send(ctx context.Context, to, subject, textBody, htmlBody string) error
}

// SMTPProvider uses the standard net/smtp package.
type SMTPProvider struct {
	Addr string
	Auth smtp.Auth
	From string
}

func (s SMTPProvider) Send(ctx context.Context, to, subject, textBody, htmlBody string) error {
	var msg []byte
	if htmlBody != "" {
		boundary := "a4web" + strings.ReplaceAll(fmt.Sprint(time.Now().UnixNano()), "-", "")
		buf := bytes.NewBuffer(nil)
		fmt.Fprintf(buf, "To: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/alternative; boundary=%s\r\n\r\n", to, subject, boundary)
		fmt.Fprintf(buf, "--%s\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s\r\n", boundary, textBody)
		fmt.Fprintf(buf, "--%s\r\nContent-Type: text/html; charset=utf-8\r\n\r\n%s\r\n--%s--", boundary, htmlBody, boundary)
		msg = buf.Bytes()
	} else {
		msg = []byte("To: " + to + "\r\nSubject: " + subject + "\r\n\r\n" + textBody)
	}
	return smtp.SendMail(s.Addr, s.Auth, s.From, []string{to}, msg)
}

// LocalProvider relies on the local sendmail binary.
type LocalProvider struct{}

func (LocalProvider) Send(ctx context.Context, to, subject, textBody, htmlBody string) error {
	cmd := exec.CommandContext(ctx, "sendmail", to)
	body := textBody
	if htmlBody != "" {
		boundary := "a4web" + strings.ReplaceAll(fmt.Sprint(time.Now().UnixNano()), "-", "")
		buf := bytes.NewBuffer(nil)
		fmt.Fprintf(buf, "Subject: %s\nMIME-Version: 1.0\nContent-Type: multipart/alternative; boundary=%s\n\n", subject, boundary)
		fmt.Fprintf(buf, "--%s\nContent-Type: text/plain; charset=utf-8\n\n%s\n", boundary, textBody)
		fmt.Fprintf(buf, "--%s\nContent-Type: text/html; charset=utf-8\n\n%s\n--%s--", boundary, htmlBody, boundary)
		body = buf.String()
		cmd.Stdin = strings.NewReader(body)
	} else {
		cmd.Stdin = strings.NewReader("Subject: " + subject + "\n\n" + body)
	}
	return cmd.Run()
}

// JMAPProvider is a placeholder for sending mail via JMAP.
type JMAPProvider struct {
	Endpoint  string
	Username  string
	Password  string
	AccountID string
	Identity  string
}

// Send delivers a message using the JMAP EmailSubmission API. The provider uploads
// the RFC822 message to the JMAP server and then creates an EmailSubmission referencing
// the uploaded blob.
func (j JMAPProvider) Send(ctx context.Context, to, subject, textBody, htmlBody string) error {
	var msg bytes.Buffer
	boundary := "a4web" + strings.ReplaceAll(fmt.Sprint(time.Now().UnixNano()), "-", "")
	fmt.Fprintf(&msg, "From: %s\r\n", SourceEmail)
	fmt.Fprintf(&msg, "To: %s\r\n", to)
	if htmlBody != "" {
		fmt.Fprintf(&msg, "Subject: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/alternative; boundary=%s\r\n\r\n", subject, boundary)
		fmt.Fprintf(&msg, "--%s\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s\r\n", boundary, textBody)
		fmt.Fprintf(&msg, "--%s\r\nContent-Type: text/html; charset=utf-8\r\n\r\n%s\r\n--%s--", boundary, htmlBody, boundary)
	} else {
		fmt.Fprintf(&msg, "Subject: %s\r\n\r\n", subject)
		msg.WriteString(textBody)
	}

	uploadURL := fmt.Sprintf("%s/upload/%s", strings.TrimRight(j.Endpoint, "/"), j.AccountID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, bytes.NewReader(msg.Bytes()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(j.Username, j.Password)
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

	payload := map[string]interface{}{
		"using": []string{"urn:ietf:params:jmap:core", "urn:ietf:params:jmap:mail"},
		"methodCalls": [][]interface{}{
			{
				"Email/import",
				map[string]interface{}{
					"accountId": j.AccountID,
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
					"accountId": j.AccountID,
					"create": map[string]interface{}{
						"sub": map[string]interface{}{
							"emailId":    "#msg",
							"identityId": j.Identity,
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

	req, err = http.NewRequestWithContext(ctx, http.MethodPost, j.Endpoint, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	req.SetBasicAuth(j.Username, j.Password)
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

// LogProvider just logs emails for development purposes.
type LogProvider struct{}

func (LogProvider) Send(ctx context.Context, to, subject, textBody, htmlBody string) error {
	if htmlBody != "" {
		log.Printf("sending mail to %s subject %q\nTEXT:\n%s\nHTML:\n%s", to, subject, textBody, htmlBody)
	} else {
		log.Printf("sending mail to %s subject %q\n%s", to, subject, textBody)
	}
	return nil
}

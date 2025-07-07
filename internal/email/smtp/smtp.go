package smtp

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/runtimeconfig"
)

// Provider uses the standard net/smtp package.
type Provider struct {
	Addr     string
	Auth     smtp.Auth
	From     string
	StartTLS bool
}

type loginAuth struct {
	username, password string
}

func (a loginAuth) Start(*smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", nil, nil
}

func (a loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		s := strings.TrimSpace(string(fromServer))
		if strings.Contains(strings.ToLower(s), "username") {
			return []byte(a.username), nil
		}
		if strings.Contains(strings.ToLower(s), "password") {
			return []byte(a.password), nil
		}
	}
	return nil, nil
}

func (s Provider) Send(ctx context.Context, to, subject, textBody, htmlBody string) error {
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
	if !s.StartTLS {
		return smtp.SendMail(s.Addr, s.Auth, s.From, []string{to}, msg)
	}
	c, err := smtp.Dial(s.Addr)
	if err != nil {
		return err
	}
	defer c.Close()
	if err = c.Hello("localhost"); err != nil {
		return err
	}
	if ok, _ := c.Extension("STARTTLS"); !ok {
		return fmt.Errorf("smtp: server does not support STARTTLS")
	}
	if err = c.StartTLS(&tls.Config{ServerName: strings.Split(s.Addr, ":")[0]}); err != nil {
		return err
	}
	if s.Auth != nil {
		if ok, _ := c.Extension("AUTH"); !ok {
			return fmt.Errorf("smtp: server doesn't support AUTH")
		}
		if err = c.Auth(s.Auth); err != nil {
			return err
		}
	}
	if err = c.Mail(s.From); err != nil {
		return err
	}
	if err = c.Rcpt(to); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err = w.Write(msg); err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}
	return c.Quit()
}

func providerFromConfig(cfg runtimeconfig.RuntimeConfig) email.Provider {
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
		switch strings.ToLower(cfg.EmailSMTPAuth) {
		case "", "plain":
			auth = smtp.PlainAuth("", user, pass, host)
		case "cram-md5":
			auth = smtp.CRAMMD5Auth(user, pass)
		case "login":
			auth = loginAuth{user, pass}
		default:
			log.Printf("unknown SMTP auth %q", cfg.EmailSMTPAuth)
			auth = smtp.PlainAuth("", user, pass, host)
		}
	}
	return Provider{Addr: addr, Auth: auth, From: cfg.EmailFrom, StartTLS: cfg.EmailSMTPStartTLS}
}

// Register registers the SMTP provider factory.
func Register() { email.RegisterProvider("smtp", providerFromConfig) }

package smtp

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/mail"
	"net/smtp"
	"strings"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/email"
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

func (s Provider) Send(ctx context.Context, to mail.Address, rawEmailMessage []byte) error {
	host, port, err := net.SplitHostPort(s.Addr)
	if err != nil {
		return fmt.Errorf("invalid addr %q: %w", s.Addr, err)
	}

	var c *smtp.Client

	switch port {
	case "25", "587":
		// Plain TCP then upgrade to TLS
		conn, err := net.Dial("tcp", s.Addr)
		if err != nil {
			return fmt.Errorf("smtp dial (plain): %w", err)
		}
		c, err = smtp.NewClient(conn, host)
		if err != nil {
			return fmt.Errorf("smtp client: %w", err)
		}

		// STARTTLS
		if ok, _ := c.Extension("STARTTLS"); !ok {
			return fmt.Errorf("smtp: server does not support STARTTLS")
		}
		if err = c.StartTLS(&tls.Config{ServerName: host}); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}

	case "465":
		// Immediate TLS connection
		tlsConn, err := tls.Dial("tcp", s.Addr, &tls.Config{ServerName: host})
		if err != nil {
			return fmt.Errorf("smtp dial (tls): %w", err)
		}
		c, err = smtp.NewClient(tlsConn, host)
		if err != nil {
			return fmt.Errorf("smtp client (tls): %w", err)
		}

	default:
		return fmt.Errorf("unsupported SMTP port: %s", port)
	}

	defer func() {
		if err := c.Quit(); err != nil {
			log.Printf("smtp.Quit: %v", err)
		}
	}()

	// Optional (can skip if HELO already done inside .NewClient)
	if err = c.Hello("localhost"); err != nil {
		return fmt.Errorf("smtp hello: %w", err)
	}

	if s.Auth != nil {
		if ok, _ := c.Extension("AUTH"); !ok {
			return fmt.Errorf("smtp: server doesn't support AUTH")
		}
		if err = c.Auth(s.Auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}
	if err = c.Mail(s.From); err != nil {
		return fmt.Errorf("smtp from: %w", err)
	}
	if err = c.Rcpt(to.Address); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if _, err = w.Write(rawEmailMessage); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	return c.Quit()
}

func (s Provider) TestConfig(ctx context.Context) (string, error) {
	host, port, err := net.SplitHostPort(s.Addr)
	if err != nil {
		return "", fmt.Errorf("invalid addr %q: %w", s.Addr, err)
	}

	var c *smtp.Client

	switch port {
	case "25", "587":
		conn, err := net.Dial("tcp", s.Addr)
		if err != nil {
			return "", fmt.Errorf("smtp dial (plain): %w", err)
		}
		c, err = smtp.NewClient(conn, host)
		if err != nil {
			return "", fmt.Errorf("smtp client: %w", err)
		}
		if s.StartTLS {
			if ok, _ := c.Extension("STARTTLS"); !ok {
				return "", fmt.Errorf("smtp: server does not support STARTTLS")
			}
			if err = c.StartTLS(&tls.Config{ServerName: host}); err != nil {
				return "", fmt.Errorf("starttls: %w", err)
			}
		}

	case "465":
		tlsConn, err := tls.Dial("tcp", s.Addr, &tls.Config{ServerName: host})
		if err != nil {
			return "", fmt.Errorf("smtp dial (tls): %w", err)
		}
		c, err = smtp.NewClient(tlsConn, host)
		if err != nil {
			return "", fmt.Errorf("smtp client (tls): %w", err)
		}

	default:
		return "", fmt.Errorf("unsupported SMTP port: %s", port)
	}

	defer func() {
		if err := c.Quit(); err != nil {
			log.Printf("smtp.Quit: %v", err)
		}
	}()

	if err = c.Hello("localhost"); err != nil {
		return "", fmt.Errorf("smtp hello: %w", err)
	}

	if s.Auth != nil {
		if ok, _ := c.Extension("AUTH"); !ok {
			return "", fmt.Errorf("smtp: server doesn't support AUTH")
		}
		if err = c.Auth(s.Auth); err != nil {
			return "", fmt.Errorf("smtp auth: %w", err)
		}
	}

	return "SMTP provider is configured correctly", c.Quit()
}

func providerFromConfig(cfg *config.RuntimeConfig) (email.Provider, error) {
	host := cfg.EmailSMTPHost
	port := cfg.EmailSMTPPort
	user := cfg.EmailSMTPUser
	pass := cfg.EmailSMTPPass
	if host == "" {
		return nil, fmt.Errorf("Email disabled: %s not set", config.EnvSMTPHost)
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
	return Provider{Addr: addr, Auth: auth, From: cfg.EmailFrom, StartTLS: cfg.EmailSMTPStartTLS}, nil
}

// Register registers the SMTP provider factory.
func Register(r *email.Registry) { r.RegisterProvider("smtp", providerFromConfig) }

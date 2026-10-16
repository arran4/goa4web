package email

import (
	"bytes"
	"fmt"
	"mime"
	"mime/multipart"
	"net/mail"
	"net/textproto"
	"strings"
	"time"

	"github.com/google/uuid"
)

// DefaultFromName is used when encoding From headers without a name component.
var DefaultFromName string

// SetDefaultFromName extracts the name from addr and stores it in DefaultFromName.
func SetDefaultFromName(addr string) {
	if a, err := mail.ParseAddress(addr); err == nil {
		DefaultFromName = a.Name
	}
}

// MessageOptions configures optional RFC-style headers for an email message.
type MessageOptions struct {
	AutoSubmitted bool
	ReplyTo       *mail.Address
}

// MessageOption configures optional message metadata.
type MessageOption func(*MessageOptions)

// WithAutoSubmitted marks a message as generated automatically by the application.
func WithAutoSubmitted() MessageOption {
	return func(options *MessageOptions) {
		options.AutoSubmitted = true
	}
}

// WithReplyTo sets the address to which recipients should send replies.
func WithReplyTo(address mail.Address) MessageOption {
	return func(options *MessageOptions) {
		options.ReplyTo = &address
	}
}

// BuildMessage constructs a MIME email message with optional HTML content.
func BuildMessage(from, to mail.Address, subject, textBody, htmlBody string, opts ...MessageOption) ([]byte, error) {
	options := MessageOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	if from.Name == "" {
		from.Name = DefaultFromName
	}
	var msg bytes.Buffer
	hdr := textproto.MIMEHeader{}
	hdr.Set("From", from.String())
	hdr.Set("To", to.String())
	hdr.Set("Subject", mime.QEncoding.Encode("utf-8", subject))
	hdr.Set("MIME-Version", "1.0")
	hdr.Set("Date", time.Now().Format(time.RFC1123Z))
	if options.AutoSubmitted {
		hdr.Set("Auto-Submitted", "auto-generated")
	}
	if options.ReplyTo != nil {
		replyTo, err := mail.ParseAddress(options.ReplyTo.String())
		if err != nil {
			return nil, fmt.Errorf("parse Reply-To address: %w", err)
		}
		hdr.Set("Reply-To", replyTo.String())
	}

	domain := "localhost"
	if parts := strings.Split(from.Address, "@"); len(parts) > 1 {
		domain = parts[len(parts)-1]
	}
	hdr.Set("Message-ID", fmt.Sprintf("<%s@%s>", uuid.New().String(), domain))

	if htmlBody != "" {
		w := multipart.NewWriter(&msg)
		boundary := "a4web" + strings.ReplaceAll(fmt.Sprint(time.Now().UnixNano()), "-", "")
		if err := w.SetBoundary(boundary); err != nil {
			return nil, err
		}
		hdr.Set("Content-Type", fmt.Sprintf("multipart/alternative; boundary=%s", boundary))
		for k, v := range hdr {
			for _, vv := range v {
				fmt.Fprintf(&msg, "%s: %s\r\n", k, vv)
			}
		}
		msg.WriteString("\r\n")
		part, err := w.CreatePart(textproto.MIMEHeader{"Content-Type": {"text/plain; charset=utf-8"}})
		if err != nil {
			return nil, err
		}
		_, _ = part.Write([]byte(textBody))
		part, err = w.CreatePart(textproto.MIMEHeader{"Content-Type": {"text/html; charset=utf-8"}})
		if err != nil {
			return nil, err
		}
		_, _ = part.Write([]byte(htmlBody))
		if err := w.Close(); err != nil {
			return nil, err
		}
	} else {
		hdr.Set("Content-Type", "text/plain; charset=utf-8")
		for k, v := range hdr {
			for _, vv := range v {
				fmt.Fprintf(&msg, "%s: %s\r\n", k, vv)
			}
		}
		msg.WriteString("\r\n")
		msg.WriteString(textBody)
	}
	return msg.Bytes(), nil
}

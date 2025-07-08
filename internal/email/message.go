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
)

// DefaultFromName is used when encoding From headers without a name component.
var DefaultFromName string

// SetDefaultFromName extracts the name from addr and stores it in DefaultFromName.
func SetDefaultFromName(addr string) {
	if a, err := mail.ParseAddress(addr); err == nil {
		DefaultFromName = a.Name
	}
}

func formatAddress(a mail.Address) string {
	if a.Name != "" {
		return mime.QEncoding.Encode("utf-8", a.Name) + " <" + a.Address + ">"
	}
	return a.Address
}

// BuildMessage constructs a MIME email message with optional HTML content.
func BuildMessage(from, to mail.Address, subject, textBody, htmlBody string) ([]byte, error) {
	if from.Name == "" {
		from.Name = DefaultFromName
	}
	var msg bytes.Buffer
	hdr := textproto.MIMEHeader{}
	hdr.Set("From", formatAddress(from.String()))
	hdr.Set("To", formatAddress(to.String()))
	hdr.Set("Subject", mime.QEncoding.Encode("utf-8", subject))
	hdr.Set("MIME-Version", "1.0")

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
		part.Write([]byte(textBody))
		part, err = w.CreatePart(textproto.MIMEHeader{"Content-Type": {"text/html; charset=utf-8"}})
		if err != nil {
			return nil, err
		}
		part.Write([]byte(htmlBody))
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

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

// DefaultFromName is used when building messages without a specified sender
// name. The value may be updated at runtime from configuration.
var DefaultFromName = "Goa4Web"

// SetDefaultFromName updates DefaultFromName based on the provided address.
// When addr contains a name component, that name is stored. Otherwise the
// default "Goa4Web" is used.
func SetDefaultFromName(addr string) {
	a, err := mail.ParseAddress(addr)
	if err == nil && a.Name != "" {
		DefaultFromName = a.Name
	} else {
		DefaultFromName = "Goa4Web"
	}
}

// BuildMessage constructs a MIME email message with optional HTML content.
// The from and to parameters can include optional Name fields which will be
// formatted as "Name <address>" in the resulting headers.
func BuildMessage(from, to mail.Address, subject, textBody, htmlBody string) ([]byte, error) {
	var msg bytes.Buffer
	hdr := textproto.MIMEHeader{}
	if from.Name == "" {
		from.Name = DefaultFromName
	}
	if from.Name != "" {
		enc := mail.Address{Name: mime.QEncoding.Encode("utf-8", from.Name), Address: from.Address}
		hdr.Set("From", enc.String())
	} else {
		hdr.Set("From", from.Address)
	}
	if to.Name != "" {
		enc := mail.Address{Name: mime.QEncoding.Encode("utf-8", to.Name), Address: to.Address}
		hdr.Set("To", enc.String())
	} else {
		hdr.Set("To", to.Address)
	}
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

package email

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
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

// BuildMessage constructs a MIME email message with optional HTML content.
func BuildMessage(from, to mail.Address, subject, textBody, htmlBody string) ([]byte, error) {
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
		part, err := w.CreatePart(textproto.MIMEHeader{
			"Content-Type":              {"text/plain; charset=utf-8"},
			"Content-Transfer-Encoding": {"quoted-printable"},
		})
		if err != nil {
			return nil, err
		}
		if err := writeQuotedPrintable(part, textBody); err != nil {
			return nil, err
		}
		part, err = w.CreatePart(textproto.MIMEHeader{
			"Content-Type":              {"text/html; charset=utf-8"},
			"Content-Transfer-Encoding": {"quoted-printable"},
		})
		if err != nil {
			return nil, err
		}
		if err := writeQuotedPrintable(part, htmlBody); err != nil {
			return nil, err
		}
		if err := w.Close(); err != nil {
			return nil, err
		}
	} else {
		hdr.Set("Content-Type", "text/plain; charset=utf-8")
		hdr.Set("Content-Transfer-Encoding", "quoted-printable")
		for k, v := range hdr {
			for _, vv := range v {
				fmt.Fprintf(&msg, "%s: %s\r\n", k, vv)
			}
		}
		msg.WriteString("\r\n")
		if err := writeQuotedPrintable(&msg, textBody); err != nil {
			return nil, err
		}
	}
	return msg.Bytes(), nil
}

// writeQuotedPrintable encodes body as MIME quoted-printable text.
func writeQuotedPrintable(dst io.Writer, body string) error {
	writer := quotedprintable.NewWriter(dst)
	if _, err := io.WriteString(writer, body); err != nil {
		return err
	}
	return writer.Close()
}

package email

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"testing"
)

func TestBuildMessageEncodesPlainTextAsQuotedPrintable(t *testing.T) {
	textBody := "Hello, café — こんにちは"
	message, err := BuildMessage(mail.Address{Address: "from@example.com"}, mail.Address{Address: "to@example.com"}, "Subject", textBody, "")
	if err != nil {
		t.Fatalf("BuildMessage: %v", err)
	}
	parsed, err := mail.ReadMessage(bytes.NewReader(message))
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	if got := parsed.Header.Get("Content-Transfer-Encoding"); got != "quoted-printable" {
		t.Fatalf("Content-Transfer-Encoding=%q, want quoted-printable", got)
	}
	decoded, err := io.ReadAll(quotedprintable.NewReader(parsed.Body))
	if err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if got := string(decoded); got != textBody {
		t.Errorf("decoded body=%q, want %q", got, textBody)
	}
}

func TestBuildMessageEncodesMultipartPartsAsQuotedPrintable(t *testing.T) {
	textBody := "Plain café"
	htmlBody := "<p>HTML こんにちは</p>"
	message, err := BuildMessage(mail.Address{Address: "from@example.com"}, mail.Address{Address: "to@example.com"}, "Subject", textBody, htmlBody)
	if err != nil {
		t.Fatalf("BuildMessage: %v", err)
	}
	parsed, err := mail.ReadMessage(bytes.NewReader(message))
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	if got := parsed.Header.Get("Content-Transfer-Encoding"); got != "" {
		t.Fatalf("multipart Content-Transfer-Encoding=%q, want empty", got)
	}
	mediaType, params, err := mime.ParseMediaType(parsed.Header.Get("Content-Type"))
	if err != nil {
		t.Fatalf("parse Content-Type: %v", err)
	}
	if mediaType != "multipart/alternative" {
		t.Fatalf("media type=%q, want multipart/alternative", mediaType)
	}

	parts := map[string]string{}
	reader := multipart.NewReader(parsed.Body, params["boundary"])
	for {
		part, err := reader.NextRawPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("read part: %v", err)
		}
		if got := part.Header.Get("Content-Transfer-Encoding"); got != "quoted-printable" {
			t.Fatalf("part Content-Transfer-Encoding=%q, want quoted-printable", got)
		}
		decoded, err := io.ReadAll(quotedprintable.NewReader(part))
		if err != nil {
			t.Fatalf("decode part: %v", err)
		}
		parts[part.Header.Get("Content-Type")] = string(decoded)
	}
	if got := parts["text/plain; charset=utf-8"]; got != textBody {
		t.Errorf("plain part=%q, want %q", got, textBody)
	}
	if got := parts["text/html; charset=utf-8"]; got != htmlBody {
		t.Errorf("html part=%q, want %q", got, htmlBody)
	}
}

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

func TestBuildMessageMetadata(t *testing.T) {
	from := mail.Address{Name: "Sender", Address: "sender@example.com"}
	to := mail.Address{Name: "Recipient", Address: "recipient@example.com"}
	replyTo := mail.Address{Name: "Replies", Address: "replies@example.com"}

	t.Run("default omits optional headers", func(t *testing.T) {
		message, err := BuildMessage(from, to, "Subject", "Text", "")
		if err != nil {
			t.Fatalf("BuildMessage: %v", err)
		}
		parsed, err := mail.ReadMessage(bytes.NewReader(message))
		if err != nil {
			t.Fatalf("ReadMessage: %v", err)
		}
		if got := parsed.Header.Get("Auto-Submitted"); got != "" {
			t.Errorf("Auto-Submitted=%q, want empty", got)
		}
		if got := parsed.Header.Get("Reply-To"); got != "" {
			t.Errorf("Reply-To=%q, want empty", got)
		}
		if got := parsed.Header.Get("List-Unsubscribe"); got != "" {
			t.Errorf("List-Unsubscribe=%q, want empty", got)
		}
	})

	t.Run("requested metadata is emitted", func(t *testing.T) {
		message, err := BuildMessage(from, to, "Subject", "Text", "", WithAutoSubmitted(), WithReplyTo(replyTo), WithListUnsubscribe("https://example.com/usr/subscriptions"))
		if err != nil {
			t.Fatalf("BuildMessage: %v", err)
		}
		parsed, err := mail.ReadMessage(bytes.NewReader(message))
		if err != nil {
			t.Fatalf("ReadMessage: %v", err)
		}
		if got := parsed.Header.Get("Auto-Submitted"); got != "auto-generated" {
			t.Errorf("Auto-Submitted=%q, want auto-generated", got)
		}
		if got := parsed.Header.Get("Reply-To"); got != replyTo.String() {
			t.Errorf("Reply-To=%q, want %q", got, replyTo.String())
		}
		if got := parsed.Header.Get("List-Unsubscribe"); got != "<https://example.com/usr/subscriptions>" {
			t.Errorf("List-Unsubscribe=%q, want subscription URL", got)
		}
	})

	t.Run("invalid reply address is rejected", func(t *testing.T) {
		_, err := BuildMessage(from, to, "Subject", "Text", "", WithReplyTo(mail.Address{Address: "not an address"}))
		if err == nil {
			t.Fatal("BuildMessage succeeded with invalid Reply-To address")
		}
	})

	t.Run("invalid unsubscribe URL is rejected", func(t *testing.T) {
		for _, bad := range []string{
			"http://example.com/unsubscribe",
			"https://example.com/u>\r\nBcc:evil@example.com",
			"https://example.com/u<injected>",
		} {
			if _, err := BuildMessage(from, to, "Subject", "Text", "", WithListUnsubscribe(bad)); err == nil {
				t.Fatalf("BuildMessage accepted invalid List-Unsubscribe URL %q", bad)
			}
		}
	})
}

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

package email

import (
	"bytes"
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
	})

	t.Run("requested metadata is emitted", func(t *testing.T) {
		message, err := BuildMessage(from, to, "Subject", "Text", "", WithAutoSubmitted(), WithReplyTo(replyTo))
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
	})

	t.Run("invalid reply address is rejected", func(t *testing.T) {
		_, err := BuildMessage(from, to, "Subject", "Text", "", WithReplyTo(mail.Address{Address: "not an address"}))
		if err == nil {
			t.Fatal("BuildMessage succeeded with invalid Reply-To address")
		}
	})
}

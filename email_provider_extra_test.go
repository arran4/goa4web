package main

import (
	"os"
	"testing"
)

func TestGetEmailProviderSMTP(t *testing.T) {
	os.Setenv("EMAIL_PROVIDER", "smtp")
	os.Setenv("SMTP_HOST", "localhost")
	os.Setenv("SMTP_PORT", "25")
	t.Cleanup(func() {
		os.Unsetenv("EMAIL_PROVIDER")
		os.Unsetenv("SMTP_HOST")
		os.Unsetenv("SMTP_PORT")
	})
	p := getEmailProvider()
	s, ok := p.(smtpMailProvider)
	if !ok {
		t.Fatalf("expected smtpMailProvider, got %#v", p)
	}
	if s.addr != "localhost:25" || s.from != SourceEmail {
		t.Errorf("unexpected provider values: %#v", s)
	}
}

func TestGetEmailProviderLocal(t *testing.T) {
	os.Setenv("EMAIL_PROVIDER", "local")
	t.Cleanup(func() { os.Unsetenv("EMAIL_PROVIDER") })
	if _, ok := getEmailProvider().(localMailProvider); !ok {
		t.Fatalf("expected localMailProvider")
	}
}

func TestGetEmailProviderJMAP(t *testing.T) {
	os.Setenv("EMAIL_PROVIDER", "jmap")
	os.Setenv("JMAP_ENDPOINT", "http://example.com")
	os.Setenv("JMAP_ACCOUNT", "acct")
	os.Setenv("JMAP_IDENTITY", "id")
	t.Cleanup(func() {
		os.Unsetenv("EMAIL_PROVIDER")
		os.Unsetenv("JMAP_ENDPOINT")
		os.Unsetenv("JMAP_ACCOUNT")
		os.Unsetenv("JMAP_IDENTITY")
	})
	p := getEmailProvider()
	j, ok := p.(jmapMailProvider)
	if !ok {
		t.Fatalf("expected jmapMailProvider, got %#v", p)
	}
	if j.endpoint != "http://example.com" || j.accountID != "acct" || j.identity != "id" {
		t.Errorf("unexpected provider values: %#v", j)
	}
}

func TestGetEmailProviderSESNoCreds(t *testing.T) {
	os.Setenv("EMAIL_PROVIDER", "ses")
	os.Setenv("AWS_REGION", "us-east-1")
	t.Cleanup(func() {
		os.Unsetenv("EMAIL_PROVIDER")
		os.Unsetenv("AWS_REGION")
	})
	if p := getEmailProvider(); p != nil {
		t.Errorf("expected nil provider, got %#v", p)
	}
}

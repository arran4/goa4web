package main

import "testing"

func TestGetEmailProviderSMTP(t *testing.T) {
	p := providerFromConfig(RuntimeConfig{
		EmailProvider: "smtp",
		EmailSMTPHost: "localhost",
		EmailSMTPPort: "25",
	})
	s, ok := p.(smtpMailProvider)
	if !ok {
		t.Fatalf("expected smtpMailProvider, got %#v", p)
	}
	if s.addr != "localhost:25" || s.from != SourceEmail {
		t.Errorf("unexpected provider values: %#v", s)
	}
}

func TestGetEmailProviderLocal(t *testing.T) {
	if _, ok := providerFromConfig(RuntimeConfig{EmailProvider: "local"}).(localMailProvider); !ok {
		t.Fatalf("expected localMailProvider")
	}
}

func TestGetEmailProviderJMAP(t *testing.T) {
	p := providerFromConfig(RuntimeConfig{
		EmailProvider:     "jmap",
		EmailJMAPEndpoint: "http://example.com",
		EmailJMAPAccount:  "acct",
		EmailJMAPIdentity: "id",
	})
	j, ok := p.(jmapMailProvider)
	if !ok {
		t.Fatalf("expected jmapMailProvider, got %#v", p)
	}
	if j.endpoint != "http://example.com" || j.accountID != "acct" || j.identity != "id" {
		t.Errorf("unexpected provider values: %#v", j)
	}
}

func TestGetEmailProviderSESNoCreds(t *testing.T) {
	if p := providerFromConfig(RuntimeConfig{EmailProvider: "ses", EmailAWSRegion: "us-east-1"}); p != nil {
		t.Errorf("expected nil provider, got %#v", p)
	}
}

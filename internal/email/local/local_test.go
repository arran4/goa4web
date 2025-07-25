package local

import (
	"context"
	"net/mail"
	"testing"
)

func TestProviderInvalidAddr(t *testing.T) {
	p := Provider{}
	cases := []string{
		"foo@example.com\n",
		"foo@example.com,bar@example.com",
		"foo\x01@example.com",
	}
	for _, c := range cases {
		if err := p.Send(context.Background(), mail.Address{Address: c}, nil); err == nil {
			t.Errorf("expected error for %q", c)
		}
	}
}

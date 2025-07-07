package mock

import (
	"context"
	"testing"
)

func TestProvider(t *testing.T) {
	p := &Provider{}
	if err := p.Record(context.Background(), "msg"); err != nil {
		t.Fatalf("record: %v", err)
	}
	if len(p.Records) != 1 {
		t.Fatalf("records len=%d", len(p.Records))
	}
	if p.Records[0].Message != "msg" {
		t.Fatalf("unexpected message: %q", p.Records[0].Message)
	}
}

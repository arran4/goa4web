package sharesign_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/sharesign"
)

func TestSigner(t *testing.T) {
	cfg := &config.RuntimeConfig{
		HTTPHostname: "http://localhost:8080",
	}
	s := sharesign.NewSigner(cfg, "secret")
	link := "/news/news/1"
	ts, sig := s.Sign(link)
	if !s.Verify(link, fmt.Sprint(ts), sig) {
		t.Errorf("Verify failed")
	}
	if s.Verify(link, fmt.Sprint(ts), "invalid") {
		t.Errorf("Verify succeeded with invalid signature")
	}
	if s.Verify("invalid", fmt.Sprint(ts), sig) {
		t.Errorf("Verify succeeded with invalid link")
	}
	if s.Verify(link, fmt.Sprint(ts+1), sig) {
		t.Errorf("Verify succeeded with invalid timestamp")
	}
	if s.Verify(link, fmt.Sprint(time.Now().Add(-48*time.Hour).Unix()), sig) {
		t.Errorf("Verify succeeded with expired timestamp")
	}
}

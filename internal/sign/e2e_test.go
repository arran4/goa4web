package sign_test

import (
	"context"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers/share"
	"github.com/arran4/goa4web/internal/sign"
	"github.com/gorilla/mux"
)

func TestSigningE2E(t *testing.T) {
	cfg := &config.RuntimeConfig{
		HTTPHostname: "https://example.com",
	}
	cd := common.NewCoreData(context.Background(), nil, cfg)
	cd.ShareSignKey = "share-key"
	cd.ImageSignKey = "image-key"
	cd.LinkSignKey = "link-key"
	cd.FeedSignKey = "feed-key"

	t.Run("ShareURL_Path", func(t *testing.T) {
		path := "/private/topic/1"
		signedURL, err := cd.SignShareURL(path, sign.WithNonce("test-nonce"))
		if err != nil {
			t.Fatalf("SignShareURL failed: %v", err)
		}

		u, _ := url.Parse(signedURL)
		req := httptest.NewRequest("GET", u.String(), nil)
		// Extract signature parts from path
		parts := strings.Split(u.Path, "/")
		vars := map[string]string{
			"nonce": "test-nonce",
			"sign":  parts[len(parts)-1],
		}
		// In a real app, mux would populate these. We'll simulate VerifyAndGetPath logic.
		req = mux.SetURLVars(req, vars)

		verifiedPath := share.VerifyAndGetPath(req, cd.ShareSignKey)
		if verifiedPath == "" {
			t.Errorf("VerifyAndGetPath failed for path-based share")
		}
	})

	t.Run("ImageURL_Prefix", func(t *testing.T) {
		imageID := "abc-123.png"
		ttl := 1 * time.Hour
		signedURL := cd.SignImageURL(imageID, ttl)

		u, _ := url.Parse(signedURL)

		_, sig, opts, err := sign.ExtractQuerySig(u.String())
		if err != nil || sig == "" {
			t.Fatalf("ExtractQuerySig failed: %v", err)
		}

		data := "image:" + imageID
		if err := sign.Verify(data, sig, cd.ImageSignKey, opts...); err != nil {
			t.Errorf("Verify failed for image: %v", err)
		}
	})

	t.Run("IssuedAt_its", func(t *testing.T) {
		data := "test-data"
		now := time.Now().Truncate(time.Second)
		sig := sign.Sign(data, "key", sign.WithIssuedAt(now))

		u, _ := sign.AddQuerySig("https://example.com", sig, sign.WithIssuedAt(now))

		_, extSig, opts, err := sign.ExtractQuerySig(u)
		if err != nil {
			t.Fatalf("ExtractQuerySig failed: %v", err)
		}

		if err := sign.Verify(data, extSig, "key", opts...); err != nil {
			t.Errorf("Verify with its failed: %v", err)
		}
	})

	t.Run("AbsoluteExpiry_ets", func(t *testing.T) {
		data := "test-data"
		future := time.Now().Add(1 * time.Hour).Truncate(time.Second)
		sig := sign.Sign(data, "key", sign.WithAbsoluteExpiry(future))

		u, _ := sign.AddQuerySig("https://example.com", sig, sign.WithAbsoluteExpiry(future))

		_, extSig, opts, err := sign.ExtractQuerySig(u)
		if err != nil {
			t.Fatalf("ExtractQuerySig failed: %v", err)
		}

		// Should check ets
		if err := sign.Verify(data, extSig, "key", opts...); err != nil {
			t.Errorf("Verify with ets failed: %v", err)
		}
	})
}

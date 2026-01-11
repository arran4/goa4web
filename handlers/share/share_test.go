package share

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/sharesign"
)

func TestShareLinkUsesCoreDataSigner(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	cfg.HTTPHostname = "http://example.com"
	cfg.ShareSignSecret = "wrong-secret"

	signer := sharesign.NewSigner(cfg, "right-secret")
	cd := common.NewCoreData(context.Background(), nil, cfg, common.WithShareSigner(signer))

	req := httptest.NewRequest(http.MethodGet, "/api/forum/share?use_query=true&link="+url.QueryEscape("/private/topic/1/thread/2"), nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	rr := httptest.NewRecorder()

	ShareLink(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d got %d", http.StatusOK, rr.Code)
	}

	var payload struct {
		SignedURL string `json:"signed_url"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.SignedURL == "" {
		t.Fatalf("expected signed URL")
	}

	signedReq, err := http.NewRequest(http.MethodGet, payload.SignedURL, nil)
	if err != nil {
		t.Fatalf("new signed request: %v", err)
	}
	if got := VerifyAndGetPath(signedReq, signer); got == "" {
		t.Fatalf("expected signature to verify with core data signer")
	}

	wrongSigner := sharesign.NewSigner(cfg, "wrong-secret")
	if got := VerifyAndGetPath(signedReq, wrongSigner); got != "" {
		t.Fatalf("expected signature to fail with wrong signer")
	}
}

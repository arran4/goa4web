package externallink

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/sign"
)

func TestReloadExternalLinkTask_Action(t *testing.T) {
	cfg := config.NewRuntimeConfig()
	key := "testkey"
	link := "https://example.com/some/link"
	sig := sign.Sign("link:"+link, key)

	client := &http.Client{
		Transport: &mockTransport{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`<html><head><meta property="og:title" content="Test Title"/></head><body></body></html>`)),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	querier := &mockQuerier{}

	cd := common.NewCoreData(context.Background(), querier, cfg, func(cd *common.CoreData) {
		cd.LinkSignKey = key
	}, common.WithHTTPClient(client))

	u := "/?u=" + url.QueryEscape(link) + "&sig=" + sig
	req := httptest.NewRequest(http.MethodPost, u, nil)

	// Populate form for FormValue calls in the task
	req.ParseForm()

	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	rec := httptest.NewRecorder()

	res := reloadExternalLinkTask.Action(rec, req)

	// New behavior: returns handlers.RedirectHandler
	if _, ok := res.(handlers.RedirectHandler); !ok {
		t.Fatalf("Expected handlers.RedirectHandler, got %T", res)
	}

	val := res.(handlers.RedirectHandler)
	// Expect redirect to request URI (which includes query params)
	expectedURL := "/?u=" + url.QueryEscape(link) + "&sig=" + sig
	if string(val) != expectedURL {
		t.Errorf("Expected '%s', got '%s'", expectedURL, string(val))
	}
}

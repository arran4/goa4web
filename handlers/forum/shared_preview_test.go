package forum

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/sign"
	"github.com/arran4/goa4web/internal/sign/signutil"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/gorilla/mux"
)

func TestSharedTopicPreviewPage_GuestRedirectsToLogin(t *testing.T) {
	// Setup CoreData and dependencies
	queries := testhelpers.NewQuerierStub()
	queries.GetForumTopicByIdReturns = &db.Forumtopic{
		Idforumtopic: 1,
		Title:        sql.NullString{String: "Test Topic", Valid: true},
		Description:  sql.NullString{String: "Test Description", Valid: true},
	}

	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(context.Background(), queries, cfg)
	cd.ShareSignKey = "test-secret-key"
	cd.UserID = 0 // Guest

	// Generate a valid signed URL path
	// The handler expects /shared/topic/{topic}/nonce/{nonce}/sign/{sign}
	topicID := "1"
	nonce := signutil.GenerateNonce()

	// Create the signature.
	opts := []sign.SignOption{sign.WithNonce(nonce)}
	sig := sign.Sign("/forum/shared/topic/"+topicID, cd.ShareSignKey, opts...)

	// Create request
	reqURL := "/forum/shared/topic/" + topicID + "/nonce/" + nonce + "/sign/" + sig
	req := httptest.NewRequest("GET", reqURL, nil)

	// Set vars as if matched by router
	vars := map[string]string{
		"topic": topicID,
		"nonce": nonce,
		"sign":  sig,
	}
	req = mux.SetURLVars(req, vars)

	// Inject CoreData
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Call handler
	SharedTopicPreviewPage(w, req)

	// Check response code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 (OK) for preview page, got %d", w.Code)
	}

	body := w.Body.String()

	// Verify Meta Refresh
	// The URL in meta refresh might be HTML escaped (e.g. & matches &amp;), but url.QueryEscape handles special chars.
	// /login?return_url=...
	expectedReturnURL := url.QueryEscape(reqURL)
	expectedRefresh := fmt.Sprintf("content=\"0;url=/login?return_url=%s\"", expectedReturnURL)

	if !strings.Contains(body, expectedRefresh) {
		t.Errorf("Expected meta refresh to login. Got body:\n%s\nExpected to contain: %s", body, expectedRefresh)
	}

	// Verify OG URL (should be the content URL)
	// The code sets ContentURL to cd.AbsoluteURL(r.URL.RequestURI())
	// Since we didn't configure BaseURL in config, AbsoluteURL uses DefaultBaseURL or request host.
	// check if og:url matches the request URL (roughly)
	if !strings.Contains(body, fmt.Sprintf("<meta property=\"og:url\" content=\"%s\" />", cd.AbsoluteURL(reqURL))) {
		t.Errorf("Expected og:url to match request URL. Got body:\n%s", body)
	}
}

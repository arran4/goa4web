package user_test

import (
	"bytes"
	"context"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/subscriptions"
)

func TestSubscriptionsTemplateRender(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	cd := &common.CoreData{Config: config.NewRuntimeConfig()}
	// Funcs might rely on context value for cd
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	tmpl := templates.GetCompiledSiteTemplates(cd.Funcs(req))

	data := struct {
		Groups []*subscriptions.SubscriptionGroup
	}{
		Groups: []*subscriptions.SubscriptionGroup{
			{
				Definition: &subscriptions.Definition{
					Name:        "Test Sub",
					Description: "Test Desc",
					Pattern:     "test:/test/*",
				},
				Instances: []*subscriptions.SubscriptionInstance{
					{
						Original: "test:/test/1",
						Methods:  []string{"internal"},
						Parameters: []subscriptions.Parameter{
							{Key: "id", Value: "1"},
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "user/subscriptions.gohtml", data); err != nil {
		t.Fatalf("render user/subscriptions.gohtml: %v", err)
	}

	if !bytes.Contains(buf.Bytes(), []byte("<form method=\"post\" action=\"/usr/subscriptions/update\">")) {
		t.Errorf("Expected form tag not found")
	}
	// Check for the checkbox with correct value
	expectedValue := "test:/test/1|internal"
	if !bytes.Contains(buf.Bytes(), []byte(expectedValue)) {
		t.Errorf("Expected checkbox value %q not found", expectedValue)
	}
	// Check for hidden presented input
	if !bytes.Contains(buf.Bytes(), []byte("name=\"presented_subs\" value=\"test:/test/1|internal\"")) {
		t.Errorf("Expected hidden presented input not found")
	}
}

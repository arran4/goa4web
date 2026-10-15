package common_test

import (
	"context"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
)

func TestIncludeUsesInjectedSiteTemplateRenderer(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	called := false
	cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig(), common.WithSiteTemplateRenderer(func(_ *common.CoreData, w io.Writer, _ *http.Request, _ string, _ any) error {
		called = true
		_, err := io.WriteString(w, "included")
		return err
	}))

	include, ok := cd.Funcs(req)["include"].(func(string, any) (template.HTML, error))
	if !ok {
		t.Fatal("include helper has unexpected type")
	}
	got, err := include("fragment", nil)
	if err != nil {
		t.Fatalf("include: %v", err)
	}
	if !called || got != template.HTML("included") {
		t.Fatalf("include result = %q, called = %v", got, called)
	}
}

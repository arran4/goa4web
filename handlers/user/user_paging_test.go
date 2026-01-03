package user

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

func TestUserPagingPage_Render(t *testing.T) {
	req := httptest.NewRequest("GET", "/usr/paging", nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, nil, config.NewRuntimeConfig())
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	userPagingPage(rr, req)

	// TemplateHandler doesn't set status 500 on template error, but it renders an error page.
	// We check if the body contains the specific error message reported by the user.
	if strings.Contains(rr.Body.String(), `html/template: "pagingPage.gohtml" is undefined`) {
		t.Fatalf("Template not defined error detected in response: %s", rr.Body.String())
	}

	if rr.Code != http.StatusOK {
		t.Fatalf("Status code: %d, Body: %s", rr.Code, rr.Body.String())
	}
}

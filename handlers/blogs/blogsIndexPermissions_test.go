package blogs

import (
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
)

func TestCustomBlogIndexRoles(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs", nil)

	cd := common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}))
	cd.AdminMode = true
	BlogsMiddlewareIndex(cd, req)
	if !common.ContainsItem(cd.CustomIndexItems, "Blogs Admin") {
		t.Errorf("admin should see blogs admin link")
	}
	if !common.ContainsItem(cd.CustomIndexItems, "Write blog") {
		t.Errorf("admin should see write blog")
	}

	cd = common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"content writer"}))
	BlogsMiddlewareIndex(cd, req)
	if common.ContainsItem(cd.CustomIndexItems, "Blogs Admin") {
		t.Errorf("content writer should not see blogs admin link")
	}
	if !common.ContainsItem(cd.CustomIndexItems, "Write blog") {
		t.Errorf("content writer should see write blog")
	}

	cd = common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"anyone"}))
	BlogsMiddlewareIndex(cd, req)
	if common.ContainsItem(cd.CustomIndexItems, "Blogs Admin") || common.ContainsItem(cd.CustomIndexItems, "Write blog") {
		t.Errorf("anyone should not see writer/admin items")
	}
}

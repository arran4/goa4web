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
	CustomBlogIndex(cd, req)
	if !common.ContainsItem(cd.CustomIndexItems, "Blogs Admin") {
		t.Errorf("admin should see blogs admin link")
	}
	if !common.ContainsItem(cd.CustomIndexItems, "Write blog") {
		t.Errorf("admin should see write blog")
	}

	cd = common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"content writer"}))
	CustomBlogIndex(cd, req)
	if common.ContainsItem(cd.CustomIndexItems, "Blogs Admin") {
		t.Errorf("content writer should not see blogs admin link")
	}
	if !common.ContainsItem(cd.CustomIndexItems, "Write blog") {
		t.Errorf("content writer should see write blog")
	}

	cd = common.NewCoreData(req.Context(), nil, config.NewRuntimeConfig(), common.WithUserRoles([]string{"anonymous"}))
	CustomBlogIndex(cd, req)
	if common.ContainsItem(cd.CustomIndexItems, "Blogs Admin") || common.ContainsItem(cd.CustomIndexItems, "Write blog") {
		t.Errorf("anonymous should not see writer/admin items")
	}
}

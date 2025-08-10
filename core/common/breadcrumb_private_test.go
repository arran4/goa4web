package common

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/config"
)

func TestPrivateForumBreadcrumbBasePath(t *testing.T) {
	cd := NewCoreData(context.Background(), nil, config.NewRuntimeConfig())
	cd.SetCurrentSection("privateforum")
	cd.ForumBasePath = "/private"
	crumbs, err := cd.forumBreadcrumbs()
	if err != nil {
		t.Fatalf("forumBreadcrumbs error: %v", err)
	}
	if len(crumbs) == 0 || crumbs[0].Link != "/private" {
		t.Fatalf("crumbs=%v", crumbs)
	}
}

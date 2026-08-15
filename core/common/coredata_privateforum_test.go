package common

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestCreatePrivateForumCommentUsesPrivateThreadGrant(t *testing.T) {
	queries := testhelpers.NewQuerierStub()
	queries.CreateCommentInSectionForCommenterResult = 42
	cd := NewCoreData(context.Background(), queries, config.NewRuntimeConfig())

	commentID, err := cd.CreatePrivateForumCommentForCommenter(1, 4, 2, 0, "new private thread")
	if err != nil {
		t.Fatalf("CreatePrivateForumCommentForCommenter: %v", err)
	}
	if commentID != 42 {
		t.Fatalf("comment ID = %d, want 42", commentID)
	}
	if len(queries.CreateCommentInSectionForCommenterCalls) != 1 {
		t.Fatalf("comment insert calls = %d, want 1", len(queries.CreateCommentInSectionForCommenterCalls))
	}

	call := queries.CreateCommentInSectionForCommenterCalls[0]
	if call.Section != consts.PermissionSectionPrivateForumThread.String() {
		t.Errorf("grant section = %q, want privateforum_thread", call.Section)
	}
	if !call.ItemType.Valid || call.ItemType.String != consts.PermissionItemThread.String() {
		t.Errorf("grant item = %#v, want thread", call.ItemType)
	}
	if !call.ItemID.Valid || call.ItemID.Int32 != 4 {
		t.Errorf("grant item ID = %#v, want thread ID 4", call.ItemID)
	}
}

package permissions

import (
	"testing"
)

func TestAppendPermissionsRegistered(t *testing.T) {
	var forumFound, privateFound int

	for _, def := range Definitions {
		if def.Section == "forum" && def.Item == "topic" && def.Action == "append" {
			forumFound++
			if !def.RequireItemID {
				t.Errorf("ForumTopicAppend must require ItemID")
			}
		}
		if def.Section == "privateforum_thread" && def.Item == "thread" && def.Action == "append" {
			privateFound++
			if !def.RequireItemID {
				t.Errorf("PrivateforumThreadAppend must require ItemID")
			}
		}
	}

	if forumFound != 1 {
		t.Errorf("expected ForumTopicAppend to be registered exactly once, found %d times", forumFound)
	}
	if privateFound != 1 {
		t.Errorf("expected PrivateforumThreadAppend to be registered exactly once, found %d times", privateFound)
	}
}

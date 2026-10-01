//go:build sqlite || sqlite3

package main

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/scenario"
	"github.com/arran4/goa4web/testdata/scenarios"
	"github.com/stretchr/testify/require"
)

func TestE2EPrivateForumEditsAndLabels(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	root, err := parseRoot([]string{"goa4web", "scenario", "serve", "100-private-forum"})
	require.NoError(t, err)
	defer root.Close()

	parent, err := parseScenarioCmd(root, []string{"serve", "100-private-forum"})
	require.NoError(t, err)
	serveCmd, err := parseScenarioServeCmd(parent, []string{"100-private-forum"})
	require.NoError(t, err)
	serveCmd.fsys = scenarios.FS

	_, _, cleanup, err := serveCmd.Bootstrap(ctx)
	require.NoError(t, err)
	defer cleanup()

	reg := serveCmd.Registry
	appCD := serveCmd.CoreData

	// Resolve the participants via the registry
	aliceID := getScenarioUserID(t, reg, "alice")
	bobID := getScenarioUserID(t, reg, "bob")
	carolID := getScenarioUserID(t, reg, "carol")
	daveID := getScenarioUserID(t, reg, "dave")

	// Resolve the topics/threads
	staffRoomTopicID := getScenarioEntityID(t, reg, scenario.RefTypeForum, "staff-room")
	staffWelcomeThreadID := getScenarioEntityID(t, reg, scenario.RefTypeThread, "staff-welcome")
	bobWelcomeReplyID := getScenarioEntityID(t, reg, scenario.RefTypePost, "bob-welcome-reply")

	// 1. Verify permitted edits changed state
	t.Run("Verify Permitted Edits", func(t *testing.T) {
		aliceCD := appCD.ForUser(aliceID)
		topic, err := aliceCD.Queries().GetForumTopicById(ctx, staffRoomTopicID)
		if err != nil {
			t.Fatalf("Failed to fetch staff-room topic: %v", err)
		}

		if topic.Title.String != "Staff Room - Edited Title" {
			t.Errorf("Expected topic title 'Staff Room - Edited Title', got %q", topic.Title.String)
		}
		if topic.Description.String != "Confidential team discussions. Edited." {
			t.Errorf("Expected topic description 'Confidential team discussions. Edited.', got %q", topic.Description.String)
		}

		bobCD := appCD.ForUser(bobID)
		bobReply, err := bobCD.Queries().GetCommentByIdForUser(ctx, db.GetCommentByIdForUserParams{ ViewerID: bobID, ID: bobWelcomeReplyID, UserID: sql.NullInt32{Int32: bobID, Valid: bobID != 0} })
		if err != nil {
			t.Fatalf("Failed to fetch bob's reply: %v", err)
		}

		if !strings.Contains(bobReply.Text.String, "Thanks Alice! Edited.") {
			t.Errorf("Expected bob's reply to contain 'Thanks Alice! Edited.', got %q", bobReply.Text.String)
		}
	})

	// 2. Verify Private Label Isolation
	t.Run("Verify Private Label Isolation", func(t *testing.T) {
		aliceCD := appCD.ForUser(aliceID)
		// Instead of thread.Labels, we use the specific label retrieval methods on CD.
		// Note: the test just verified the application paths work; the DB state is the source of truth for the test.
		alicePrivate, err := aliceCD.ThreadPrivateLabels(staffWelcomeThreadID, aliceID)
		if err != nil {
			t.Fatalf("Failed to fetch private labels for Alice: %v", err)
		}
		hasAlicePrivateLabel := false
		for _, lbl := range alicePrivate {
			if lbl == "alice-followup" {
				hasAlicePrivateLabel = true
				break
			}
		}
		if !hasAlicePrivateLabel {
			t.Errorf("Expected Alice to see her private label 'alice-followup', labels: %v", alicePrivate)
		}

		bobCD := appCD.ForUser(bobID)
		bobPrivate, err := bobCD.ThreadPrivateLabels(staffWelcomeThreadID, bobID)
		if err != nil {
			t.Fatalf("Failed to fetch private labels for Bob: %v", err)
		}
		for _, lbl := range bobPrivate {
			if lbl == "alice-followup" {
				t.Errorf("Bob should not see Alice's private label 'alice-followup'")
			}
		}
	})

	// 3. Verify Unauthorized Mutations are Rejected
	t.Run("Verify Unauthorized Mutations Rejected", func(t *testing.T) {
		daveCD := appCD.ForUser(daveID) // Dave is not a participant

		// Attempt to edit a topic Dave cannot access
		err := daveCD.EditPrivateTopic(ctx, common.EditPrivateTopicParams{
			ActorID:     daveID,
			TopicID:     staffRoomTopicID,
			Title:       "Dave's Hack",
			Description: "Hacked",
		})
		if err == nil {
			t.Error("Expected error when Dave edits private topic, got nil")
		}

		// Verify state did not change
		freshAliceCD := appCD.ForUser(aliceID)
		topic, _ := freshAliceCD.Queries().GetForumTopicById(ctx, staffRoomTopicID)
		if topic.Title.String == "Dave's Hack" {
			t.Error("Dave successfully mutated topic title")
		}

		// Attempt to edit Bob's reply as Carol (Carol is participant but not author, no edit-any grant)
		carolCD := appCD.ForUser(carolID)
		err = carolCD.EditForumCommentAction(ctx, common.EditForumCommentParams{
			ActorID:    carolID,
			CommentID:  bobWelcomeReplyID,
			LanguageID: 1,
			Text:       "Carol edited this",
		})
		if err == nil {
			t.Error("Expected error when Carol edits Bob's reply, got nil")
		}

		// Verify state did not change
		freshBobCD := appCD.ForUser(bobID)
		bobReply, _ := freshBobCD.Queries().GetCommentByIdForUser(ctx, db.GetCommentByIdForUserParams{ ViewerID: bobID, ID: bobWelcomeReplyID, UserID: sql.NullInt32{Int32: bobID, Valid: bobID != 0} })
		if strings.Contains(bobReply.Text.String, "Carol edited this") {
			t.Error("Carol successfully mutated Bob's reply text")
		}
	})
}

// getScenarioUserID safely retrieves a user ID.
func getScenarioUserID(t *testing.T, reg *scenario.RefRegistry, name string) int32 {
	t.Helper()
	id, ok := reg.ResolveUser(name)
	if !ok {
		t.Fatalf("Failed to resolve user ID %q", name)
	}
	return id
}

// getScenarioEntityID safely retrieves an entity ID from the registry.
func getScenarioEntityID(t *testing.T, reg *scenario.RefRegistry, typ scenario.RefType, name string) int32 {
	t.Helper()
	val, ok := reg.Resolve(typ, name)
	if !ok {
		t.Fatalf("Failed to resolve %s ID %q", typ, name)
	}
	if i, ok := val.(int32); ok {
		return i
	}
	if i, ok := val.(int); ok {
		return int32(i)
	}
	t.Fatalf("Resolved %s ID %q but it is not an int/int32, got %T", typ, name, val)
	return 0
}

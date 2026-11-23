//go:build sqlite || sqlite3

package main

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/scenario"
	"github.com/arran4/goa4web/testdata/scenarios"
)

func TestScenarioServePrivateForumPermissionMatrix(t *testing.T) {
	ctx := context.Background()

	root, err := parseRoot([]string{"goa4web", "scenario", "serve", "100-private-forum"})
	if err != nil {
		t.Fatalf("parseRoot: %v", err)
	}
	defer root.Close()
	parent, err := parseScenarioCmd(root, []string{"serve", "100-private-forum"})
	if err != nil {
		t.Fatalf("parseScenarioCmd: %v", err)
	}
	serveCmd, err := parseScenarioServeCmd(parent, []string{"100-private-forum"})
	if err != nil {
		t.Fatalf("parseScenarioServeCmd: %v", err)
	}
	serveCmd.fsys = scenarios.FS

	srv, dbConn, cleanup, err := serveCmd.Bootstrap(ctx)
	if err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}
	defer cleanup()

	querier := db.NewForDriver(dbConn, "sqlite3")
	baseCD := common.NewCoreData(ctx, querier, srv.Config)
	userIDs := loadScenarioUserIDs(t, ctx, dbConn)
	topicIDs := loadScenarioTopicIDs(t, ctx, dbConn)
	threadIDs := loadScenarioThreadIDs(t, serveCmd.Registry)

	wantTopics := map[string][]string{
		"alice": {"Coordination", "Staff Room - Edited Title"},
		"bob":   {"Staff Room - Edited Title"},
		"carol": {"Coordination", "Project Room"},
		"dave":  {"Project Room"},
	}
	for username, expected := range wantTopics {
		t.Run("topic-list/"+username, func(t *testing.T) {
			userCD := baseCD.ForUser(userIDs[username])
			topics, err := userCD.PrivateForumTopics()
			if err != nil {
				t.Fatalf("PrivateForumTopics: %v", err)
			}
			got := make([]string, 0, len(topics))
			for _, topic := range topics {
				got = append(got, topic.DisplayTitle)
				if topic.TotalParticipants != 2 {
					t.Errorf("%q participant count = %d, want 2", topic.DisplayTitle, topic.TotalParticipants)
				}
			}
			sort.Strings(got)
			if fmt.Sprint(got) != fmt.Sprint(expected) {
				t.Fatalf("visible topics = %v, want %v", got, expected)
			}
		})
	}

	members := map[string]map[string]bool{
		"alice": {"Staff Room - Edited Title": true, "Coordination": true},
		"bob":   {"Staff Room - Edited Title": true},
		"carol": {"Project Room": true, "Coordination": true},
		"dave":  {"Project Room": true},
	}
	threadTopics := map[string]string{
		"staff-welcome":     "Staff Room - Edited Title",
		"staff-check-in":    "Staff Room - Edited Title",
		"project-kickoff":   "Project Room",
		"coordination-plan": "Coordination",
	}
	for username, userID := range userIDs {
		for topicName, topicID := range topicIDs {
			t.Run("topic-open/"+username+"/"+topicName, func(t *testing.T) {
				_, err := baseCD.ForUser(userID).ForumTopicByID(topicID)
				assertScenarioVisibility(t, err, members[username][topicName])
			})
		}
		for threadName, threadID := range threadIDs {
			t.Run("thread-open/"+username+"/"+threadName, func(t *testing.T) {
				_, err := baseCD.ForUser(userID).ForumThreadByID(threadID)
				assertScenarioVisibility(t, err, members[username][threadTopics[threadName]])
			})
		}
	}

	for username, userID := range userIDs {
		for threadName, threadID := range threadIDs {
			topicName := threadTopics[threadName]
			if members[username][topicName] {
				continue
			}
			t.Run("reply-denied/"+username+"/"+threadName, func(t *testing.T) {
				var before int
				if err := dbConn.QueryRowContext(ctx, "SELECT count(*) FROM comments WHERE forumthread_id = ?", threadID).Scan(&before); err != nil {
					t.Fatalf("count comments before reply: %v", err)
				}
				_, err := baseCD.ReplyForumThread(ctx, common.ReplyForumThreadParams{
					ActorID:        userID,
					ThreadID:       threadID,
					Text:           "This unauthorized reply must not be stored.",
					Private:        true,
					EnforceHandler: true,
				})
				if err == nil {
					t.Fatal("unauthorized reply unexpectedly succeeded")
				}
				var notFound common.ForumResourceNotFoundError
				var forbidden common.ForumOperationForbiddenError
				if !errors.As(err, &notFound) && !errors.As(err, &forbidden) {
					t.Fatalf("unauthorized reply error = %T %v, want not-found or forbidden", err, err)
				}
				var after int
				if err := dbConn.QueryRowContext(ctx, "SELECT count(*) FROM comments WHERE forumthread_id = ?", threadID).Scan(&after); err != nil {
					t.Fatalf("count comments after reply: %v", err)
				}
				if after != before {
					t.Fatalf("comment count changed from %d to %d after denied reply", before, after)
				}
			})
		}
	}

	// Assert explicit view/reply/append capabilities per thread based on the test plan
	type userGrants struct {
		view   bool
		reply  bool
		append bool
	}
	wantThreadGrants := map[string]map[string]userGrants{
		"staff-welcome": {
			"alice": {true, true, true},
			"bob":   {true, true, true},
			"carol": {false, false, false},
			"dave":  {false, false, false},
		},
		"staff-check-in": { // Only staff-welcome got the append grant via user.grant in 100-private-forum? Wait, the PR requested staff-check-in to also have it? The prompt says "staff-welcome: Alice/Bob" and "bob-staff-check-in: Alice/Bob". The seed uses staff-check-in. Let's give append to both.
			"alice": {true, true, true},
			"bob":   {true, true, true},
			"carol": {false, false, false},
			"dave":  {false, false, false},
		},
		"project-kickoff": {
			"alice": {false, false, false},
			"bob":   {false, false, false},
			"carol": {true, true, true},
			"dave":  {true, true, true},
		},
		"coordination-plan": {
			"alice": {true, true, true},
			"bob":   {false, false, false},
			"carol": {true, true, true},
			"dave":  {false, false, false},
		},
	}

	for threadName, expectedUsers := range wantThreadGrants {
		t.Run("thread-grants/"+threadName, func(t *testing.T) {
			for username, expected := range expectedUsers {
				hasGrant := func(action string) bool {
					var count int
					dbConn.QueryRowContext(ctx, "SELECT COUNT(*) FROM grants WHERE section = 'privateforum_thread' AND item = 'thread' AND item_id = ? AND action = ? AND user_id = ? AND active = 1", threadIDs[threadName], action, userIDs[username]).Scan(&count)
					return count > 0
				}

				gotView := hasGrant("view")
				gotReply := hasGrant("reply")
				gotAppend := hasGrant("append")

				if gotView != expected.view {
					t.Errorf("%s in %s: view = %v, want %v", username, threadName, gotView, expected.view)
				}
				if gotReply != expected.reply {
					t.Errorf("%s in %s: reply = %v, want %v", username, threadName, gotReply, expected.reply)
				}
				if gotAppend != expected.append {
					t.Errorf("%s in %s: append = %v, want %v", username, threadName, gotAppend, expected.append)
				}
			}
		})
	}
}

func loadScenarioUserIDs(t *testing.T, ctx context.Context, dbConn *sql.DB) map[string]int32 {
	t.Helper()
	ids := make(map[string]int32)
	for _, username := range []string{"alice", "bob", "carol", "dave"} {
		var id int32
		if err := dbConn.QueryRowContext(ctx, "SELECT idusers FROM users WHERE username = ?", username).Scan(&id); err != nil {
			t.Fatalf("query user %q: %v", username, err)
		}
		ids[username] = id
	}
	return ids
}

func loadScenarioTopicIDs(t *testing.T, ctx context.Context, dbConn *sql.DB) map[string]int32 {
	t.Helper()
	ids := make(map[string]int32)
	for _, title := range []string{"Staff Room - Edited Title", "Project Room", "Coordination"} {
		var id int32
		if err := dbConn.QueryRowContext(ctx, "SELECT idforumtopic FROM forumtopic WHERE title = ?", title).Scan(&id); err != nil {
			t.Fatalf("query topic %q: %v", title, err)
		}
		ids[title] = id
	}
	return ids
}

func loadScenarioThreadIDs(t *testing.T, reg *scenario.RefRegistry) map[string]int32 {
	t.Helper()
	if reg == nil {
		t.Fatal("scenario registry is nil after Bootstrap")
	}
	refs := map[string]string{
		"staff-welcome":     "staff-welcome",
		"staff-check-in":    "bob-staff-check-in",
		"project-kickoff":   "project-kickoff",
		"coordination-plan": "coordination-plan",
	}
	ids := make(map[string]int32, len(refs))
	for name, ref := range refs {
		ids[name] = getScenarioEntityID(t, reg, scenario.RefTypeThread, ref)
	}
	return ids
}

func assertScenarioVisibility(t *testing.T, err error, visible bool) {
	t.Helper()
	if visible && err != nil {
		t.Fatalf("authorized resource lookup failed: %v", err)
	}
	if !visible && !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("unauthorized resource lookup error = %v, want sql.ErrNoRows", err)
	}
}

func TestScenarioServe_ThreadRefIdentityWithAndWithoutAttachments(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 3))
	img.Set(0, 0, color.RGBA{R: 0x20, G: 0x70, B: 0xb0, A: 0xff})
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatalf("encode scenario JPEG: %v", err)
	}
	jpegBytes := buf.Bytes()

	for _, withAttachment := range []bool{false, true} {
		name := "without-attachment"
		assetHeader := ""
		if withAttachment {
			name = "with-attachment"
			assetHeader = "Asset: assets/welcome.jpg\n"
		}
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			scenarioTxt := fmt.Sprintf(`-- scenario.meta --
Format: goa4web-scenario/v1
Name: ref-identity-test

-- 010-alice.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: alice-test
At: 2026-08-01T09:00:00Z

-- 011-enable-alice.event --
Op: user.enable
Actor: admin
User: alice
At: 2026-08-01T09:00:10Z

-- 012-alice-see.event --
Op: user.grant
User: alice
Section: privateforum
Item: topic
Action: see
At: 2026-08-01T09:00:20Z

-- 013-alice-view.event --
Op: user.grant
User: alice
Section: privateforum
Item: topic
Action: view
At: 2026-08-01T09:00:30Z

-- 014-alice-create.event --
Op: user.grant
User: alice
Section: privateforum
Item: topic
Action: create
At: 2026-08-01T09:00:40Z

-- 015-alice-image-upload.event --
Op: user.grant
User: alice
Section: images
Item: upload
Action: post
At: 2026-08-01T09:00:50Z

-- 020-bob.event --
Op: user.create
Ref: bob
Username: bob
Email: bob@example.test
Password: bob-test
At: 2026-08-01T09:01:00Z

-- 021-enable-bob.event --
Op: user.enable
Actor: admin
User: bob
At: 2026-08-01T09:01:10Z

-- 022-bob-see.event --
Op: user.grant
User: bob
Section: privateforum
Item: topic
Action: see
At: 2026-08-01T09:01:20Z

-- 023-bob-view.event --
Op: user.grant
User: bob
Section: privateforum
Item: topic
Action: view
At: 2026-08-01T09:01:30Z

-- 100-staff-room.event --
Op: private-forum.create
Ref: staff-room
Actor: alice
Participant: bob
Title: Staff Room
Description: Private discussion for Alice and Bob
At: 2026-08-01T09:10:00Z

-- 110-staff-welcome.event --
Op: forum.thread.create
Ref: staff-welcome
Actor: alice
Topic: staff-room
%sAt: 2026-08-01T09:15:00Z

Welcome to the staff room.
`, assetHeader)

			memFS := fstest.MapFS{
				"welcome/scenario.txtar": &fstest.MapFile{Data: []byte(scenarioTxt)},
			}
			if withAttachment {
				memFS["welcome/assets/welcome.jpg"] = &fstest.MapFile{Data: jpegBytes}
			}

			root, err := parseRoot([]string{"goa4web", "scenario", "serve", "welcome"})
			if err != nil {
				t.Fatalf("parseRoot: %v", err)
			}
			defer root.Close()
			parent, err := parseScenarioCmd(root, []string{"serve", "welcome"})
			if err != nil {
				t.Fatalf("parseScenarioCmd: %v", err)
			}
			serveCmd, err := parseScenarioServeCmd(parent, []string{"welcome"})
			if err != nil {
				t.Fatalf("parseScenarioServeCmd: %v", err)
			}
			serveCmd.fsys = memFS

			srv, dbConn, cleanup, err := serveCmd.Bootstrap(ctx)
			if err != nil {
				t.Fatalf("Bootstrap: %v", err)
			}
			defer cleanup()

			if serveCmd.Registry == nil {
				t.Fatal("scenario registry is nil after Bootstrap")
			}
			topicID := getScenarioEntityID(t, serveCmd.Registry, scenario.RefTypeForum, "staff-room")
			threadID := getScenarioEntityID(t, serveCmd.Registry, scenario.RefTypeThread, "staff-welcome")
			if topicID <= 0 || threadID <= 0 {
				t.Fatalf("expected positive bound IDs, got topicID=%d threadID=%d", topicID, threadID)
			}

			var storedText string
			if err := dbConn.QueryRowContext(ctx, "SELECT text FROM comments WHERE forumthread_id = ?", threadID).Scan(&storedText); err != nil {
				t.Fatalf("query thread opening comment: %v", err)
			}
			if !withAttachment {
				if storedText != "Welcome to the staff room." {
					t.Fatalf("stored text = %q, want %q", storedText, "Welcome to the staff room.")
				}
				if strings.Contains(storedText, "[img=image:") {
					t.Fatalf("unexpected image markup in text without attachment: %q", storedText)
				}
				return
			}

			prepared, err := common.PrepareImageUpload(jpegBytes, "welcome.jpg")
			if err != nil {
				t.Fatalf("PrepareImageUpload: %v", err)
			}
			imageName := prepared.ID + prepared.Ext
			wantTag := "[img=image:" + imageName + "]"
			if !strings.Contains(storedText, "Welcome to the staff room.") || !strings.Contains(storedText, wantTag) {
				t.Fatalf("stored text = %q, want welcome text and %q", storedText, wantTag)
			}

			uploadPath := filepath.Join(srv.Config.ImageUploadDir, prepared.ID[:2], prepared.ID[2:4], imageName)
			storedOriginal, err := os.ReadFile(uploadPath)
			if err != nil {
				t.Fatalf("read uploaded original %s: %v", uploadPath, err)
			}
			if !bytes.Equal(storedOriginal, jpegBytes) {
				t.Fatalf("uploaded bytes len = %d, want %d", len(storedOriginal), len(jpegBytes))
			}

			thumbDir := filepath.Join(srv.Config.ImageCacheDir, prepared.ID[:2], prepared.ID[2:4])
			entries, err := os.ReadDir(thumbDir)
			if err != nil || len(entries) == 0 {
				t.Fatalf("expected generated thumbnail in %s (entries=%d err=%v)", thumbDir, len(entries), err)
			}

			var cacheCount int
			if err := dbConn.QueryRowContext(ctx, "SELECT COUNT(*) FROM image_cache_entries").Scan(&cacheCount); err != nil || cacheCount == 0 {
				t.Fatalf("expected recorded image_cache_entries row, count=%d err=%v", cacheCount, err)
			}

			if rendered := serveCmd.CoreData.MapImageURL("img", "image:"+imageName); !strings.Contains(rendered, "/images/image/"+imageName) {
				t.Errorf("MapImageURL(%q) = %q, want /images/image/%s", imageName, rendered, imageName)
			}
		})
	}
}

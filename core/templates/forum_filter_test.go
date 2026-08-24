package templates_test

import (
	"bytes"
	"database/sql"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/db"
)

func TestTableTopicsFilterMarkers(t *testing.T) {
	r := httptest.NewRequest("GET", "/private", nil)
	cd := &common.CoreData{}
	compiled := templates.GetCompiledSiteTemplates(cd.Funcs(r), templates.WithSilence(true))

	topic := &common.PrivateTopic{
		ListPrivateTopicsByUserIDRow: &db.ListPrivateTopicsByUserIDRow{
			Idforumtopic:       1,
			Title:              sql.NullString{String: "Private chat with alice", Valid: true},
			Description:        sql.NullString{String: "A description", Valid: true},
			Lastaddition:       sql.NullTime{Time: time.Now(), Valid: true},
			Lastposterusername: sql.NullString{String: "testposter", Valid: true},
			Lastposter:         2,
			Threads:            sql.NullInt32{Int32: 5, Valid: true},
			Comments:           sql.NullInt32{Int32: 10, Valid: true},
		},
		DisplayTitle:       "alice",
		ParticipantsString: "alice, bob",
		Participants:       []string{"alice", "bob"},
		TotalParticipants:  2,
	}

	data := map[string]any{
		"Topics":    []*common.PrivateTopic{topic},
		"BasePath":  "/private",
		"IsPrivate": true,
		"IsAdmin":   false,
	}

	var buf bytes.Buffer
	if err := compiled.ExecuteTemplate(&buf, "tableTopics", data); err != nil {
		t.Fatalf("failed to execute tableTopics: %v", err)
	}

	out := buf.String()

	if !strings.Contains(out, "topic-name-filter") {
		t.Errorf("tableTopics rendered output missing 'topic-name-filter' class: %s", out)
	}
	if !strings.Contains(out, "poster-name") {
		t.Errorf("tableTopics rendered output missing 'poster-name' class: %s", out)
	}
	if !strings.Contains(out, "participant") {
		t.Errorf("tableTopics rendered output missing 'participant' class: %s", out)
	}
	if !strings.Contains(out, "placeholder=\"Filter by label, poster, participant, or topic...\"") {
		t.Errorf("tableTopics rendered output missing expected filter placeholder: %s", out)
	}
}

func TestTopicThreadsFilterMarkers(t *testing.T) {
	r := httptest.NewRequest("GET", "/forum/topic/1", nil)
	cd := &common.CoreData{}
	compiled := templates.GetCompiledSiteTemplates(cd.Funcs(r), templates.WithSilence(true))

	type threadWithLabels struct {
		*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow
		Labels []templates.TopicLabel
	}

	thread := &threadWithLabels{
		GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow: &db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow{
			Idforumthread:          1,
			ForumtopicIdforumtopic: 1,
			Firstpostwritten:       sql.NullTime{Time: time.Now(), Valid: true},
			Firstpostusername:      sql.NullString{String: "firstposter", Valid: true},
			Firstpostuserid:        sql.NullInt32{Int32: 1, Valid: true},
			Firstposttext:          sql.NullString{String: "thread content", Valid: true},
			Lastaddition:           sql.NullTime{Time: time.Now(), Valid: true},
			Lastposterusername:     sql.NullString{String: "lastposter", Valid: true},
			Lastposterid:           sql.NullInt32{Int32: 2, Valid: true},
			Comments:               sql.NullInt32{Int32: 3, Valid: true},
		},
	}

	data := map[string]any{
		"Threads":  []*threadWithLabels{thread},
		"BasePath": "/forum",
	}

	var buf bytes.Buffer
	if err := compiled.ExecuteTemplate(&buf, "topicThreads", data); err != nil {
		t.Fatalf("failed to execute topicThreads: %v", err)
	}

	out := buf.String()

	if !strings.Contains(out, "poster-name") {
		t.Errorf("topicThreads rendered output missing 'poster-name' class: %s", out)
	}
	if !strings.Contains(out, "placeholder=\"Filter by label, poster, participant, or topic...\"") {
		t.Errorf("topicThreads rendered output missing expected filter placeholder: %s", out)
	}
}

func TestUnreadFilterMarkers(t *testing.T) {
	r := httptest.NewRequest("GET", "/private/unread", nil)
	cd := &common.CoreData{}
	compiled := templates.GetCompiledSiteTemplates(cd.Funcs(r), templates.WithSilence(true))

	type decorThread struct {
		*db.ListUnreadPrivateThreadsForUserRow
		DisplayTitle string
	}

	thread := &decorThread{
		ListUnreadPrivateThreadsForUserRow: &db.ListUnreadPrivateThreadsForUserRow{
			Idforumthread:      1,
			TopicID:            1,
			Firstpostwritten:   sql.NullTime{Time: time.Now(), Valid: true},
			Firstpostusername:  sql.NullString{String: "firstposter", Valid: true},
			Firstpostuserid:    sql.NullInt32{Int32: 1, Valid: true},
			Firstposttext:      sql.NullString{String: "unread content", Valid: true},
			Lastaddition:       sql.NullTime{Time: time.Now(), Valid: true},
			Lastposterusername: sql.NullString{String: "lastposter", Valid: true},
			Lastposterid:       sql.NullInt32{Int32: 2, Valid: true},
			Comments:           sql.NullInt32{Int32: 5, Valid: true},
		},
		DisplayTitle: "Private Topic 1",
	}

	data := struct {
		Threads      []*decorThread
		CurrentError string
		Page         int
		PrevPage     int
		NextPage     int
		HasNextPage  bool
		CD           *common.CoreData
		TopicTitle   string
	}{
		Threads: []*decorThread{thread},
		CD:      cd,
	}

	var buf bytes.Buffer
	if err := compiled.ExecuteTemplate(&buf, "domains/privateforum/unread.gohtml", data); err != nil {
		t.Fatalf("failed to execute unread.gohtml: %v", err)
	}

	out := buf.String()

	if !strings.Contains(out, "topic-name-filter") {
		t.Errorf("unread.gohtml rendered output missing 'topic-name-filter' class: %s", out)
	}
	if !strings.Contains(out, "poster-name") {
		t.Errorf("unread.gohtml rendered output missing 'poster-name' class: %s", out)
	}
	if !strings.Contains(out, "placeholder=\"Filter by label, poster, participant, or topic...\"") {
		t.Errorf("unread.gohtml rendered output missing expected filter placeholder: %s", out)
	}
}

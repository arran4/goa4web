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
			Title:              sql.NullString{String: "Secret Discussion", Valid: true},
			Description:        sql.NullString{String: "A secret topic description", Valid: true},
			Lastaddition:       sql.NullTime{Time: time.Now(), Valid: true},
			Lastposterusername: sql.NullString{String: "charlie_poster", Valid: true},
			Lastposter:         2,
			Threads:            sql.NullInt32{Int32: 5, Valid: true},
			Comments:           sql.NullInt32{Int32: 10, Valid: true},
		},
		DisplayTitle:       "Secret Discussion",
		ParticipantsString: "alice_user, bob_user",
		Participants:       []string{"alice_user", "bob_user"},
		TotalParticipants:  2,
		Labels: []templates.TopicLabel{
			{Name: "important_label", Type: "public"},
		},
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

	// Assert semantic markers wrap actual values
	if !strings.Contains(out, `class="topic-name-filter"`) || !strings.Contains(out, "Secret Discussion") {
		t.Errorf("tableTopics missing topic-name-filter wrapping topic title: %s", out)
	}
	if !strings.Contains(out, `class="poster-name">charlie_poster`) {
		t.Errorf("tableTopics missing poster-name wrapping poster username: %s", out)
	}
	if !strings.Contains(out, `<a class="participant">alice_user</a>`) || !strings.Contains(out, `<a class="participant">bob_user</a>`) {
		t.Errorf("tableTopics missing participant markers wrapping participants: %s", out)
	}
	if !strings.Contains(out, "important_label") || !strings.Contains(out, "class=\"label") {
		t.Errorf("tableTopics missing label markers: %s", out)
	}
	if !strings.Contains(out, `placeholder="Filter by label, poster, participant, or topic..."`) {
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
			Firstpostusername:      sql.NullString{String: "first_author", Valid: true},
			Firstpostuserid:        sql.NullInt32{Int32: 1, Valid: true},
			Firstposttext:          sql.NullString{String: "thread content", Valid: true},
			Lastaddition:           sql.NullTime{Time: time.Now(), Valid: true},
			Lastposterusername:     sql.NullString{String: "last_replier", Valid: true},
			Lastposterid:           sql.NullInt32{Int32: 2, Valid: true},
			Comments:               sql.NullInt32{Int32: 3, Valid: true},
		},
		Labels: []templates.TopicLabel{
			{Name: "question_label", Type: "public"},
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

	// Assert semantic markers wrap actual values
	if !strings.Contains(out, `class="poster-name first"`) || !strings.Contains(out, `first_author`) {
		t.Errorf("topicThreads missing poster-name wrapping first poster: %s", out)
	}
	if !strings.Contains(out, `class="poster-name last"`) || !strings.Contains(out, `last_replier`) {
		t.Errorf("topicThreads missing poster-name wrapping last poster: %s", out)
	}
	if !strings.Contains(out, "question_label") || !strings.Contains(out, "class=\"label") {
		t.Errorf("topicThreads missing label markers: %s", out)
	}
	if !strings.Contains(out, `placeholder="Filter by label or poster..."`) {
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
		Participants []string
	}

	thread := &decorThread{
		ListUnreadPrivateThreadsForUserRow: &db.ListUnreadPrivateThreadsForUserRow{
			Idforumthread:      1,
			TopicID:            10,
			Firstpostwritten:   sql.NullTime{Time: time.Now(), Valid: true},
			Firstpostusername:  sql.NullString{String: "unread_starter", Valid: true},
			Firstpostuserid:    sql.NullInt32{Int32: 1, Valid: true},
			Firstposttext:      sql.NullString{String: "unread content", Valid: true},
			Lastaddition:       sql.NullTime{Time: time.Now(), Valid: true},
			Lastposterusername: sql.NullString{String: "unread_last_replier", Valid: true},
			Lastposterid:       sql.NullInt32{Int32: 2, Valid: true},
			Comments:           sql.NullInt32{Int32: 5, Valid: true},
		},
		DisplayTitle: "Private Conversation A",
		Participants: []string{"participant_one", "participant_two"},
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

	// Assert semantic markers wrap actual values
	if !strings.Contains(out, `<strong class="topic-name-filter">Private Conversation A</strong>`) {
		t.Errorf("unread.gohtml missing topic-name-filter wrapping topic title: %s", out)
	}
	if !strings.Contains(out, `class="poster-name first"`) || !strings.Contains(out, `unread_starter`) {
		t.Errorf("unread.gohtml missing poster-name wrapping first poster: %s", out)
	}
	if !strings.Contains(out, `class="poster-name last"`) || !strings.Contains(out, `unread_last_replier`) {
		t.Errorf("unread.gohtml missing poster-name wrapping last poster: %s", out)
	}
	if !strings.Contains(out, `<a class="participant">participant_one</a>`) || !strings.Contains(out, `<a class="participant">participant_two</a>`) {
		t.Errorf("unread.gohtml missing participant markers wrapping participants: %s", out)
	}
	if !strings.Contains(out, `placeholder="Filter by poster, participant, or topic..."`) {
		t.Errorf("unread.gohtml rendered output missing expected filter placeholder: %s", out)
	}
}

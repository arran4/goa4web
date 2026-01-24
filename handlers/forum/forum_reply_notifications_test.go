package forum

import (
	"github.com/arran4/goa4web/handlers/forumcommon"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/mail"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/lazy"
	"github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/arran4/goa4web/workers/emailqueue"
	"github.com/arran4/goa4web/workers/postcountworker"
)

type captureDLQ struct {
	lastError string
}

func (c *captureDLQ) Record(ctx context.Context, message string) error {
	c.lastError = message
	return nil
}

type MockEmailProvider struct {
	SentMessages [][]byte
	Recipients   []mail.Address
}

func (m *MockEmailProvider) Send(ctx context.Context, to mail.Address, rawEmailMessage []byte) error {
	m.SentMessages = append(m.SentMessages, rawEmailMessage)
	m.Recipients = append(m.Recipients, to)
	return nil
}

func (m *MockEmailProvider) TestConfig(ctx context.Context) error { return nil }

// TestForumReply verifies that forum reply use correct templates with real event data and exact string matching.
func TestForumReply(t *testing.T) {
	replierUID := int32(1)
	subscriberUID := int32(2)
	missingEmailUID := int32(3)
	adminUID := int32(99)
	topicID := int32(5)
	threadID := int32(42)

	qs := testhelpers.NewQuerierStub()
	qs.GetPermissionsByUserIDFn = func(idusers int32) ([]*db.GetPermissionsByUserIDRow, error) {
		return []*db.GetPermissionsByUserIDRow{}, nil
	}
	qs.SystemGetUserByIDFn = func(ctx context.Context, idusers int32) (*db.SystemGetUserByIDRow, error) {
		switch idusers {
		case replierUID:
			return &db.SystemGetUserByIDRow{
				Idusers:  replierUID,
				Username: sql.NullString{String: "replier", Valid: true},
				Email:    sql.NullString{String: "replier@example.com", Valid: true},
			}, nil
		case subscriberUID:
			return &db.SystemGetUserByIDRow{
				Idusers:  subscriberUID,
				Username: sql.NullString{String: "subscriber", Valid: true},
				Email:    sql.NullString{String: "subscriber@example.com", Valid: true},
			}, nil
		case missingEmailUID:
			return &db.SystemGetUserByIDRow{
				Idusers:  missingEmailUID,
				Username: sql.NullString{String: "missing-email", Valid: true},
			}, nil
		case adminUID:
			return &db.SystemGetUserByIDRow{
				Idusers:  adminUID,
				Username: sql.NullString{String: "adminuser", Valid: true},
				Email:    sql.NullString{String: "admin@example.com", Valid: true},
			}, nil
		}
		return nil, sql.ErrNoRows
	}
	qs.SystemGetUserByEmailFn = func(ctx context.Context, email string) (*db.SystemGetUserByEmailRow, error) {
		if email == "admin@example.com" {
			return &db.SystemGetUserByEmailRow{Idusers: adminUID}, nil
		}
		return nil, sql.ErrNoRows
	}
	qs.CreateCommentInSectionForCommenterFn = func(ctx context.Context, arg db.CreateCommentInSectionForCommenterParams) (int64, error) {
		return 999, nil
	}
	qs.GetCommentByIdForUserRow = &db.GetCommentByIdForUserRow{
		Idcomments: 999,
	}
	qs.GetThreadBySectionThreadIDForReplierFn = func(ctx context.Context, arg db.GetThreadBySectionThreadIDForReplierParams) (*db.Forumthread, error) {
		return &db.Forumthread{
			Idforumthread:          threadID,
			ForumtopicIdforumtopic: topicID,
		}, nil
	}
	qs.GetThreadLastPosterAndPermsFn = func(ctx context.Context, arg db.GetThreadLastPosterAndPermsParams) (*db.GetThreadLastPosterAndPermsRow, error) {
		return &db.GetThreadLastPosterAndPermsRow{
			Idforumthread:          threadID,
			ForumtopicIdforumtopic: topicID,
			Lastposterusername:     sql.NullString{String: "replier", Valid: true},
			Comments:               sql.NullInt32{Int32: 1, Valid: true},
		}, nil
	}
	qs.GetForumTopicByIdForUserFn = func(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
		return &db.GetForumTopicByIdForUserRow{
			Idforumtopic: topicID,
			Title:        sql.NullString{String: "Test Topic", Valid: true},
		}, nil
	}
	qs.ListSubscribersForPatternsReturn = map[string][]int32{
		fmt.Sprintf("reply:/forum/topic/%d/thread/%d/*", topicID, threadID): {subscriberUID, missingEmailUID},
	}
	qs.GetPreferenceForListerReturn = map[int32]*db.Preference{
		replierUID:    {AutoSubscribeReplies: true},
		subscriberUID: {AutoSubscribeReplies: true},
	}
	qs.AdminListAdministratorEmailsReturns = []string{"admin@example.com"}
	qs.SystemGetLastNotificationForRecipientByMessageErr = sql.ErrNoRows
	qs.UpsertContentReadMarkerFn = func(ctx context.Context, arg db.UpsertContentReadMarkerParams) error { return nil }
	qs.ClearUnreadContentPrivateLabelExceptUserFn = func(ctx context.Context, arg db.ClearUnreadContentPrivateLabelExceptUserParams) error { return nil }
	qs.AdminDeletePendingEmailFn = func(ctx context.Context, id int32) error { return nil }
	qs.SystemMarkPendingEmailSentFn = func(ctx context.Context, id int32) error { return nil }

	qs.SystemListPendingEmailsFn = func(ctx context.Context, arg db.SystemListPendingEmailsParams) ([]*db.SystemListPendingEmailsRow, error) {
		var rows []*db.SystemListPendingEmailsRow
		if len(qs.InsertPendingEmailCalls) > 0 {
			call := qs.InsertPendingEmailCalls[0]
			rows = append(rows, &db.SystemListPendingEmailsRow{
				ID:          1,
				ToUserID:    call.ToUserID,
				Body:        call.Body,
				DirectEmail: call.DirectEmail,
			})
			qs.InsertPendingEmailCalls = qs.InsertPendingEmailCalls[1:]
		}
		return rows, nil
	}

	bus := eventbus.NewBus()
	cfg := config.NewRuntimeConfig()
	cfg.NotificationsEnabled = true
	cfg.AdminNotify = true
	cfg.EmailEnabled = true
	cfg.EmailFrom = "noreply@example.com"
	cfg.HTTPHostname = "http://example.com"

	mockProvider := &MockEmailProvider{}
	n := notifications.New(
		notifications.WithQueries(qs),
		notifications.WithConfig(cfg),
		notifications.WithEmailProvider(mockProvider),
	)
	cdlq := &captureDLQ{}
	n.RegisterSync(bus, cdlq)

	store := sessions.NewCookieStore([]byte("test"))
	core.Store = store
	core.SessionName = "test"
	sess, _ := store.Get(httptest.NewRequest(http.MethodGet, "http://example.com", nil), core.SessionName)
	sess.Values["UID"] = replierUID

	task := replyTask
	evt := &eventbus.TaskEvent{
		Data:    map[string]any{},
		UserID:  replierUID,
		Path:    "/forum/topic/5/thread/42/reply",
		Task:    task,
		Outcome: eventbus.TaskOutcomeSuccess,
	}
	ctx := context.Background()
	cd := common.NewCoreData(ctx, qs, cfg, common.WithSession(sess), common.WithEvent(evt), common.WithUserRoles([]string{"member"}))
	cd.UserID = replierUID

	thread := &db.GetThreadLastPosterAndPermsRow{Idforumthread: threadID, ForumtopicIdforumtopic: topicID, Lastposterusername: sql.NullString{String: "replier", Valid: true}, Comments: sql.NullInt32{Int32: 1, Valid: true}}
	topic := &db.GetForumTopicByIdForUserRow{Idforumtopic: topicID, Title: sql.NullString{String: "Test Topic", Valid: true}}
	cd.SetCurrentThreadAndTopic(threadID, topicID)
	_, _ = cd.ForumThreadByID(threadID, lazy.Set(thread))
	_, _ = cd.ForumTopicByID(topicID, lazy.Set(topic))

	ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

	form := url.Values{"replytext": {"This is a test message with a link [a https://example.com] and enough words to trigger the truncation of twenty words limit plus more."}, "language": {"1"}}
	req := httptest.NewRequest(http.MethodPost, "http://example.com/forum/topic/5/thread/42/reply", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"topic": "5", "thread": "42"})

	rr := httptest.NewRecorder()
	task.Action(rr, req)

	// Trigger synchronous processing
	bus.Publish(*evt)

	if cdlq.lastError != "" {
		t.Errorf("sync process error: %s", cdlq.lastError)
	}

	notificationsByRecipient := make(map[int32][]string)
	for _, call := range qs.SystemCreateNotificationCalls {
		notificationsByRecipient[call.RecipientID] = append(notificationsByRecipient[call.RecipientID], call.Message.String)
	}

	findNotification := func(recipient int32, expected string) bool {
		for _, msg := range notificationsByRecipient[recipient] {
			if msg == expected {
				return true
			}
		}
		return false
	}

	t.Run("internal notifications", func(t *testing.T) {
		t.Helper()
		expectedSubscriberNotif := "New reply in \"Test Topic\" by replier\n"
		if !findNotification(subscriberUID, expectedSubscriberNotif) {
			t.Fatalf("expected subscriber notification %q, got %q", expectedSubscriberNotif, notificationsByRecipient[subscriberUID])
		}

		expectedAdminNotif := "User replier replied to a forum thread.\nThis is a test message with a link https://example.com and enough words to trigger the truncation of twenty words limit...\n"
		if !findNotification(adminUID, expectedAdminNotif) {
			t.Fatalf("expected admin notification %q, got %q", expectedAdminNotif, notificationsByRecipient[adminUID])
		}

		if !findNotification(missingEmailUID, expectedSubscriberNotif) {
			t.Fatalf("expected missing email subscriber to receive internal notification %q, got %q", expectedSubscriberNotif, notificationsByRecipient[missingEmailUID])
		}
	})

	t.Run("missing email notifications", func(t *testing.T) {
		t.Helper()
		expected := "missing email address"
		if !findNotification(missingEmailUID, expected) {
			t.Fatalf("expected missing email notification %q, got %q", expected, notificationsByRecipient[missingEmailUID])
		}
	})

	t.Run("emails", func(t *testing.T) {
		for emailqueue.ProcessPendingEmail(ctx, qs, mockProvider, cdlq, cfg) {
		}

		if len(mockProvider.SentMessages) != 2 {
			t.Fatalf("expected 2 emails sent, got %d", len(mockProvider.SentMessages))
		}

		var subscriberEmail, adminEmail *mail.Message
		for i, raw := range mockProvider.SentMessages {
			msg, err := mail.ReadMessage(strings.NewReader(string(raw)))
			if err != nil {
				t.Fatalf("parse email %d: %v", i, err)
			}
			to := mockProvider.Recipients[i].Address
			if to == "subscriber@example.com" {
				subscriberEmail = msg
			} else if to == "admin@example.com" {
				adminEmail = msg
			}
		}

		if subscriberEmail == nil {
			t.Fatal("subscriber email not found")
		}
		if subscriberEmail.Header.Get("Subject") != "[goa4web] New forum reply" {
			t.Errorf("subscriber email subject mismatch: %s", subscriberEmail.Header.Get("Subject"))
		}
		subBody := getEmailBody(t, subscriberEmail)
		expectedSubBody := "Hi replier,\nYour reply has been posted.\n\nView comment:\nhttp://example.com/forum/topic/5/thread/42#c999\n\nManage notifications: http://example.com/usr/subscriptions"
		if subBody != expectedSubBody {
			t.Errorf("subscriber email body mismatch: %q, want %q", subBody, expectedSubBody)
		}

		if adminEmail == nil {
			t.Fatal("admin email not found")
		}
		if adminEmail.Header.Get("Subject") != "[goa4web Admin] Forum reply posted" {
			t.Errorf("admin email subject mismatch: %s", adminEmail.Header.Get("Subject"))
		}
		adminBody := getEmailBody(t, adminEmail)
		expectedAdminBody := "User replier replied to a forum thread.\nThis is a test message with a link [a https://example.com] and enough words to trigger the truncation of twenty words limit plus more.\n\nView comment:\nhttp://example.com/forum/topic/5/thread/42#c999\n\nManage notifications: http://example.com/usr/subscriptions"
		if adminBody != expectedAdminBody {
			t.Errorf("admin email body mismatch: %q, want %q", adminBody, expectedAdminBody)
		}
	})

	t.Run("auto subscribe path", func(t *testing.T) {
		// TODO: add subscription-chaining checks (topic -> thread) once modeled in the reply task.
		actionName, path, err := task.AutoSubscribePath(*evt)
		if err != nil {
			t.Errorf("AutoSubscribePath error: %v", err)
		}
		if actionName != "Reply" {
			t.Errorf("expected action name Reply, got %q", actionName)
		}
		if path != "/forum/topic/5/thread/42" {
			t.Errorf("expected path /forum/topic/5/thread/42, got %q", path)
		}
	})
}

func getEmailBody(t *testing.T, msg *mail.Message) string {
	t.Helper()
	ct := msg.Header.Get("Content-Type")
	if strings.Contains(ct, "multipart/alternative") {
		_, params, err := mime.ParseMediaType(ct)
		if err != nil {
			t.Fatal(err)
		}
		mr := multipart.NewReader(msg.Body, params["boundary"])
		for {
			p, err := mr.NextPart()
			if err != nil {
				break
			}
			pct := p.Header.Get("Content-Type")
			if strings.Contains(pct, "text/plain") {
				buf := new(bytes.Buffer)
				if _, err := io.Copy(buf, p); err != nil {
					t.Fatal(err)
				}
				return buf.String()
			}
		}
	}

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, msg.Body); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestForumAutoSubscribeTasks(t *testing.T) {
	if _, ok := interface{}(replyTask).(notifications.AutoSubscribeProvider); !ok {
		t.Fatalf("ReplyTask should implement AutoSubscribeProvider so users get notified about thread replies")
	}
	if _, ok := interface{}(createThreadTask).(notifications.AutoSubscribeProvider); !ok {
		t.Fatalf("CreateThreadTask should implement AutoSubscribeProvider so thread authors follow their threads")
	}

	replyEvt := eventbus.TaskEvent{
		Data: map[string]any{
			postcountworker.EventKey: postcountworker.UpdateEventData{
				ThreadID:  77,
				TopicID:   88,
				CommentID: 999,
			},
		},
		Path: "/forum/topic/1/thread/2/reply",
	}
	actionName, path, err := replyTask.AutoSubscribePath(replyEvt)
	if err != nil {
		t.Fatalf("reply AutoSubscribePath error: %v", err)
	}
	if actionName != string(forumcommon.TaskReply) {
		t.Fatalf("expected action name %q, got %q", forumcommon.TaskReply, actionName)
	}
	if path != "/forum/topic/88/thread/77" {
		t.Fatalf("expected reply auto-subscribe path /forum/topic/88/thread/77, got %q", path)
	}

	createThreadEvt := eventbus.TaskEvent{
		Data: map[string]any{
			postcountworker.EventKey: postcountworker.UpdateEventData{
				ThreadID:  55,
				TopicID:   44,
				CommentID: 777,
			},
		},
		Path: "/forum/topic/9/thread/10/new",
	}
	actionName, path, err = createThreadTask.AutoSubscribePath(createThreadEvt)
	if err != nil {
		t.Fatalf("create thread AutoSubscribePath error: %v", err)
	}
	if actionName != string(forumcommon.TaskCreateThread) {
		t.Fatalf("expected action name %q, got %q", forumcommon.TaskCreateThread, actionName)
	}
	if path != "/forum/topic/44/thread/55" {
		t.Fatalf("expected create thread auto-subscribe path /forum/topic/44/thread/55, got %q", path)
	}
}

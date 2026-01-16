package news

import (
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
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/arran4/goa4web/workers/emailqueue"
)

type newsQueries struct {
	*db.QuerierStub
	forumTopic *db.Forumtopic
	newsPost   *db.GetNewsPostByIdWithWriterIdAndThreadCommentCountRow
}

func (q *newsQueries) SystemGetForumTopicByTitle(ctx context.Context, title sql.NullString) (*db.Forumtopic, error) {
	return q.forumTopic, nil
}

func (q *newsQueries) GetNewsPostByIdWithWriterIdAndThreadCommentCount(ctx context.Context, arg db.GetNewsPostByIdWithWriterIdAndThreadCommentCountParams) (*db.GetNewsPostByIdWithWriterIdAndThreadCommentCountRow, error) {
	return q.newsPost, nil
}

type captureDLQ struct {
	lastError string
}

func (c *captureDLQ) Record(ctx context.Context, message string) error {
	c.lastError = message
	return nil
}

type mockEmailProvider struct {
	sentMessages [][]byte
	recipients   []mail.Address
}

func (m *mockEmailProvider) Send(ctx context.Context, to mail.Address, rawEmailMessage []byte) error {
	m.sentMessages = append(m.sentMessages, rawEmailMessage)
	m.recipients = append(m.recipients, to)
	return nil
}

func (m *mockEmailProvider) TestConfig(ctx context.Context) error { return nil }

func TestNewsReply(t *testing.T) {
	replierUID := int32(1)
	subscriberUID := int32(2)
	adminUID := int32(99)
	postID := int32(1)
	threadID := int32(42)
	topicID := int32(7)
	fixedTime := time.Date(2024, 2, 3, 4, 5, 0, 0, time.UTC)

	qs := &newsQueries{
		QuerierStub: testhelpers.NewQuerierStub(testhelpers.WithGrantResult(true)),
		forumTopic:  &db.Forumtopic{Idforumtopic: topicID, Title: sql.NullString{String: NewsTopicName, Valid: true}},
		newsPost: &db.GetNewsPostByIdWithWriterIdAndThreadCommentCountRow{
			Idsitenews:    postID,
			ForumthreadID: threadID,
			LanguageID:    sql.NullInt32{Int32: 1, Valid: true},
		},
	}
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
	qs.GetThreadLastPosterAndPermsFn = func(ctx context.Context, arg db.GetThreadLastPosterAndPermsParams) (*db.GetThreadLastPosterAndPermsRow, error) {
		return &db.GetThreadLastPosterAndPermsRow{
			Idforumthread:          threadID,
			ForumtopicIdforumtopic: topicID,
			Lastposterusername:     sql.NullString{String: "replier", Valid: true},
			Comments:               sql.NullInt32{Int32: 2, Valid: true},
		}, nil
	}
	qs.GetForumTopicByIdFn = func(ctx context.Context, idforumtopic int32) (*db.Forumtopic, error) {
		return &db.Forumtopic{Idforumtopic: topicID, Title: sql.NullString{String: NewsTopicName, Valid: true}}, nil
	}
	qs.ListSubscribersForPatternsReturn = map[string][]int32{
		fmt.Sprintf("reply:/news/news/%d", postID): {subscriberUID},
	}
	qs.GetPreferenceForListerReturn = map[int32]*db.Preference{
		replierUID: {AutoSubscribeReplies: true},
	}
	qs.AdminListAdministratorEmailsReturns = []string{"admin@example.com"}
	qs.UpsertContentReadMarkerFn = func(ctx context.Context, arg db.UpsertContentReadMarkerParams) error { return nil }
	qs.ClearUnreadContentPrivateLabelExceptUserFn = func(ctx context.Context, arg db.ClearUnreadContentPrivateLabelExceptUserParams) error { return nil }
	qs.AdminDeletePendingEmailFn = func(ctx context.Context, id int32) error { return nil }
	qs.SystemMarkPendingEmailSentFn = func(ctx context.Context, id int32) error { return nil }
	qs.CreateCommentInSectionForCommenterFn = func(ctx context.Context, arg db.CreateCommentInSectionForCommenterParams) (int64, error) {
		return 999, nil
	}

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

	mockProvider := &mockEmailProvider{}
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

	evt := &eventbus.TaskEvent{
		Data:    map[string]any{"Time": fixedTime},
		UserID:  replierUID,
		Path:    "/news/news/1",
		Task:    replyTask,
		Outcome: eventbus.TaskOutcomeSuccess,
	}
	ctx := context.Background()
	cd := common.NewCoreData(ctx, qs, cfg, common.WithSession(sess), common.WithEvent(evt), common.WithUserRoles([]string{"member"}))
	cd.UserID = replierUID

	ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

	form := url.Values{"replytext": {"Hello News"}, "language": {"1"}}
	req := httptest.NewRequest(http.MethodPost, "http://example.com/news/news/1", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"news": "1"})

	rr := httptest.NewRecorder()
	replyTask.Action(rr, req)

	bus.Publish(*evt)

	if cdlq.lastError != "" {
		t.Errorf("sync process error: %s", cdlq.lastError)
	}

	var subscriberNotif, adminNotif string
	for _, call := range qs.SystemCreateNotificationCalls {
		if call.RecipientID == subscriberUID {
			subscriberNotif = call.Message.String
		}
		if call.RecipientID == adminUID {
			adminNotif = call.Message.String
		}
	}

	expectedSubscriberNotif := fmt.Sprintf("New reply in %q by replier\n", NewsTopicName)
	if subscriberNotif != expectedSubscriberNotif {
		t.Errorf("expected subscriber notif %q, got %q", expectedSubscriberNotif, subscriberNotif)
	}

	expectedAdminNotif := "User replier replied to a news post\n\n"
	if adminNotif != expectedAdminNotif {
		t.Errorf("expected admin notif %q, got %q", expectedAdminNotif, adminNotif)
	}

	for emailqueue.ProcessPendingEmail(ctx, qs, mockProvider, cdlq, cfg) {
	}

	if len(mockProvider.sentMessages) != 2 {
		t.Fatalf("expected 2 emails sent, got %d", len(mockProvider.sentMessages))
	}

	var subscriberEmail, adminEmail *mail.Message
	for i, raw := range mockProvider.sentMessages {
		msg, err := mail.ReadMessage(strings.NewReader(string(raw)))
		if err != nil {
			t.Fatalf("parse email %d: %v", i, err)
		}
		to := mockProvider.recipients[i].Address
		if to == "subscriber@example.com" {
			subscriberEmail = msg
		} else if to == "admin@example.com" {
			adminEmail = msg
		}
	}

	if subscriberEmail == nil {
		t.Fatal("subscriber email not found")
	}
	if subscriberEmail.Header.Get("Subject") != "[goa4web] New reply in "+NewsTopicName {
		t.Errorf("subscriber email subject mismatch: %s", subscriberEmail.Header.Get("Subject"))
	}
	subBody := getEmailBody(t, subscriberEmail)
	expectedSubBody := fmt.Sprintf(
		"Hi replier,\n\nA new reply was posted in %q (thread #%d) on %s.\nThere are now %d comments in the discussion.\n\nView it here:\nhttp://example.com/news/news/1\n\n\nManage notifications: http://example.com/usr/subscriptions",
		NewsTopicName,
		threadID,
		fixedTime.Format(consts.DisplayDateTimeFormat),
		2,
	)
	if subBody != expectedSubBody {
		t.Errorf("subscriber email body mismatch: %q, want %q", subBody, expectedSubBody)
	}

	if adminEmail == nil {
		t.Fatal("admin email not found")
	}
	if adminEmail.Header.Get("Subject") != "[goa4web Admin] News comment posted" {
		t.Errorf("admin email subject mismatch: %s", adminEmail.Header.Get("Subject"))
	}
	adminBody := getEmailBody(t, adminEmail)
	expectedAdminBody := "User replier replied to a news post.\n\nView post:\nhttp://example.com/news/news/1\n\nManage notifications: http://example.com/usr/subscriptions"
	if adminBody != expectedAdminBody {
		t.Errorf("admin email body mismatch: %q, want %q", adminBody, expectedAdminBody)
	}

	actionName, path, err := replyTask.AutoSubscribePath(*evt)
	if err != nil {
		t.Errorf("AutoSubscribePath error: %v", err)
	}
	if actionName != "Reply" {
		t.Errorf("expected action name Reply, got %q", actionName)
	}
	if path != fmt.Sprintf("/forum/topic/%d/thread/%d", topicID, threadID) {
		t.Errorf("expected path /forum/topic/%d/thread/%d, got %q", topicID, threadID, path)
	}
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

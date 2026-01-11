package imagebbs

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
	"github.com/arran4/goa4web/workers/emailqueue"
)

type imageBbsQueries struct {
	*db.QuerierStub
	forumTopic *db.Forumtopic
	imagePost  *db.GetImagePostByIDForListerRow
}

func (q *imageBbsQueries) SystemGetForumTopicByTitle(ctx context.Context, title sql.NullString) (*db.Forumtopic, error) {
	return q.forumTopic, nil
}

func (q *imageBbsQueries) GetImagePostByIDForLister(ctx context.Context, arg db.GetImagePostByIDForListerParams) (*db.GetImagePostByIDForListerRow, error) {
	return q.imagePost, nil
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

func TestImageBbsReply(t *testing.T) {
	replierUID := int32(1)
	subscriberUID := int32(2)
	boardID := int32(9)
	threadID := int32(101)
	topicID := int32(202)
	fixedTime := time.Date(2024, 5, 6, 7, 8, 0, 0, time.UTC)

	qs := &imageBbsQueries{
		QuerierStub: &db.QuerierStub{
			SystemCheckGrantReturns: 1,
			GetPermissionsByUserIDFn: func(idusers int32) ([]*db.GetPermissionsByUserIDRow, error) {
				return []*db.GetPermissionsByUserIDRow{}, nil
			},
			SystemGetUserByIDFn: func(ctx context.Context, idusers int32) (*db.SystemGetUserByIDRow, error) {
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
				}
				return nil, sql.ErrNoRows
			},
			GetThreadLastPosterAndPermsFn: func(ctx context.Context, arg db.GetThreadLastPosterAndPermsParams) (*db.GetThreadLastPosterAndPermsRow, error) {
				return &db.GetThreadLastPosterAndPermsRow{
					Idforumthread:          threadID,
					ForumtopicIdforumtopic: topicID,
					Lastposterusername:     sql.NullString{String: "replier", Valid: true},
					Comments:               sql.NullInt32{Int32: 6, Valid: true},
				}, nil
			},
			GetForumTopicByIdFn: func(ctx context.Context, idforumtopic int32) (*db.Forumtopic, error) {
				return &db.Forumtopic{Idforumtopic: topicID, Title: sql.NullString{String: ImageBBSTopicName, Valid: true}}, nil
			},
			ListSubscribersForPatternsReturn: map[string][]int32{
				fmt.Sprintf("reply:/imagebbss/imagebbs/%d/comments", boardID): {subscriberUID},
			},
			GetPreferenceForListerReturn: map[int32]*db.Preference{
				replierUID: {AutoSubscribeReplies: true},
			},
			UpsertContentReadMarkerFn:                  func(ctx context.Context, arg db.UpsertContentReadMarkerParams) error { return nil },
			ClearUnreadContentPrivateLabelExceptUserFn: func(ctx context.Context, arg db.ClearUnreadContentPrivateLabelExceptUserParams) error { return nil },
			AdminDeletePendingEmailFn:                  func(ctx context.Context, id int32) error { return nil },
			SystemMarkPendingEmailSentFn:               func(ctx context.Context, id int32) error { return nil },
			CreateCommentInSectionForCommenterFn: func(ctx context.Context, arg db.CreateCommentInSectionForCommenterParams) (int64, error) {
				return 888, nil
			},
		},
		forumTopic: &db.Forumtopic{Idforumtopic: topicID, Title: sql.NullString{String: ImageBBSTopicName, Valid: true}},
		imagePost: &db.GetImagePostByIDForListerRow{
			Idimagepost:   boardID,
			ForumthreadID: threadID,
		},
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
		Path:    "/imagebbss/imagebbs/9/comments",
		Task:    replyTask,
		Outcome: eventbus.TaskOutcomeSuccess,
	}
	ctx := context.Background()
	cd := common.NewCoreData(ctx, qs, cfg, common.WithSession(sess), common.WithEvent(evt), common.WithUserRoles([]string{"member"}))
	cd.UserID = replierUID

	ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

	form := url.Values{"replytext": {"Hello ImageBBS"}, "language": {"1"}}
	req := httptest.NewRequest(http.MethodPost, "http://example.com/imagebbss/imagebbs/9/comments", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"board": "9"})

	rr := httptest.NewRecorder()
	replyTask.Action(rr, req)

	bus.Publish(*evt)

	if cdlq.lastError != "" {
		t.Errorf("sync process error: %s", cdlq.lastError)
	}

	var subscriberNotif string
	for _, call := range qs.SystemCreateNotificationCalls {
		if call.RecipientID == subscriberUID {
			subscriberNotif = call.Message.String
		}
	}

	expectedSubscriberNotif := fmt.Sprintf("New reply in %q by replier\n", ImageBBSTopicName)
	if subscriberNotif != expectedSubscriberNotif {
		t.Errorf("expected subscriber notif %q, got %q", expectedSubscriberNotif, subscriberNotif)
	}

	for emailqueue.ProcessPendingEmail(ctx, qs, mockProvider, cdlq, cfg) {
	}

	if len(mockProvider.sentMessages) != 1 {
		t.Fatalf("expected 1 email sent, got %d", len(mockProvider.sentMessages))
	}

	msg, err := mail.ReadMessage(strings.NewReader(string(mockProvider.sentMessages[0])))
	if err != nil {
		t.Fatalf("parse email: %v", err)
	}
	if msg.Header.Get("Subject") != "[goa4web] New reply in "+ImageBBSTopicName {
		t.Errorf("subscriber email subject mismatch: %s", msg.Header.Get("Subject"))
	}
	subBody := getEmailBody(t, msg)
	expectedSubBody := fmt.Sprintf(
		"Hi replier,\n\nA new reply was posted in %q (thread #%d) on %s.\nThere are now %d comments in the discussion.\n\nView it here:\n/imagebbss/imagebbs/9/comments\n\n\nManage notifications: http://example.com/usr/subscriptions",
		ImageBBSTopicName,
		threadID,
		fixedTime.Format(consts.DisplayDateTimeFormat),
		6,
	)
	if subBody != expectedSubBody {
		t.Errorf("subscriber email body mismatch: %q, want %q", subBody, expectedSubBody)
	}

	actionName, path, err := replyTask.AutoSubscribePath(*evt)
	if err != nil {
		t.Errorf("AutoSubscribePath error: %v", err)
	}
	if actionName != "Reply" {
		t.Errorf("expected action name Reply, got %q", actionName)
	}
	if path != "/imagebbss/imagebbs/9/comments" {
		t.Errorf("expected path /imagebbss/imagebbs/9/comments, got %q", path)
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

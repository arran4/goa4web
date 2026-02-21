package writings

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

type writingsQueries struct {
	*db.QuerierStub
	forumTopic *db.Forumtopic
}

func (q *writingsQueries) SystemGetForumTopicByTitle(ctx context.Context, title sql.NullString) (*db.Forumtopic, error) {
	return q.forumTopic, nil
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

func (m *mockEmailProvider) TestConfig(ctx context.Context) (string, error) { return "", nil }

func TestWritingReply_Notifications(t *testing.T) {
	t.Run("Happy Path - Reply and Notify", func(t *testing.T) {
		replierUID := int32(1)
		subscriberUID := int32(2)
		writingID := int32(3)
		threadID := int32(22)
		topicID := int32(33)
		fixedTime := time.Date(2024, 3, 4, 5, 6, 0, 0, time.UTC)

		qs := &writingsQueries{
			QuerierStub: testhelpers.NewQuerierStub(testhelpers.WithGrantResult(true)),
			forumTopic:  &db.Forumtopic{Idforumtopic: topicID, Title: sql.NullString{String: WritingTopicName, Valid: true}},
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
			}
			return nil, sql.ErrNoRows
		}
		qs.GetThreadLastPosterAndPermsFn = func(ctx context.Context, arg db.GetThreadLastPosterAndPermsParams) (*db.GetThreadLastPosterAndPermsRow, error) {
			return &db.GetThreadLastPosterAndPermsRow{
				Idforumthread:          threadID,
				ForumtopicIdforumtopic: topicID,
				Lastposterusername:     sql.NullString{String: "replier", Valid: true},
				Comments:               sql.NullInt32{Int32: 4, Valid: true},
			}, nil
		}
		qs.GetForumTopicByIdFn = func(ctx context.Context, idforumtopic int32) (*db.Forumtopic, error) {
			return &db.Forumtopic{Idforumtopic: topicID, Title: sql.NullString{String: WritingTopicName, Valid: true}}, nil
		}
		qs.GetWritingForListerByIDRow = &db.GetWritingForListerByIDRow{
			Idwriting:     writingID,
			ForumthreadID: threadID,
			LanguageID:    sql.NullInt32{Int32: 1, Valid: true},
			Title:         sql.NullString{String: "Test Writing", Valid: true},
		}
		qs.ListSubscribersForPatternsReturn = map[string][]int32{
			fmt.Sprintf("reply:/writings/article/%d", writingID): {subscriberUID},
		}
		qs.GetPreferenceForListerReturn = map[int32]*db.Preference{
			replierUID: {AutoSubscribeReplies: true},
		}
		qs.UpsertContentReadMarkerFn = func(ctx context.Context, arg db.UpsertContentReadMarkerParams) error { return nil }
		qs.ClearUnreadContentPrivateLabelExceptUserFn = func(ctx context.Context, arg db.ClearUnreadContentPrivateLabelExceptUserParams) error { return nil }
		qs.AdminDeletePendingEmailFn = func(ctx context.Context, id int32) error { return nil }
		qs.SystemMarkPendingEmailSentFn = func(ctx context.Context, id int32) error { return nil }
		qs.CreateCommentInSectionForCommenterFn = func(ctx context.Context, arg db.CreateCommentInSectionForCommenterParams) (int64, error) {
			return 555, nil
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
		cfg.BaseURL = "http://example.com"

		mockProvider := &mockEmailProvider{}
		n := notifications.New(notifications.WithSilence(true),
			notifications.WithQueries(qs),
			notifications.WithConfig(cfg),
			notifications.WithEmailProvider(mockProvider),
		)
		cdlq := &captureDLQ{}
		n.RegisterSync(bus, cdlq)

		store := sessions.NewCookieStore([]byte("test"))
		core.Store = store
		core.SessionName = "test"
		sess := testhelpers.Must(store.Get(httptest.NewRequest(http.MethodGet, "http://example.com", nil), core.SessionName))
		sess.Values["UID"] = replierUID

		evt := &eventbus.TaskEvent{
			Data:    map[string]any{"Time": fixedTime},
			UserID:  replierUID,
			Path:    "/writings/article/3",
			Task:    replyTask,
			Outcome: eventbus.TaskOutcomeSuccess,
		}
		ctx := context.Background()
		cd := common.NewCoreData(ctx, qs, cfg, common.WithSession(sess), common.WithEvent(evt), common.WithUserRoles([]string{"member"}))
		cd.UserID = replierUID
		cd.SetCurrentWriting(writingID)

		ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
		ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

		form := url.Values{"replytext": {"Hello Writing"}, "language": {"1"}}
		req := httptest.NewRequest(http.MethodPost, "http://example.com/writings/article/3", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = req.WithContext(ctx)
		req = mux.SetURLVars(req, map[string]string{"writing": "3"})

		rr := httptest.NewRecorder()
		replyTask.Action(rr, req)

		bus.Publish(*evt)

		t.Run("Event Bus Verification", func(t *testing.T) {
			if cdlq.lastError != "" {
				t.Errorf("sync process error: %s", cdlq.lastError)
			}
		})

		t.Run("Internal Notification Verification", func(t *testing.T) {
			var subscriberNotif string
			for _, call := range qs.SystemCreateNotificationCalls {
				if call.RecipientID == subscriberUID {
					subscriberNotif = call.Message.String
				}
			}

			expectedSubscriberNotif := fmt.Sprintf(" replied to %s: Hello Writing", WritingTopicName)
			if subscriberNotif != expectedSubscriberNotif {
				t.Errorf("expected subscriber notif %q, got %q", expectedSubscriberNotif, subscriberNotif)
			}
		})

		t.Run("Email Notification Verification", func(t *testing.T) {
			for emailqueue.ProcessPendingEmail(ctx, qs, mockProvider, cdlq, cfg) {
			}

			if len(mockProvider.sentMessages) != 1 {
				t.Fatalf("expected 1 email sent, got %d", len(mockProvider.sentMessages))
			}

			msg, err := mail.ReadMessage(strings.NewReader(string(mockProvider.sentMessages[0])))
			if err != nil {
				t.Fatalf("parse email: %v", err)
			}
			if msg.Header.Get("Subject") != "[goa4web] New reply in "+WritingTopicName {
				t.Errorf("subscriber email subject mismatch: %s", msg.Header.Get("Subject"))
			}
			subBody := getEmailBody(t, msg)
			expectedSubBody := fmt.Sprintf(
				"Hi replier,\n\nA new reply was posted in %q (thread #%d) on %s.\nThere are now %d comments in the discussion.\n\nView it here:\nhttp://example.com/writings/article/3\n\n\nManage notifications: http://example.com/usr/subscriptions",
				WritingTopicName,
				threadID,
				fixedTime.Format(consts.DisplayDateTimeFormat),
				4,
			)
			if subBody != expectedSubBody {
				t.Errorf("subscriber email body mismatch: %q, want %q", subBody, expectedSubBody)
			}
		})

		t.Run("Subscription Path Verification", func(t *testing.T) {
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
		})
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

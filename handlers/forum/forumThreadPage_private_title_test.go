package forum

import (
	"context"
	"database/sql"
	"fmt"
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
	"github.com/arran4/goa4web/internal/sharesign"
	"github.com/arran4/goa4web/workers/emailqueue"
)

func TestThreadPagePrivateSetsTitle(t *testing.T) {
	queries := &db.QuerierStub{
		GetThreadLastPosterAndPermsReturns: &db.GetThreadLastPosterAndPermsRow{
			Idforumthread:          1,
			Firstpost:              1,
			Lastposter:             1,
			ForumtopicIdforumtopic: 1,
			Comments:               sql.NullInt32{},
			Lastaddition:           sql.NullTime{},
			Locked:                 sql.NullBool{},
		},
		GetForumTopicByIdForUserReturns: &db.GetForumTopicByIdForUserRow{
			Idforumtopic:                 1,
			ForumcategoryIdforumcategory: 1,
			Title:                        sql.NullString{},
			Description:                  sql.NullString{},
			Threads:                      sql.NullInt32{},
			Comments:                     sql.NullInt32{},
			Lastaddition:                 sql.NullTime{},
			Handler:                      "private",
		},
		ListPrivateTopicParticipantsByTopicIDForUserReturns: []*db.ListPrivateTopicParticipantsByTopicIDForUserRow{
			{Idusers: 2, Username: sql.NullString{String: "Bob", Valid: true}},
		},
		GetCommentsByThreadIdForUserReturns: []*db.GetCommentsByThreadIdForUserRow{},
	}

	origStore := core.Store
	origName := core.SessionName
	core.Store = sessions.NewCookieStore([]byte("test"))
	core.SessionName = "test-session"
	defer func() {
		core.Store = origStore
		core.SessionName = origName
	}()

	req := httptest.NewRequest("GET", "/private/topic/1/thread/1", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "1"})
	// Inject an invalid session to force GetSessionOrFail to fail before rendering.
	ctx := context.WithValue(req.Context(), core.ContextValues("session"), "bad")

	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(ctx, queries, cfg)
	cd.ShareSigner = sharesign.NewSigner(cfg, "secret")
	cd.SetCurrentSection("privateforum")
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	ThreadPageWithBasePath(rr, req, "/private")
	if cd.PageTitle == "" {
		t.Fatalf("page title not set")
	}
}

func TestThreadPagePrivateReplyNotifications(t *testing.T) {
	replierUID := int32(1)
	subscriberUID := int32(2)
	adminUID := int32(99)
	topicID := int32(5)
	threadID := int32(42)

	qs := &db.QuerierStub{
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
			case adminUID:
				return &db.SystemGetUserByIDRow{
					Idusers:  adminUID,
					Username: sql.NullString{String: "adminuser", Valid: true},
					Email:    sql.NullString{String: "admin@example.com", Valid: true},
				}, nil
			}
			return nil, sql.ErrNoRows
		},
		SystemGetUserByEmailFn: func(ctx context.Context, email string) (*db.SystemGetUserByEmailRow, error) {
			if email == "admin@example.com" {
				return &db.SystemGetUserByEmailRow{Idusers: adminUID}, nil
			}
			return nil, sql.ErrNoRows
		},
		CreateCommentInSectionForCommenterFn: func(ctx context.Context, arg db.CreateCommentInSectionForCommenterParams) (int64, error) {
			return 999, nil
		},
		GetCommentByIdForUserRow: &db.GetCommentByIdForUserRow{
			Idcomments: 999,
		},
		GetThreadBySectionThreadIDForReplierFn: func(ctx context.Context, arg db.GetThreadBySectionThreadIDForReplierParams) (*db.Forumthread, error) {
			return &db.Forumthread{
				Idforumthread:          threadID,
				ForumtopicIdforumtopic: topicID,
			}, nil
		},
		GetThreadLastPosterAndPermsFn: func(ctx context.Context, arg db.GetThreadLastPosterAndPermsParams) (*db.GetThreadLastPosterAndPermsRow, error) {
			return &db.GetThreadLastPosterAndPermsRow{
				Idforumthread:          threadID,
				ForumtopicIdforumtopic: topicID,
				Lastposterusername:     sql.NullString{String: "replier", Valid: true},
				Comments:               sql.NullInt32{Int32: 1, Valid: true},
			}, nil
		},
		GetForumTopicByIdForUserFn: func(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
			return &db.GetForumTopicByIdForUserRow{
				Idforumtopic: topicID,
				Title:        sql.NullString{String: "Test Topic", Valid: true},
				Handler:      "private",
			}, nil
		},
		ListSubscribersForPatternsReturn: map[string][]int32{
			fmt.Sprintf("reply:/private/topic/%d/thread/%d/*", topicID, threadID): {subscriberUID},
		},
		GetPreferenceForListerReturn: map[int32]*db.Preference{
			replierUID:    {AutoSubscribeReplies: true},
			subscriberUID: {AutoSubscribeReplies: true},
		},
		AdminListAdministratorEmailsReturns:        []string{"admin@example.com"},
		UpsertContentReadMarkerFn:                  func(ctx context.Context, arg db.UpsertContentReadMarkerParams) error { return nil },
		ClearUnreadContentPrivateLabelExceptUserFn: func(ctx context.Context, arg db.ClearUnreadContentPrivateLabelExceptUserParams) error { return nil },
		AdminDeletePendingEmailFn:                  func(ctx context.Context, id int32) error { return nil },
		SystemMarkPendingEmailSentFn:               func(ctx context.Context, id int32) error { return nil },
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
		Path:    "/private/topic/5/thread/42/reply",
		Task:    task,
		Outcome: eventbus.TaskOutcomeSuccess,
	}
	ctx := context.Background()
	cd := common.NewCoreData(ctx, qs, cfg, common.WithSession(sess), common.WithEvent(evt), common.WithUserRoles([]string{"member"}))
	cd.UserID = replierUID
	cd.ForumBasePath = "/private"

	thread := &db.GetThreadLastPosterAndPermsRow{Idforumthread: threadID, ForumtopicIdforumtopic: topicID, Lastposterusername: sql.NullString{String: "replier", Valid: true}, Comments: sql.NullInt32{Int32: 1, Valid: true}}
	topic := &db.GetForumTopicByIdForUserRow{Idforumtopic: topicID, Title: sql.NullString{String: "Test Topic", Valid: true}, Handler: "private"}
	cd.SetCurrentThreadAndTopic(threadID, topicID)
	_, _ = cd.ForumThreadByID(threadID, lazy.Set(thread))
	_, _ = cd.ForumTopicByID(topicID, lazy.Set(topic))

	ctx = context.WithValue(ctx, core.ContextValues("session"), sess)
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)

	form := url.Values{"replytext": {"Hello World"}, "language": {"1"}}
	req := httptest.NewRequest(http.MethodPost, "http://example.com/private/topic/5/thread/42/reply", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, map[string]string{"topic": "5", "thread": "42"})

	rr := httptest.NewRecorder()
	task.Action(rr, req)

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

	expectedSubscriberNotif := "New reply in \"Test Topic\" by replier\n"
	if subscriberNotif != expectedSubscriberNotif {
		t.Errorf("expected subscriber notif %q, got %q", expectedSubscriberNotif, subscriberNotif)
	}

	expectedAdminNotif := "User replier replied to a forum thread.\nHello World\n"
	if adminNotif != expectedAdminNotif {
		t.Errorf("expected admin notif %q, got %q", expectedAdminNotif, adminNotif)
	}

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
	expectedSubBody := "Hi replier,\nYour reply has been posted.\n\nView comment:\n/private/topic/5/thread/42#c999\n\nManage notifications: http://example.com/usr/subscriptions"
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
	expectedAdminBody := "User replier replied to a forum thread.\nHello World\n\nView comment:\n/private/topic/5/thread/42#c999\n\nManage notifications: http://example.com/usr/subscriptions"
	if adminBody != expectedAdminBody {
		t.Errorf("admin email body mismatch: %q, want %q", adminBody, expectedAdminBody)
	}

	actionName, path, err := task.AutoSubscribePath(*evt)
	if err != nil {
		t.Errorf("AutoSubscribePath error: %v", err)
	}
	if actionName != "Reply" {
		t.Errorf("expected action name Reply, got %q", actionName)
	}
	if path != "/private/topic/5/thread/42" {
		t.Errorf("expected path /private/topic/5/thread/42, got %q", path)
	}
}

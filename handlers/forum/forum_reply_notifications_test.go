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
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/arran4/go-be-lazy"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/testhelpers"
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

// TestHappyPathForumReply verifies that forum reply use correct templates with real event data and exact string matching.
func TestHappyPathForumReply(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
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
			Text:       sql.NullString{String: "Original thread content", Valid: true},
		}
		qs.GetThreadBySectionThreadIDForReplierFn = func(ctx context.Context, arg db.GetThreadBySectionThreadIDForReplierParams) (*db.Forumthread, error) {
			return &db.Forumthread{
				Idforumthread:          threadID,
				ForumtopicIdforumtopic: topicID,
			}, nil
		}
		qs.GetThreadLastPosterAndPermsForUserFn = func(ctx context.Context, arg db.GetThreadLastPosterAndPermsForUserParams) (*db.GetThreadLastPosterAndPermsForUserRow, error) {
			return &db.GetThreadLastPosterAndPermsForUserRow{
				Idforumthread:          threadID,
				ForumtopicIdforumtopic: topicID,
				Lastposterusername:     sql.NullString{String: "replier", Valid: true},
				Comments:               sql.NullInt32{Int32: 1, Valid: true},
				Firstpost:              1,
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
			"notify:/admin/*": {adminUID},
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
		cfg.BaseURL = "http://example.com"

		mockProvider := &MockEmailProvider{}
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

		thread := &db.GetThreadLastPosterAndPermsForUserRow{Idforumthread: threadID, ForumtopicIdforumtopic: topicID, Lastposterusername: sql.NullString{String: "replier", Valid: true}, Comments: sql.NullInt32{Int32: 1, Valid: true}}
		topic := &db.GetForumTopicByIdForUserRow{Idforumtopic: topicID, Title: sql.NullString{String: "Test Topic", Valid: true}}
		cd.SetCurrentThreadAndTopic(threadID, topicID)
		_, _ = cd.ForumThreadByID(threadID, lazy.Set[int32, *db.GetThreadLastPosterAndPermsForUserRow](thread))
		_, _ = cd.ForumTopicByID(topicID, lazy.Set[int32, *db.GetForumTopicByIdForUserRow](topic))

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
		_ = bus.Publish(*evt)

		if cdlq.lastError != "" {
			t.Errorf("sync process error: %s", cdlq.lastError)
		}

		notificationsByRecipient := make(map[int32][]string)
		notificationsLinksByRecipient := make(map[int32][]string)
		for _, call := range qs.SystemCreateNotificationCalls {
			notificationsByRecipient[call.RecipientID] = append(notificationsByRecipient[call.RecipientID], call.Message.String)
			notificationsLinksByRecipient[call.RecipientID] = append(notificationsLinksByRecipient[call.RecipientID], call.Link.String)
		}

		findNotification := func(recipient int32, expectedMsg string, expectedLink string) bool {
			msgs := notificationsByRecipient[recipient]
			links := notificationsLinksByRecipient[recipient]
			for i, msg := range msgs {
				if msg == expectedMsg {
					if expectedLink == "" || links[i] == expectedLink {
						return true
					}
				}
			}
			return false
		}

		t.Run("Internal Notifications", func(t *testing.T) {
			t.Helper()
			expectedSubscriberNotif := "replier replied to Test Topic: This is a test message with a link [a https://example.com] and enough words to trigger the truncation of twenty words..."
			expectedLink := "http://example.com/forum/topic/5/thread/42#c999"
			if !findNotification(subscriberUID, expectedSubscriberNotif, expectedLink) {
				t.Fatalf("expected subscriber notification %q with link %q, got %q with links %q", expectedSubscriberNotif, expectedLink, notificationsByRecipient[subscriberUID], notificationsLinksByRecipient[subscriberUID])
			}
			if len(notificationsByRecipient[adminUID]) != 0 {
				t.Fatalf("expected non-subscriber admin %d to receive zero ordinary reply notifications, got %q", adminUID, notificationsByRecipient[adminUID])
			}

			if !findNotification(missingEmailUID, expectedSubscriberNotif, expectedLink) {
				t.Fatalf("expected missing email subscriber to receive internal notification %q, got %q", expectedSubscriberNotif, notificationsByRecipient[missingEmailUID])
			}
		})

		t.Run("Optional Email Activities", func(t *testing.T) {
			byRecipient := make(map[int32]db.SystemQueueForumReplyEmailActivityParams)
			for _, call := range qs.SystemQueueForumReplyEmailActivityCalls {
				byRecipient[call.RecipientID] = call
			}
			if len(qs.SystemQueueForumReplyEmailActivityCalls) != 2 || byRecipient[subscriberUID].ThreadID != threadID || byRecipient[missingEmailUID].ThreadID != threadID {
				t.Fatalf("expected durable activity for both subscribers: %+v", byRecipient)
			}
			if len(mockProvider.SentMessages) != 0 || len(qs.InsertPendingEmailCalls) != 0 {
				t.Fatal("reply email must wait for the batch worker")
			}
		})

		t.Run("Auto Subscribe Path", func(t *testing.T) {
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

			reqs, err := task.AutoSubscribeGrants(*evt)
			if err != nil {
				t.Errorf("AutoSubscribeGrants error: %v", err)
			}
			if len(reqs) != 1 {
				t.Errorf("expected 1 grant requirement, got %d", len(reqs))
			} else {
				req := reqs[0]
				if req.Section != "forum" || req.Item != "thread" || req.ItemID != threadID || req.Action != "view" {
					t.Errorf("unexpected grant requirement: %+v", req)
				}
			}
		})
	})
}

func TestForumAutoSubscribeTasks(t *testing.T) {
	if _, ok := any(replyTask).(notifications.AutoSubscribeProvider); !ok {
		t.Fatalf("ReplyTask should implement AutoSubscribeProvider so users get notified about thread replies")
	}
	if _, ok := any(createThreadTask).(notifications.AutoSubscribeProvider); !ok {
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
	if actionName != string(TaskReply) {
		t.Fatalf("expected action name %q, got %q", TaskReply, actionName)
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
	if actionName != string(TaskReply) {
		t.Fatalf("expected action name %q, got %q", TaskReply, actionName)
	}
	if path != "/forum/topic/44/thread/55" {
		t.Fatalf("expected create thread auto-subscribe path /forum/topic/44/thread/55, got %q", path)
	}

	privateThreadEvt := eventbus.TaskEvent{
		Data: map[string]any{
			postcountworker.EventKey: postcountworker.UpdateEventData{
				ThreadID:  64,
				TopicID:   20,
				CommentID: 100,
			},
		},
		Path: "/private/topic/20/thread/64",
	}
	actionName, path, err = createThreadTask.AutoSubscribePath(privateThreadEvt)
	if err != nil {
		t.Fatalf("private create thread AutoSubscribePath error: %v", err)
	}
	if actionName != string(TaskReply) {
		t.Fatalf("expected action name %q, got %q", TaskReply, actionName)
	}
	if path != "/private/topic/20/thread/64" {
		t.Fatalf("expected private create thread auto-subscribe path /private/topic/20/thread/64, got %q", path)
	}
}

func TestPrivateForumReply_CanonicalMatrixTransportAndAppend(t *testing.T) {
	const (
		actorUID            int32 = 1
		overlappingAdminUID int32 = 2
		ordinarySubUID      int32 = 3
		nonMemberAdminUID   int32 = 4
		revokedSubUID       int32 = 5
		topicID             int32 = 15
		threadID            int32 = 88
		firstCommentID      int32 = 500
	)

	buildStub := func(currentCommentText *string) *db.QuerierStub {
		qs := testhelpers.NewQuerierStub()
		qs.GetPermissionsByUserIDFn = func(idusers int32) ([]*db.GetPermissionsByUserIDRow, error) {
			return []*db.GetPermissionsByUserIDRow{}, nil
		}
		qs.SystemGetUserByIDFn = func(ctx context.Context, idusers int32) (*db.SystemGetUserByIDRow, error) {
			switch idusers {
			case actorUID:
				return &db.SystemGetUserByIDRow{Idusers: actorUID, Username: sql.NullString{String: "bob", Valid: true}, Email: sql.NullString{String: "bob@example.com", Valid: true}}, nil
			case overlappingAdminUID:
				return &db.SystemGetUserByIDRow{Idusers: overlappingAdminUID, Username: sql.NullString{String: "alice", Valid: true}, Email: sql.NullString{String: "alice@example.com", Valid: true}}, nil
			case ordinarySubUID:
				return &db.SystemGetUserByIDRow{Idusers: ordinarySubUID, Username: sql.NullString{String: "dave", Valid: true}, Email: sql.NullString{String: "dave@example.com", Valid: true}}, nil
			case nonMemberAdminUID:
				return &db.SystemGetUserByIDRow{Idusers: nonMemberAdminUID, Username: sql.NullString{String: "carol", Valid: true}, Email: sql.NullString{String: "carol@example.com", Valid: true}}, nil
			case revokedSubUID:
				return &db.SystemGetUserByIDRow{Idusers: revokedSubUID, Username: sql.NullString{String: "erin", Valid: true}, Email: sql.NullString{String: "erin@example.com", Valid: true}}, nil
			default:
				return nil, sql.ErrNoRows
			}
		}
		qs.SystemGetUserByEmailFn = func(ctx context.Context, addr string) (*db.SystemGetUserByEmailRow, error) {
			switch addr {
			case "alice@example.com":
				return &db.SystemGetUserByEmailRow{Idusers: overlappingAdminUID, Email: addr}, nil
			case "carol@example.com":
				return &db.SystemGetUserByEmailRow{Idusers: nonMemberAdminUID, Email: addr}, nil
			default:
				return nil, sql.ErrNoRows
			}
		}
		qs.AdminListAdministratorEmailsReturns = []string{"alice@example.com", "carol@example.com"}
		qs.GetForumTopicByIdForUserFn = func(ctx context.Context, arg db.GetForumTopicByIdForUserParams) (*db.GetForumTopicByIdForUserRow, error) {
			return &db.GetForumTopicByIdForUserRow{
				Idforumtopic: topicID,
				Title:        sql.NullString{String: "Staff Room", Valid: true},
				Handler:      "private",
			}, nil
		}
		qs.GetThreadBySectionThreadIDForReplierFn = func(ctx context.Context, arg db.GetThreadBySectionThreadIDForReplierParams) (*db.Forumthread, error) {
			return &db.Forumthread{
				Idforumthread:          threadID,
				ForumtopicIdforumtopic: topicID,
			}, nil
		}
		qs.GetThreadLastPosterAndPermsForUserFn = func(ctx context.Context, arg db.GetThreadLastPosterAndPermsForUserParams) (*db.GetThreadLastPosterAndPermsForUserRow, error) {
			return &db.GetThreadLastPosterAndPermsForUserRow{
				Idforumthread:          threadID,
				ForumtopicIdforumtopic: topicID,
				Lastposterusername:     sql.NullString{String: "bob", Valid: true},
				Comments:               sql.NullInt32{Int32: 1, Valid: true},
				Firstpost:              100,
			}, nil
		}
		qs.GetCommentByIdForUserFn = func(ctx context.Context, arg db.GetCommentByIdForUserParams) (*db.GetCommentByIdForUserRow, error) {
			if arg.ID == 100 {
				return &db.GetCommentByIdForUserRow{
					Idcomments: 100,
					Text:       sql.NullString{String: "Welcome to the private staff room.", Valid: true},
				}, nil
			}
			return &db.GetCommentByIdForUserRow{
				Idcomments: arg.ID,
				Text:       sql.NullString{String: *currentCommentText, Valid: true},
			}, nil
		}
		qs.CreateCommentInSectionForCommenterFn = func(ctx context.Context, arg db.CreateCommentInSectionForCommenterParams) (int64, error) {
			*currentCommentText = arg.Text.String
			return int64(firstCommentID), nil
		}
		qs.AppendCommentInSectionForCommenterFn = func(ctx context.Context, arg db.AppendCommentInSectionForCommenterParams) (int64, error) {
			*currentCommentText = *currentCommentText + fmt.Sprint(arg.Text)
			return 1, nil
		}
		qs.ListSubscribersForPatternsReturn = map[string][]int32{
			fmt.Sprintf("reply:/private/topic/%d/thread/%d/*", topicID, threadID): {actorUID, overlappingAdminUID, ordinarySubUID, revokedSubUID},
			"notify:/admin/*": {overlappingAdminUID, nonMemberAdminUID},
		}
		qs.GetPreferenceForListerReturn = map[int32]*db.Preference{
			actorUID:            {AutoSubscribeReplies: true},
			overlappingAdminUID: {AutoSubscribeReplies: true},
			ordinarySubUID:      {AutoSubscribeReplies: true},
		}
		qs.SystemCheckGrantFn = func(arg db.SystemCheckGrantParams) (int32, error) {
			switch arg.UserID.Int32 {
			case actorUID, overlappingAdminUID, ordinarySubUID:
				if (arg.Section == string(consts.PermissionSectionPrivateForum) && arg.Item.String == string(consts.PermissionItemTopic) && arg.ItemID.Int32 == topicID) ||
					(arg.Section == string(consts.PermissionSectionPrivateForumThread) && arg.Item.String == string(consts.PermissionItemThread) && arg.ItemID.Int32 == threadID) {
					return 1, nil
				}
			case revokedSubUID:
				// Revoked subscriber retained topic view but lost thread view.
				if arg.Section == string(consts.PermissionSectionPrivateForum) && arg.Item.String == string(consts.PermissionItemTopic) && arg.ItemID.Int32 == topicID && arg.Action == string(consts.PermissionActionView) {
					return 1, nil
				}
			}
			return 0, sql.ErrNoRows
		}
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
		return qs
	}

	for _, transportPath := range []string{
		fmt.Sprintf("/private/topic/%d/thread/%d/reply", topicID, threadID),
		fmt.Sprintf("/private/api/topic/%d/thread/%d/reply", topicID, threadID),
	} {
		t.Run("Transport_"+transportPath, func(t *testing.T) {
			commentBody := ""
			qs := buildStub(&commentBody)

			cfg := config.NewRuntimeConfig()
			cfg.NotificationsEnabled = true
			cfg.AdminNotify = true
			cfg.EmailEnabled = true
			cfg.EmailFrom = "noreply@example.com"
			cfg.BaseURL = "http://example.com"

			mockProvider := &MockEmailProvider{}
			n := notifications.New(
				notifications.WithSilence(true),
				notifications.WithQueries(qs),
				notifications.WithConfig(cfg),
				notifications.WithEmailProvider(mockProvider),
			)
			cdlq := &captureDLQ{}

			evt := &eventbus.TaskEvent{
				Data:    map[string]any{},
				UserID:  actorUID,
				Path:    transportPath,
				Task:    replyTask,
				Outcome: eventbus.TaskOutcomeSuccess,
			}
			ctx := context.Background()
			cd := common.NewCoreData(ctx, qs, cfg, common.WithEvent(evt), common.WithAbsoluteURLBase(cfg.BaseURL))
			cd.UserID = actorUID
			cd.ForumBasePath = "/private"

			result, err := cd.ReplyForumThread(ctx, common.ReplyForumThreadParams{
				ActorID:        actorUID,
				ThreadID:       threadID,
				LanguageID:     1,
				Text:           "Confidential reply from Bob",
				Private:        true,
				EnforceHandler: true,
				BasePath:       "/private",
			})
			if err != nil {
				t.Fatalf("ReplyForumThread: %v", err)
			}
			cd.SetEventData("URL", cd.AbsoluteURL(result.URL))

			if err := n.ProcessEvent(ctx, *evt, cdlq); err != nil {
				t.Fatalf("ProcessEvent: %v", err)
			}

			expectedMsg := "bob replied to Staff Room: Confidential reply from Bob"
			expectedLink := fmt.Sprintf("http://example.com/private/topic/%d/thread/%d#c%d", topicID, threadID, firstCommentID)

			notifsByRecipient := map[int32][]db.SystemCreateNotificationForActivityParams{}
			for _, call := range qs.SystemCreateNotificationForActivityCalls {
				notifsByRecipient[call.RecipientID] = append(notifsByRecipient[call.RecipientID], call)
			}

			if len(notifsByRecipient[overlappingAdminUID]) != 1 {
				t.Fatalf("overlapping admin+subscriber (user %d) notifications = %d, want 1: %+v", overlappingAdminUID, len(notifsByRecipient[overlappingAdminUID]), notifsByRecipient[overlappingAdminUID])
			}
			if notifsByRecipient[overlappingAdminUID][0].Message.String != expectedMsg {
				t.Errorf("user %d message = %q, want %q", overlappingAdminUID, notifsByRecipient[overlappingAdminUID][0].Message.String, expectedMsg)
			}
			if notifsByRecipient[overlappingAdminUID][0].Link.String != expectedLink {
				t.Errorf("user %d link = %q, want %q", overlappingAdminUID, notifsByRecipient[overlappingAdminUID][0].Link.String, expectedLink)
			}

			if len(notifsByRecipient[ordinarySubUID]) != 1 {
				t.Fatalf("ordinary subscriber (user %d) notifications = %d, want 1: %+v", ordinarySubUID, len(notifsByRecipient[ordinarySubUID]), notifsByRecipient[ordinarySubUID])
			}
			if notifsByRecipient[ordinarySubUID][0].Message.String != expectedMsg || notifsByRecipient[ordinarySubUID][0].Link.String != expectedLink {
				t.Errorf("user %d notification = %+v, want message %q and link %q", ordinarySubUID, notifsByRecipient[ordinarySubUID][0], expectedMsg, expectedLink)
			}

			if len(notifsByRecipient[nonMemberAdminUID]) != 0 {
				t.Errorf("non-participant admin (user %d) received private forum notification: %+v", nonMemberAdminUID, notifsByRecipient[nonMemberAdminUID])
			}
			if len(notifsByRecipient[revokedSubUID]) != 0 {
				t.Errorf("revoked-grant subscriber (user %d) received private forum notification: %+v", revokedSubUID, notifsByRecipient[revokedSubUID])
			}
			if len(notifsByRecipient[actorUID]) != 0 {
				t.Errorf("actor (user %d) received self-notification: %+v", actorUID, notifsByRecipient[actorUID])
			}

			byRecipient := make(map[int32]int64)
			for _, call := range qs.SystemQueueForumReplyEmailActivityCalls {
				byRecipient[call.RecipientID] = call.ActivityPosition
			}
			if len(qs.SystemQueueForumReplyEmailActivityCalls) != 2 || byRecipient[overlappingAdminUID] == 0 || byRecipient[ordinarySubUID] == 0 {
				t.Fatalf("expected email activity for Alice and Dave only: %+v", byRecipient)
			}
			if len(mockProvider.Recipients) != 0 || len(qs.InsertPendingEmailCalls) != 0 {
				t.Fatal("private replies must wait for the batch worker")
			}

		})
	}

	t.Run("AppendedConsecutiveReplyProducesDistinctActivity", func(t *testing.T) {
		commentBody := "Initial reply from Bob"
		qs := buildStub(&commentBody)
		hasFirstComment := false
		qs.GetCommentsByThreadIdForUserFn = func(ctx context.Context, arg db.GetCommentsByThreadIdForUserParams) ([]*db.GetCommentsByThreadIdForUserRow, error) {
			if !hasFirstComment {
				return nil, nil
			}
			return []*db.GetCommentsByThreadIdForUserRow{
				{
					Idcomments:    firstCommentID,
					UsersIdusers:  actorUID,
					ForumthreadID: threadID,
					IsOwner:       true,
					Written:       sql.NullTime{Time: time.Now().Add(-time.Minute), Valid: true},
					Text:          sql.NullString{String: commentBody, Valid: true},
				},
			}, nil
		}

		cfg := config.NewRuntimeConfig()
		cfg.NotificationsEnabled = true
		cfg.PrivateForumPostAppendWindow = 60
		cfg.BaseURL = "http://example.com"

		n := notifications.New(
			notifications.WithSilence(true),
			notifications.WithQueries(qs),
			notifications.WithConfig(cfg),
		)

		ctx := context.Background()

		// 1. First reply creates comment 500
		evt1 := &eventbus.TaskEvent{
			Data:    map[string]any{},
			UserID:  actorUID,
			Path:    fmt.Sprintf("/private/topic/%d/thread/%d/reply", topicID, threadID),
			Task:    replyTask,
			Outcome: eventbus.TaskOutcomeSuccess,
		}
		cd1 := common.NewCoreData(ctx, qs, cfg, common.WithEvent(evt1), common.WithAbsoluteURLBase(cfg.BaseURL))
		cd1.UserID = actorUID
		cd1.ForumBasePath = "/private"
		res1, err := cd1.ReplyForumThread(ctx, common.ReplyForumThreadParams{
			ActorID:        actorUID,
			ThreadID:       threadID,
			LanguageID:     1,
			Text:           "Initial reply from Bob",
			Private:        true,
			EnforceHandler: true,
			BasePath:       "/private",
		})
		if err != nil {
			t.Fatalf("first ReplyForumThread: %v", err)
		}
		hasFirstComment = true
		cd1.SetEventData("URL", cd1.AbsoluteURL(res1.URL))
		if err := n.ProcessEvent(ctx, *evt1, nil); err != nil {
			t.Fatalf("first ProcessEvent: %v", err)
		}

		// 2. Second reply within append window appends to comment 500
		evt2 := &eventbus.TaskEvent{
			Data:    map[string]any{},
			UserID:  actorUID,
			Path:    fmt.Sprintf("/private/topic/%d/thread/%d/reply", topicID, threadID),
			Task:    replyTask,
			Outcome: eventbus.TaskOutcomeSuccess,
		}
		cd2 := common.NewCoreData(ctx, qs, cfg, common.WithEvent(evt2), common.WithAbsoluteURLBase(cfg.BaseURL))
		cd2.UserID = actorUID
		cd2.ForumBasePath = "/private"
		res2, err := cd2.ReplyForumThread(ctx, common.ReplyForumThreadParams{
			ActorID:        actorUID,
			ThreadID:       threadID,
			LanguageID:     1,
			Text:           "Appended follow-up from Bob",
			Private:        true,
			EnforceHandler: true,
			BasePath:       "/private",
		})
		if err != nil {
			t.Fatalf("second (appended) ReplyForumThread: %v", err)
		}
		cd2.SetEventData("URL", cd2.AbsoluteURL(res2.URL))
		if err := n.ProcessEvent(ctx, *evt2, nil); err != nil {
			t.Fatalf("second ProcessEvent: %v", err)
		}

		var aliceKeys []string
		for _, call := range qs.SystemCreateNotificationForActivityCalls {
			if call.RecipientID == overlappingAdminUID {
				aliceKeys = append(aliceKeys, call.ActivityKey.String)
			}
		}
		if len(aliceKeys) != 2 {
			t.Fatalf("expected 2 notification activity calls for Alice across initial + appended reply, got %d: %v", len(aliceKeys), aliceKeys)
		}
		if aliceKeys[0] == aliceKeys[1] {
			t.Fatalf("expected appended reply to have a distinct activity key from initial reply, both were %q", aliceKeys[0])
		}
	})

	t.Run("AppendedConcurrentInterleavedRepliesProduceDistinctActivitiesAndDeduplicateRetries", func(t *testing.T) {
		commentBody := "Initial reply from Bob"
		qs := buildStub(&commentBody)

		// Reordered reload: any reload returns the final combined text containing both appends.
		finalText := "Initial reply from Bob\n\n[hr]\n\nAppended reply 1\n\n[hr]\n\nAppended reply 2"
		qs.GetCommentsByThreadIdForUserFn = func(ctx context.Context, arg db.GetCommentsByThreadIdForUserParams) ([]*db.GetCommentsByThreadIdForUserRow, error) {
			return []*db.GetCommentsByThreadIdForUserRow{
				{
					Idcomments:    firstCommentID,
					UsersIdusers:  actorUID,
					ForumthreadID: threadID,
					IsOwner:       true,
					Written:       sql.NullTime{Time: time.Now().Add(-time.Minute), Valid: true},
					Text:          sql.NullString{String: finalText, Valid: true},
				},
			}, nil
		}
		qs.GetCommentByIdForUserFn = func(ctx context.Context, arg db.GetCommentByIdForUserParams) (*db.GetCommentByIdForUserRow, error) {
			return &db.GetCommentByIdForUserRow{
				Idcomments: firstCommentID,
				Text:       sql.NullString{String: finalText, Valid: true},
			}, nil
		}

		cfg := config.NewRuntimeConfig()
		cfg.NotificationsEnabled = true
		cfg.PrivateForumPostAppendWindow = 60
		cfg.BaseURL = "http://example.com"

		n := notifications.New(
			notifications.WithSilence(true),
			notifications.WithQueries(qs),
			notifications.WithConfig(cfg),
		)

		ctx := context.Background()

		// Append 1
		evt1 := &eventbus.TaskEvent{
			Data:    map[string]any{},
			UserID:  actorUID,
			Path:    fmt.Sprintf("/private/topic/%d/thread/%d/reply", topicID, threadID),
			Task:    replyTask,
			Outcome: eventbus.TaskOutcomeSuccess,
		}
		cd1 := common.NewCoreData(ctx, qs, cfg, common.WithEvent(evt1), common.WithAbsoluteURLBase(cfg.BaseURL))
		cd1.UserID = actorUID
		cd1.ForumBasePath = "/private"
		res1, err := cd1.ReplyForumThread(ctx, common.ReplyForumThreadParams{
			ActorID:        actorUID,
			ThreadID:       threadID,
			LanguageID:     1,
			Text:           "Appended reply 1",
			Private:        true,
			EnforceHandler: true,
			BasePath:       "/private",
		})
		if err != nil {
			t.Fatalf("append 1: %v", err)
		}
		cd1.SetEventData("URL", cd1.AbsoluteURL(res1.URL))
		if err := n.ProcessEvent(ctx, *evt1, nil); err != nil {
			t.Fatalf("process append 1: %v", err)
		}

		// Append 2 (concurrent / interleaved reload seeing same finalText)
		evt2 := &eventbus.TaskEvent{
			Data:    map[string]any{},
			UserID:  actorUID,
			Path:    fmt.Sprintf("/private/topic/%d/thread/%d/reply", topicID, threadID),
			Task:    replyTask,
			Outcome: eventbus.TaskOutcomeSuccess,
		}
		cd2 := common.NewCoreData(ctx, qs, cfg, common.WithEvent(evt2), common.WithAbsoluteURLBase(cfg.BaseURL))
		cd2.UserID = actorUID
		cd2.ForumBasePath = "/private"
		res2, err := cd2.ReplyForumThread(ctx, common.ReplyForumThreadParams{
			ActorID:        actorUID,
			ThreadID:       threadID,
			LanguageID:     1,
			Text:           "Appended reply 2",
			Private:        true,
			EnforceHandler: true,
			BasePath:       "/private",
		})
		if err != nil {
			t.Fatalf("append 2: %v", err)
		}
		cd2.SetEventData("URL", cd2.AbsoluteURL(res2.URL))
		if err := n.ProcessEvent(ctx, *evt2, nil); err != nil {
			t.Fatalf("process append 2: %v", err)
		}

		// Verify distinct activity identities
		if res1.ActivityID == res2.ActivityID {
			t.Fatalf("concurrent appends produced identical ActivityID %q", res1.ActivityID)
		}
		if res1.AppendID == res2.AppendID {
			t.Fatalf("concurrent appends produced identical AppendID %d", res1.AppendID)
		}

		var aliceKeys []string
		for _, call := range qs.SystemCreateNotificationForActivityCalls {
			if call.RecipientID == overlappingAdminUID {
				aliceKeys = append(aliceKeys, call.ActivityKey.String)
			}
		}
		if len(aliceKeys) != 2 {
			t.Fatalf("expected 2 notification calls for Alice across both appends, got %d: %v", len(aliceKeys), aliceKeys)
		}
		if aliceKeys[0] == aliceKeys[1] {
			t.Fatalf("expected distinct activity keys for Alice, got %v", aliceKeys)
		}

		// Verify retry idempotency: re-delivering append 1 reuses its original ActivityID and activity key
		if err := n.ProcessEvent(ctx, *evt1, nil); err != nil {
			t.Fatalf("retry process append 1: %v", err)
		}
		var retryAliceKeys []string
		for _, call := range qs.SystemCreateNotificationForActivityCalls {
			if call.RecipientID == overlappingAdminUID {
				retryAliceKeys = append(retryAliceKeys, call.ActivityKey.String)
			}
		}
		if len(retryAliceKeys) != 3 {
			t.Fatalf("expected 3 total calls after retry, got %d", len(retryAliceKeys))
		}
		if retryAliceKeys[2] != aliceKeys[0] {
			t.Fatalf("retry activity key = %q, want original %q", retryAliceKeys[2], aliceKeys[0])
		}
	})
}

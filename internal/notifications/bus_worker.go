package notifications

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dlq"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/tasks"
)

// buildPatterns expands the task/path pair into all matching subscription patterns.
func buildPatterns(task tasks.Name, path string) []string {
	name := strings.ToLower(task.Name())
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{fmt.Sprintf("%s:/*", name)}
	}
	parts := strings.Split(path, "/")
	patterns := []string{fmt.Sprintf("%s:/%s", name, path)}
	for i := len(parts) - 1; i >= 1; i-- {
		prefix := strings.Join(parts[:i], "/")
		patterns = append(patterns, fmt.Sprintf("%s:/%s/*", name, prefix))
	}
	patterns = append(patterns, fmt.Sprintf("%s:/*", name))
	return patterns
}

// collectSubscribers returns a set of user IDs subscribed to any of the
// patterns using the specified delivery method.
func collectSubscribers(ctx context.Context, q db.Querier, patterns []string, method string) (map[int32]struct{}, error) {
	subs := map[int32]struct{}{}
	ids, err := q.ListSubscribersForPatterns(ctx, db.ListSubscribersForPatternsParams{Patterns: patterns, Method: method})
	if err != nil {
		return nil, fmt.Errorf("list subscribers: %w", err)
	}
	for _, id := range ids {
		subs[id] = struct{}{}
	}
	return subs, nil
}

func (n *Notifier) BusWorker(ctx context.Context, bus *eventbus.Bus, q dlq.DLQ) {
	if bus == nil || n.Queries == nil {
		return
	}
	ch := bus.Subscribe(eventbus.TaskMessageType)
	for {
		select {
		case msg := <-ch:
			evt, ok := msg.(eventbus.TaskEvent)
			if !ok {
				continue
			}
			if err := n.ProcessEvent(ctx, evt, q); err != nil {
				log.Printf("process event: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// RegisterSync configures the bus to deliver events synchronously to this notifier.
func (n *Notifier) RegisterSync(bus *eventbus.Bus, q dlq.DLQ) {
	bus.SyncPublish = func(msg eventbus.Message) {
		if evt, ok := msg.(eventbus.TaskEvent); ok {
			if err := n.ProcessEvent(context.Background(), evt, q); err != nil {
				log.Printf("sync process event: %v", err)
			}
		}
	}
}

func (n *Notifier) ProcessEvent(ctx context.Context, evt eventbus.TaskEvent, q dlq.DLQ) error {
	if !n.Config.NotificationsEnabled {
		return nil
	}

	if evt.Outcome != eventbus.TaskOutcomeSuccess {
		return nil
	}

	if evt.Task == nil {
		return nil
	}

	if tp, ok := evt.Task.(AdminEmailTemplateProvider); ok {
		if et, send := tp.AdminEmailTemplate(evt); send {
			if err := n.notifyAdmins(ctx, et, tp.AdminInternalNotificationTemplate(evt), evt.Data, evt.Path); err != nil {
				errW := fmt.Errorf("AdminEmailTemplateProvider: %w", err)
				if dlqErr := n.dlqRecordAndNotify(ctx, q, fmt.Sprintf("admin notify: %v", errW)); dlqErr != nil {
					return dlqErr
				}
				return errW
			}
		}
	}

	if tp, ok := evt.Task.(SelfNotificationTemplateProvider); ok {
		if err := n.notifySelf(ctx, evt, tp); err != nil {
			errW := fmt.Errorf("SelfNotificationTemplateProvider: %w", err)
			if dlqErr := n.dlqRecordAndNotify(ctx, q, fmt.Sprintf("deliver self to %d: %v", evt.UserID, errW)); dlqErr != nil {
				return dlqErr
			}
			return errW
		}

	}

	if tp, ok := evt.Task.(DirectEmailNotificationTemplateProvider); ok {
		if err := n.notifyDirectEmail(ctx, evt, tp); err != nil {
			errW := fmt.Errorf("DirectEmailNotificationTemplateProvider: %w", err)
			if dlqErr := n.dlqRecordAndNotify(ctx, q, fmt.Sprintf("direct email notify: %v", errW)); dlqErr != nil {
				return dlqErr
			}
			return errW
		}

	}

	if tp, ok := evt.Task.(TargetUsersNotificationProvider); ok {
		if err := n.notifyTargetUsers(ctx, evt, tp); err != nil {
			errW := fmt.Errorf("TargetUsersNotificationProvider: %w", err)
			if dlqErr := n.dlqRecordAndNotify(ctx, q, fmt.Sprintf("notify target users: %v", errW)); dlqErr != nil {
				return dlqErr
			}
			return errW
		}

	}

	if tp, ok := evt.Task.(SubscribersNotificationTemplateProvider); ok {
		if err := n.notifySubscribers(ctx, evt, tp); err != nil {
			errW := fmt.Errorf("SubscribersNotificationTemplateProvider: %w", err)
			if dlqErr := n.dlqRecordAndNotify(ctx, q, fmt.Sprintf("notify subscribers: %v", errW)); dlqErr != nil {
				return dlqErr
			}
			return errW
		}

	}

	if tp, ok := evt.Task.(AutoSubscribeProvider); ok {
		if err := n.handleAutoSubscribe(ctx, evt, tp); err != nil {
			errW := fmt.Errorf("AutoSubscribeProvider: %w", err)
			return errW
		}

	}

	return nil
}

func (n *Notifier) notifySelf(ctx context.Context, evt eventbus.TaskEvent, tp SelfNotificationTemplateProvider) error {
	if et, send := tp.SelfEmailTemplate(evt); send {
		if b, ok := evt.Task.(SelfEmailBroadcaster); ok && b.SelfEmailBroadcast() {
			emails, err := n.Queries.SystemListVerifiedEmailsByUserID(ctx, evt.UserID)
			if err == nil {
				for _, e := range emails {
					if err := n.renderAndQueueEmailFromTemplates(ctx, &evt.UserID, e.Email, et, evt.Data); err != nil {
						return err
					}
				}
			}
		} else {
			ue, err := n.Queries.GetNotificationEmailByUserID(ctx, evt.UserID)
			if err != nil {
				if nmErr := notifyMissingEmail(ctx, n.Queries, evt.UserID); nmErr != nil {
					log.Printf("notify missing email: %v", nmErr)
				}
			} else {
				if err := n.renderAndQueueEmailFromTemplates(ctx, &evt.UserID, ue.Email, et, evt.Data); err != nil {
					return err
				}
			}
		}
	}
	if nt := tp.SelfInternalNotificationTemplate(evt); nt != nil {
		data := struct {
			Path      string
			Item      any
			TaskEvent eventbus.TaskEvent
			UserID    int32
		}{
			Path:      evt.Path,
			Item:      evt.Data,
			TaskEvent: evt,
			UserID:    evt.UserID,
		}
		msg, err := n.renderNotification(ctx, *nt, data)
		if err != nil {
			return err
		}
		if len(msg) > 0 {
			if err := n.Queries.SystemCreateNotification(ctx, db.SystemCreateNotificationParams{
				RecipientID: evt.UserID,
				Link:        sql.NullString{String: evt.Path, Valid: true},
				Message:     sql.NullString{String: string(msg), Valid: true},
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (n *Notifier) notifyDirectEmail(ctx context.Context, evt eventbus.TaskEvent, tp DirectEmailNotificationTemplateProvider) error {
	addr, err := tp.DirectEmailAddress(evt)
	if err != nil {
		return err
	}
	if addr == "" {
		return nil
	}
	if et, send := tp.DirectEmailTemplate(evt); send {
		if err := n.renderAndQueueEmailFromTemplates(ctx, nil, addr, et, evt.Data); err != nil {
			return err
		}
	}
	return nil
}

func (n *Notifier) notifyTargetUsers(ctx context.Context, evt eventbus.TaskEvent, tp TargetUsersNotificationProvider) error {
	ids, err := tp.TargetUserIDs(evt)
	if err != nil {
		return err
	}
	for _, id := range ids {
		user, err := n.Queries.SystemGetUserByID(ctx, id)
		if err != nil || !user.Email.Valid || user.Email.String == "" {
			if nmErr := notifyMissingEmail(ctx, n.Queries, id); nmErr != nil {
				log.Printf("notify missing email: %v", nmErr)
			}
		} else {
			if et, send := tp.TargetEmailTemplate(evt); send {
				if err := n.renderAndQueueEmailFromTemplates(ctx, &id, user.Email.String, et, evt.Data); err != nil {
					return err
				}
			}
		}
		if nt := tp.TargetInternalNotificationTemplate(evt); nt != nil {
			data := struct {
				TaskEvent eventbus.TaskEvent
				Path      string
				Item      any
			}{
				TaskEvent: evt,
				Path:      evt.Path,
				Item:      evt.Data,
			}
			msg, err := n.renderNotification(ctx, *nt, data)
			if err != nil {
				return err
			}
			if len(msg) > 0 {
				if err := n.Queries.SystemCreateNotification(ctx, db.SystemCreateNotificationParams{
					RecipientID: id,
					Link:        sql.NullString{String: evt.Path, Valid: true},
					Message:     sql.NullString{String: string(msg), Valid: true},
				}); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (n *Notifier) notifySubscribers(ctx context.Context, evt eventbus.TaskEvent, tp SubscribersNotificationTemplateProvider) error {
	name := ""
	if tn, ok := evt.Task.(tasks.Name); ok {
		name = tn.Name()
	}
	patterns := buildPatterns(tasks.TaskString(name), evt.Path)

	emailSubs, err := collectSubscribers(ctx, n.Queries, patterns, "email")
	if err != nil {
		return err
	}
	internalSubs, err := collectSubscribers(ctx, n.Queries, patterns, "internal")
	if err != nil {
		return err
	}

	delete(emailSubs, evt.UserID)
	delete(internalSubs, evt.UserID)

	if gp, ok := evt.Task.(GrantsRequiredProvider); ok {
		reqs, err := gp.GrantsRequired(evt)
		if err != nil {
			return err
		}
		if len(reqs) != 0 {
			filterSubs := func(m map[int32]struct{}) {
				for id := range m {
					for _, g := range reqs {
						if _, err := n.Queries.SystemCheckGrant(ctx, db.SystemCheckGrantParams{
							ViewerID: id,
							Section:  g.Section,
							Item:     sql.NullString{String: g.Item, Valid: g.Item != ""},
							Action:   g.Action,
							ItemID:   sql.NullInt32{Int32: g.ItemID, Valid: g.ItemID != 0},
							UserID:   sql.NullInt32{Int32: id, Valid: id != 0},
						}); err != nil {
							delete(m, id)
							break
						}
					}
				}
			}
			filterSubs(emailSubs)
			filterSubs(internalSubs)
		}
	}

	var msg []byte
	data := EmailData{Item: evt.Data, any: evt.Data}
	if nt := tp.SubscribedInternalNotificationTemplate(evt); nt != nil {
		var err error
		msg, err = n.renderNotification(ctx, *nt, data)
		if err != nil {
			log.Printf("render subscriber notification: %v", err)
			return fmt.Errorf("render notification: %w", err)
		}
	}

	if et, send := tp.SubscribedEmailTemplate(evt); send {
		for id := range emailSubs {
			if err := n.sendSubscriberEmail(ctx, id, evt, et); err != nil {
				return fmt.Errorf("deliver email to %d: %w", id, err)
			}
		}
	}

	if len(msg) != 0 {
		for id := range internalSubs {
			if err := n.sendInternalNotification(ctx, id, evt.Path, string(msg)); err != nil {
				return fmt.Errorf("deliver internal to %d: %w", id, err)
			}
		}
	}

	return nil
}

func (n *Notifier) handleAutoSubscribe(ctx context.Context, evt eventbus.TaskEvent, tp AutoSubscribeProvider) error {
	var auto bool
	var email bool
	pref, err := n.Queries.GetPreferenceForLister(ctx, evt.UserID)
	if err != nil {
		return fmt.Errorf("get preference by user_id: %w", err)
	}
	auto = pref.AutoSubscribeReplies
	if pref.Emailforumupdates.Valid {
		email = pref.Emailforumupdates.Bool
	}
	if auto {
		task, path, err := tp.AutoSubscribePath(evt)
		if err != nil {
			log.Printf("auto subscribe path: %v", err)
			return fmt.Errorf("auto subscribe path: %w", err)
		}
		pattern := buildPatterns(tasks.TaskString(task), path)[0]
		if n.Config.NotificationsEnabled {
			ensureSubscription(ctx, n.Queries, evt.UserID, pattern, "internal")
		}
		if email {
			ensureSubscription(ctx, n.Queries, evt.UserID, pattern, "email")
		}
	}
	return nil
}

func ensureSubscription(ctx context.Context, q db.Querier, userID int32, pattern, method string) {
	if q == nil || userID == 0 {
		return
	}
	ids, err := q.ListSubscribersForPattern(ctx, db.ListSubscribersForPatternParams{Pattern: pattern, Method: method})
	if err == nil {
		for _, id := range ids {
			if id == userID {
				return
			}
		}
	}
	if err := q.InsertSubscription(ctx, db.InsertSubscriptionParams{UsersIdusers: userID, Pattern: pattern, Method: method}); err != nil {
		log.Printf("insert subscription: %v", err)
	}
}

func notifyMissingEmail(ctx context.Context, q db.Querier, userID int32) error {
	if q == nil || userID == 0 {
		return nil
	}
	last, err := q.SystemGetLastNotificationForRecipientByMessage(ctx, db.SystemGetLastNotificationForRecipientByMessageParams{RecipientID: userID, Message: sql.NullString{String: "missing email address", Valid: true}})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("last notification: %w", err)
	}
	if err == nil && time.Since(last.CreatedAt) < 7*24*time.Hour {
		return nil
	}
	if err := q.SystemCreateNotification(ctx, db.SystemCreateNotificationParams{RecipientID: userID, Message: sql.NullString{String: "missing email address", Valid: true}}); err != nil {
		return fmt.Errorf("insert notification: %w", err)
	}
	return nil
}

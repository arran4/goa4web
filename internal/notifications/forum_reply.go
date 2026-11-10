package notifications

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dlq"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/stats"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/workers/postcountworker"
)

const (
	// forumReplyNotificationTemplate is the canonical internal notification template for forum replies.
	forumReplyNotificationTemplate = NotificationTemplateName("reply")
	// forumReplyEmailTemplate is the canonical email template prefix for forum replies.
	forumReplyEmailTemplate = EmailTemplateName("forumReplyEmail")
	// forumReplyTaskName is the canonical subscription task name for forum replies.
	forumReplyTaskName = tasks.TaskString("Reply")
)

var _ common.ForumReplyNotifier = (*Notifier)(nil)

// NotifyForumReply resolves eligible recipients and delivers one canonical
// forum-reply notification and subscriber email per recipient.
func (n *Notifier) NotifyForumReply(ctx context.Context, intent common.ForumReplyNotificationIntent) error {
	return n.NotifyForumReplyWithDLQ(ctx, intent, nil)
}

// OnForumReply aliases NotifyForumReply for callers using event-listener naming.
func (n *Notifier) OnForumReply(ctx context.Context, intent common.ForumReplyNotificationIntent) error {
	return n.NotifyForumReply(ctx, intent)
}

// NotifyForumReplyWithDLQ delivers a canonical forum.reply notification intent and records failures to q when provided.
func (n *Notifier) NotifyForumReplyWithDLQ(ctx context.Context, intent common.ForumReplyNotificationIntent, q dlq.DLQ) error {
	intent = n.normalizeForumReplyIntent(intent, nil)
	evt := buildTaskEventFromForumReplyIntent(intent)
	return n.notifyForumReplyFromEvent(ctx, evt, intent, q)
}

func (n *Notifier) extractForumReplyIntent(evt eventbus.TaskEvent) (common.ForumReplyNotificationIntent, bool) {
	if evt.Data != nil {
		switch v := evt.Data[common.ForumReplyIntentEventKey].(type) {
		case common.ForumReplyNotificationIntent:
			return n.normalizeForumReplyIntent(v, &evt), true
		case *common.ForumReplyNotificationIntent:
			if v != nil {
				return n.normalizeForumReplyIntent(*v, &evt), true
			}
		}
	}
	if evt.Task == nil {
		return common.ForumReplyNotificationIntent{}, false
	}
	tn, ok := evt.Task.(tasks.Name)
	if !ok || !strings.EqualFold(tn.Name(), "Reply") {
		return common.ForumReplyNotificationIntent{}, false
	}
	if !isForumReplyEventPath(evt.Path, evt.Data) {
		return common.ForumReplyNotificationIntent{}, false
	}
	return n.normalizeForumReplyIntent(common.ForumReplyNotificationIntent{}, &evt), true
}

func isForumReplyEventPath(path string, data map[string]any) bool {
	trimmed := strings.TrimSpace(path)
	return strings.HasPrefix(trimmed, "/forum/") ||
		strings.HasPrefix(trimmed, "/forum") ||
		strings.HasPrefix(trimmed, "/private/") ||
		strings.HasPrefix(trimmed, "/private") ||
		strings.HasPrefix(trimmed, "/private/api/")
}

func (n *Notifier) normalizeForumReplyIntent(intent common.ForumReplyNotificationIntent, evt *eventbus.TaskEvent) common.ForumReplyNotificationIntent {
	if evt != nil {
		if intent.ActorID == 0 {
			intent.ActorID = evt.UserID
		}
		if intent.ActivityID == "" {
			intent.ActivityID = strings.TrimSpace(evt.ActivityID)
		}
		if evt.Data != nil {
			if update, ok := evt.Data[postcountworker.EventKey].(postcountworker.UpdateEventData); ok {
				if intent.TopicID == 0 {
					intent.TopicID = update.TopicID
				}
				if intent.ThreadID == 0 {
					intent.ThreadID = update.ThreadID
				}
				if intent.CommentID == 0 {
					intent.CommentID = update.CommentID
				}
			}
			if intent.CommentID == 0 {
				if cid, ok := int32FromAny(evt.Data["CommentID"]); ok {
					intent.CommentID = cid
				}
			}
			if intent.CommentAnchor == "" {
				if anchor, ok := evt.Data["CommentAnchor"].(string); ok {
					intent.CommentAnchor = anchor
				}
			}
			if intent.CommentText == "" {
				if text, ok := evt.Data["CommentText"].(string); ok && text != "" {
					intent.CommentText = text
				} else if body, ok := evt.Data["Body"].(string); ok && body != "" {
					intent.CommentText = body
				}
			}
			if intent.ThreadOpenerPreview == "" {
				if text, ok := evt.Data["Text"].(string); ok {
					intent.ThreadOpenerPreview = text
				}
			}
			if intent.Username == "" {
				if username, ok := evt.Data["Username"].(string); ok {
					intent.Username = username
				}
			}
			if intent.TopicTitle == "" {
				if title, ok := evt.Data["TopicTitle"].(string); ok {
					intent.TopicTitle = title
				}
			}
			if intent.SubjectPrefix == "" {
				if prefix, ok := evt.Data["SubjectPrefix"].(string); ok {
					intent.SubjectPrefix = prefix
				}
			}
			if priv, ok := evt.Data["Private"].(bool); ok && priv {
				intent.Private = true
			}
		}
	}

	rawPath := intent.CanonicalPath
	if rawPath == "" && evt != nil {
		rawPath = evt.Path
	}
	if strings.HasPrefix(rawPath, "/private") || (evt != nil && strings.HasPrefix(evt.Path, "/private")) {
		intent.Private = true
	}
	if intent.TopicID == 0 || intent.ThreadID == 0 {
		topicID, threadID := parseForumTopicAndThreadIDs(rawPath)
		if intent.TopicID == 0 {
			intent.TopicID = topicID
		}
		if intent.ThreadID == 0 {
			intent.ThreadID = threadID
		}
	}

	basePath := common.CanonicalForumBasePath(intent.Private, rawPath)
	if intent.TopicID != 0 && intent.ThreadID != 0 {
		intent.CanonicalPath = fmt.Sprintf("%s/topic/%d/thread/%d", basePath, intent.TopicID, intent.ThreadID)
	} else if intent.CanonicalPath == "" && rawPath != "" {
		intent.CanonicalPath = normalizeForumThreadPath(rawPath, intent.Private)
	}

	if intent.CommentAnchor == "" && intent.CommentID != 0 {
		intent.CommentAnchor = fmt.Sprintf("c%d", intent.CommentID)
	}
	if intent.ActivityID == "" && intent.CommentID != 0 {
		intent.ActivityID = fmt.Sprintf("comment:%d", intent.CommentID)
	}

	hasExplicitEventURL := false
	if evt != nil && evt.Data != nil {
		if rawURL, ok := evt.Data["URL"].(string); ok && rawURL != "" && !strings.Contains(rawURL, "/private/api/") {
			hasExplicitEventURL = true
			if intent.CommentURL == "" {
				intent.CommentURL = rawURL
			}
			if intent.AbsoluteCommentURL == "" {
				intent.AbsoluteCommentURL = rawURL
			}
		}
	}
	if intent.CommentURL == "" && intent.CanonicalPath != "" {
		if intent.CommentAnchor != "" {
			intent.CommentURL = fmt.Sprintf("%s#%s", intent.CanonicalPath, intent.CommentAnchor)
		} else {
			intent.CommentURL = intent.CanonicalPath
		}
	}

	if intent.AbsoluteCommentURL == "" {
		intent.AbsoluteCommentURL = intent.CommentURL
		if !hasExplicitEventURL && n.Config != nil && n.Config.BaseURL != "" && intent.CommentURL != "" &&
			!strings.HasPrefix(intent.CommentURL, "http://") && !strings.HasPrefix(intent.CommentURL, "https://") {
			intent.AbsoluteCommentURL = strings.TrimRight(n.Config.BaseURL, "/") + "/" + strings.TrimLeft(intent.CommentURL, "/")
		}
	}

	if intent.SubjectPrefix == "" && n.Config != nil {
		intent.SubjectPrefix = n.Config.EmailSubjectPrefix
	}
	return intent
}

func parseForumTopicAndThreadIDs(path string) (int32, int32) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	var topicID, threadID int32
	for i := 0; i < len(parts)-1; i++ {
		switch parts[i] {
		case "topic":
			if v, err := strconv.ParseInt(parts[i+1], 10, 32); err == nil && topicID == 0 {
				topicID = int32(v)
			}
		case "thread":
			if v, err := strconv.ParseInt(parts[i+1], 10, 32); err == nil && threadID == 0 {
				threadID = int32(v)
			}
		}
	}
	return topicID, threadID
}

func normalizeForumThreadPath(rawPath string, private bool) string {
	cleaned := strings.TrimSpace(rawPath)
	cleaned = strings.Replace(cleaned, "/private/api/", "/private/", 1)
	cleaned = strings.TrimSuffix(cleaned, "/reply")
	if private && strings.HasPrefix(cleaned, "/forum/") {
		cleaned = "/private/" + strings.TrimPrefix(cleaned, "/forum/")
	}
	return cleaned
}

func int32FromAny(v any) (int32, bool) {
	switch val := v.(type) {
	case int32:
		return val, true
	case int64:
		return int32(val), true
	case int:
		return int32(val), true
	default:
		return 0, false
	}
}

func buildTaskEventFromForumReplyIntent(intent common.ForumReplyNotificationIntent) eventbus.TaskEvent {
	link := intent.AbsoluteCommentURL
	if link == "" {
		link = intent.CommentURL
	}
	if link == "" {
		link = intent.CanonicalPath
	}
	data := map[string]any{
		common.ForumReplyIntentEventKey: intent,
		postcountworker.EventKey: postcountworker.UpdateEventData{
			CommentID: intent.CommentID,
			ThreadID:  intent.ThreadID,
			TopicID:   intent.TopicID,
		},
		"Username":      intent.Username,
		"TopicTitle":    intent.TopicTitle,
		"Text":          intent.ThreadOpenerPreview,
		"CommentText":   intent.CommentText,
		"Body":          intent.CommentText,
		"CommentID":     intent.CommentID,
		"CommentAnchor": intent.CommentAnchor,
		"URL":           link,
		"SubjectPrefix": intent.SubjectPrefix,
		"Private":       intent.Private,
	}
	return eventbus.TaskEvent{
		Task:       forumReplyTaskName,
		Path:       intent.CanonicalPath,
		UserID:     intent.ActorID,
		ActivityID: intent.ActivityID,
		Data:       data,
		Outcome:    eventbus.TaskOutcomeSuccess,
	}
}

func (n *Notifier) forumReplyPatterns(intent common.ForumReplyNotificationIntent) []string {
	canonical := strings.TrimSpace(intent.CanonicalPath)
	if canonical == "" {
		return buildPatterns(forumReplyTaskName, "")
	}
	replyPath := strings.TrimRight(canonical, "/") + "/reply"
	merged := make([]string, 0, 8)
	seen := make(map[string]struct{}, 8)
	for _, p := range append(buildPatterns(forumReplyTaskName, replyPath), buildPatterns(forumReplyTaskName, canonical)...) {
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		merged = append(merged, p)
	}
	return merged
}

func (n *Notifier) forumReplyGrantRequirements(evt eventbus.TaskEvent, intent common.ForumReplyNotificationIntent) ([]GrantRequirement, error) {
	var reqs []GrantRequirement
	if intent.Private {
		if intent.TopicID == 0 || intent.ThreadID == 0 {
			return nil, fmt.Errorf("missing private forum topic (%d) or thread (%d) id for reply notification", intent.TopicID, intent.ThreadID)
		}
		reqs = append(reqs,
			GrantRequirement{
				Section: consts.PermissionSectionPrivateForum,
				Item:    consts.PermissionItemTopic,
				ItemID:  intent.TopicID,
				Action:  consts.PermissionActionView,
			},
			GrantRequirement{
				Section: consts.PermissionSectionPrivateForumThread,
				Item:    consts.PermissionItemThread,
				ItemID:  intent.ThreadID,
				Action:  consts.PermissionActionView,
			},
		)
	}
	if evt.Task != nil {
		if gp, ok := evt.Task.(GrantsRequiredProvider); ok {
			extra, err := gp.GrantsRequired(evt)
			if err != nil {
				return nil, err
			}
			for _, g := range extra {
				if !containsGrantRequirement(reqs, g) {
					reqs = append(reqs, g)
				}
			}
		}
	}
	return reqs, nil
}

func containsGrantRequirement(reqs []GrantRequirement, target GrantRequirement) bool {
	for _, r := range reqs {
		if r == target {
			return true
		}
	}
	return false
}

func (n *Notifier) userHasGrants(ctx context.Context, userID int32, reqs []GrantRequirement) bool {
	if len(reqs) == 0 {
		return true
	}
	if userID == 0 || n.Queries == nil {
		return false
	}
	for _, g := range reqs {
		if _, err := n.Queries.SystemCheckGrant(ctx, db.SystemCheckGrantParams{
			ViewerID:               userID,
			Section:                g.Section.String(),
			Item:                   sql.NullString{String: g.Item.String(), Valid: g.Item != ""},
			Action:                 g.Action.String(),
			ItemID:                 sql.NullInt32{Int32: g.ItemID, Valid: g.ItemID != 0},
			IsSpecificPrivateForum: (g.Section.String() == "privateforum" || g.Section.String() == "privateforum_thread") && g.ItemID != 0,
			UserID:                 sql.NullInt32{Int32: userID, Valid: userID != 0},
		}); err != nil {
			return false
		}
	}
	return true
}

func (n *Notifier) notifyForumReplyFromEvent(ctx context.Context, evt eventbus.TaskEvent, intent common.ForumReplyNotificationIntent, q dlq.DLQ) error {
	if n.Config == nil || !n.Config.NotificationsEnabled || n.Queries == nil {
		return nil
	}
	intent = n.normalizeForumReplyIntent(intent, &evt)

	// Hydrate missing display metadata from queries when invoked with a minimal intent.
	if intent.Username == "" && intent.ActorID != 0 {
		if u, err := n.Queries.SystemGetUserByID(ctx, intent.ActorID); err == nil && u != nil && u.Username.Valid {
			intent.Username = u.Username.String
		}
	}

	canonicalEvt := buildTaskEventFromForumReplyIntent(intent)
	if evt.Task != nil {
		canonicalEvt.Task = evt.Task
	}
	if evt.Data != nil {
		for k, v := range evt.Data {
			if _, exists := canonicalEvt.Data[k]; !exists {
				canonicalEvt.Data[k] = v
			}
		}
	}
	link := intent.AbsoluteCommentURL
	if link == "" {
		link = intent.CommentURL
	}
	if link == "" {
		link = intent.CanonicalPath
	}
	canonicalEvt.Data["URL"] = link
	canonicalEvt.Data[common.ForumReplyIntentEventKey] = intent

	patterns := n.forumReplyPatterns(intent)
	emailSubs, err := collectSubscribers(ctx, n.Queries, patterns, "email")
	if err != nil {
		return n.recordForumReplyDLQ(ctx, q, "notify subscribers", err, &canonicalEvt)
	}
	internalSubs, err := collectSubscribers(ctx, n.Queries, patterns, "internal")
	if err != nil {
		return n.recordForumReplyDLQ(ctx, q, "notify subscribers", err, &canonicalEvt)
	}

	delete(emailSubs, intent.ActorID)
	delete(internalSubs, intent.ActorID)

	reqs, err := n.forumReplyGrantRequirements(canonicalEvt, intent)
	if err != nil {
		return n.recordForumReplyDLQ(ctx, q, "notify subscribers", err, &canonicalEvt)
	}
	if len(reqs) != 0 {
		for id := range emailSubs {
			if !n.userHasGrants(ctx, id, reqs) {
				delete(emailSubs, id)
			}
		}
		for id := range internalSubs {
			if !n.userHasGrants(ctx, id, reqs) {
				delete(internalSubs, id)
			}
		}
	}

	internalTmpl := forumReplyNotificationTemplate.NotificationTemplate()
	sendInternal := true
	emailTmpl := forumReplyEmailTemplate.EmailTemplates()
	sendEmail := true
	if canonicalEvt.Task != nil {
		if tp, ok := canonicalEvt.Task.(SubscribersNotificationTemplateProvider); ok {
			if nt := tp.SubscribedInternalNotificationTemplate(canonicalEvt); nt != nil && *nt != "" {
				internalTmpl = *nt
			} else if nt == nil {
				sendInternal = false
			}
			et, send := tp.SubscribedEmailTemplate(canonicalEvt)
			emailTmpl = et
			sendEmail = send && et != nil
		}
	}

	var msg []byte
	if sendInternal && internalTmpl != "" && len(internalSubs) > 0 {
		data := EmailData{Item: canonicalEvt.Data, any: canonicalEvt.Data}
		msg, err = n.renderNotification(ctx, internalTmpl, data)
		if err != nil {
			log.Printf("render subscriber notification: %v", err)
			return n.recordForumReplyDLQ(ctx, q, "notify subscribers", fmt.Errorf("render notification: %w", err), &canonicalEvt)
		}
	}

	if sendEmail && emailTmpl != nil {
		for id := range emailSubs {
			if err := n.sendSubscriberEmail(ctx, id, canonicalEvt, emailTmpl); err != nil {
				return n.recordForumReplyDLQ(ctx, q, "notify subscribers", fmt.Errorf("deliver email to %d: %w", id, err), &canonicalEvt)
			}
		}
	}

	if len(msg) > 0 {
		for id := range internalSubs {
			if err := n.sendInternalNotificationForEvent(ctx, canonicalEvt, id, link, string(msg), notificationPrioritySubscriber); err != nil {
				return n.recordForumReplyDLQ(ctx, q, "notify subscribers", fmt.Errorf("deliver internal to %d: %w", id, err), &canonicalEvt)
			}
		}
	}

	if err := n.propagateForumReplyAutoSubscribers(ctx, patterns, intent, reqs); err != nil {
		return n.recordForumReplyDLQ(ctx, q, "notify subscribers", err, &canonicalEvt)
	}

	if err := n.ensureForumReplyActorAutoSubscription(ctx, intent); err != nil {
		return fmt.Errorf("AutoSubscribeProvider: %w", err)
	}

	return nil
}

func (n *Notifier) propagateForumReplyAutoSubscribers(ctx context.Context, patterns []string, intent common.ForumReplyNotificationIntent, reqs []GrantRequirement) error {
	if intent.CanonicalPath == "" {
		return nil
	}
	autoInternalSubs, err := collectSubscribers(ctx, n.Queries, patterns, "autosub_internal")
	if err != nil {
		return fmt.Errorf("collect auto internal subs: %w", err)
	}
	autoEmailSubs, err := collectSubscribers(ctx, n.Queries, patterns, "autosub_email")
	if err != nil {
		return fmt.Errorf("collect auto email subs: %w", err)
	}
	newPatterns := buildPatterns(forumReplyTaskName, intent.CanonicalPath)
	if len(newPatterns) == 0 {
		return nil
	}
	newPattern := newPatterns[0]
	for id := range autoInternalSubs {
		if id == intent.ActorID {
			continue
		}
		if n.userHasGrants(ctx, id, reqs) {
			if err := common.EnsureSubscriptionIdempotent(ctx, n.Queries, id, newPattern, "internal"); err != nil {
				return fmt.Errorf("ensure auto internal subscription: %w", err)
			}
		}
	}
	for id := range autoEmailSubs {
		if id == intent.ActorID {
			continue
		}
		if n.userHasGrants(ctx, id, reqs) {
			if err := common.EnsureSubscriptionIdempotent(ctx, n.Queries, id, newPattern, "email"); err != nil {
				return fmt.Errorf("ensure auto email subscription: %w", err)
			}
		}
	}
	return nil
}

func (n *Notifier) ensureForumReplyActorAutoSubscription(ctx context.Context, intent common.ForumReplyNotificationIntent) error {
	if intent.ActorID == 0 || intent.CanonicalPath == "" {
		return nil
	}
	auto, emailPref, err := common.CheckAutoSubscribePreference(ctx, n.Queries, intent.ActorID)
	if err != nil {
		if errors.Is(err, common.ErrPreferenceNotFound) {
			stats.IncrementAutoSubscribePreferenceFailures()
		} else {
			return fmt.Errorf("get preference by user_id: %w", err)
		}
	}
	if !auto {
		return nil
	}
	newPatterns := buildPatterns(forumReplyTaskName, intent.CanonicalPath)
	if len(newPatterns) == 0 {
		return nil
	}
	pattern := newPatterns[0]
	if n.Config != nil && n.Config.NotificationsEnabled {
		if err := common.EnsureSubscriptionIdempotent(ctx, n.Queries, intent.ActorID, pattern, "internal"); err != nil {
			return fmt.Errorf("ensure internal subscription: %w", err)
		}
	}
	if emailPref {
		if err := common.EnsureSubscriptionIdempotent(ctx, n.Queries, intent.ActorID, pattern, "email"); err != nil {
			return fmt.Errorf("ensure email subscription: %w", err)
		}
	}
	return nil
}

func (n *Notifier) recordForumReplyDLQ(ctx context.Context, q dlq.DLQ, prefix string, err error, evt *eventbus.TaskEvent) error {
	errW := fmt.Errorf("SubscribersNotificationTemplateProvider: %w", err)
	if dlqErr := n.dlqRecordAndNotify(ctx, q, fmt.Sprintf("%s: %v", prefix, errW), evt); dlqErr != nil {
		return dlqErr
	}
	return errW
}

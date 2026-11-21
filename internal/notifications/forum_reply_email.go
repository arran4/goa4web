package notifications

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
	"github.com/google/uuid"
)

const (
	// replyEmailLease allows another worker to resume a window if preparation stops.
	replyEmailLease = 2 * time.Minute
	// replyEmailThreadsPerMessage bounds the number of thread links in one summary.
	replyEmailThreadsPerMessage = 20
)

// WithReplyEmailClock supplies a clock shared by the reply email enqueue and worker.
func WithReplyEmailClock(now func() time.Time) Option {
	return func(n *Notifier) { n.replyEmailClock = now }
}

func (n *Notifier) replyEmailTime() time.Time {
	if n.replyEmailClock != nil {
		return n.replyEmailClock().UTC()
	}
	return time.Now().UTC()
}

func (n *Notifier) replyEmailInterval() time.Duration {
	if minutes := n.Config.ForumReplyEmailBatchMinutes; minutes > 0 && minutes <= int((1<<63-1)/time.Minute) {
		return time.Duration(minutes) * time.Minute
	}
	return 15 * time.Minute
}

func (n *Notifier) queueForumReplyEmail(ctx context.Context, recipientID int32, intent common.ForumReplyNotificationIntent) error {
	if intent.ActivityPosition <= 0 || intent.ThreadID <= 0 {
		return nil
	}
	interval := n.replyEmailInterval()
	windowEnd := n.replyEmailTime().Truncate(interval).Add(interval)
	return n.Queries.SystemQueueForumReplyEmailActivity(ctx, db.SystemQueueForumReplyEmailActivityParams{
		RecipientID: recipientID, ThreadID: intent.ThreadID,
		ActivityPosition: intent.ActivityPosition, WindowEnd: windowEnd.Unix(),
	})
}

type replyEmailThread struct {
	Label string
	Link  string
	Count int64
}

type replyEmailSummary struct {
	Threads     []replyEmailThread
	MoreThreads bool
}

// ProcessForumReplyEmailWindow sends at most one due optional reply summary.
// A durable lease covers preparation, and the send is fenced before calling the
// provider. A provider error can mean the remote server accepted the message, so
// automatic recovery only retries work that stopped before the send fence.
func (n *Notifier) ProcessForumReplyEmailWindow(ctx context.Context) error {
	if n.Config == nil || !n.Config.EmailEnabled || n.Queries == nil || n.EmailProvider == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	now := n.replyEmailTime()
	window, err := n.Queries.SystemGetDueForumReplyEmailWindow(ctx, db.SystemGetDueForumReplyEmailWindowParams{
		DueAt: now.Unix(), LeaseAvailableAt: now.Unix(),
		ThrottleAfter: sql.NullInt64{Int64: now.Add(-n.replyEmailInterval()).Unix(), Valid: true},
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("select due forum reply email window: %w", err)
	}
	token := sql.NullString{String: uuid.NewString(), Valid: true}
	claimed, err := n.Queries.SystemClaimForumReplyEmailWindow(ctx, db.SystemClaimForumReplyEmailWindowParams{
		RecipientID: window.RecipientID, WindowEnd: window.WindowEnd, LeaseToken: token,
		LeaseAvailableAt: now.Unix(), LeaseUntil: now.Add(replyEmailLease).Unix(),
	})
	if err != nil {
		return fmt.Errorf("claim forum reply email window: %w", err)
	}
	if claimed == 0 {
		return nil
	}

	user, err := n.Queries.SystemGetUserByID(ctx, window.RecipientID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("resolve forum reply email recipient: %w", err)
	}
	var address string
	var payload []byte
	summary, err := n.replyEmailSummary(ctx, window, token)
	if err != nil {
		return err
	}
	if user != nil && len(summary.Threads) != 0 {
		if user.Email.Valid && user.Email.String != "" {
			address = user.Email.String
			payload, err = n.RenderEmailFromTemplates(ctx, address, NewEmailTemplates("forumReplyBatch"), summary)
			if err != nil {
				return fmt.Errorf("render forum reply email summary: %w", err)
			}
		} else if n.Config.NotificationsEnabled {
			if err := notifyMissingEmail(ctx, n.Queries, window.RecipientID); err != nil {
				return fmt.Errorf("record missing forum reply email address: %w", err)
			}
		}
	}

	finishedAt := n.replyEmailTime().Unix()
	attemptedAt := int64(0)
	if len(payload) != 0 {
		attemptedAt = finishedAt
	}
	sealed, err := n.Queries.SystemSealForumReplyEmailWindow(ctx, db.SystemSealForumReplyEmailWindowParams{
		RecipientID: window.RecipientID, WindowEnd: window.WindowEnd, LeaseToken: token,
		AttemptedAt: sql.NullInt64{Int64: attemptedAt, Valid: true}, LeaseValidAt: finishedAt,
	})
	if err != nil {
		return fmt.Errorf("seal forum reply email window: %w", err)
	}
	if sealed == 0 || len(payload) == 0 {
		return nil
	}
	if err := n.EmailProvider.Send(ctx, email.ParseAddress(address), payload); err != nil {
		return fmt.Errorf("forum reply email provider reported an uncertain outcome for window ending at %d", window.WindowEnd)
	}
	if err := n.Queries.SystemMarkForumReplyEmailWindowSent(ctx, db.SystemMarkForumReplyEmailWindowSentParams{
		RecipientID: window.RecipientID, WindowEnd: window.WindowEnd, LeaseToken: token,
		SentAt: sql.NullInt64{Int64: n.replyEmailTime().Unix(), Valid: true},
	}); err != nil {
		return fmt.Errorf("record forum reply email provider acceptance: %w", err)
	}
	return nil
}

func (n *Notifier) replyEmailSummary(ctx context.Context, window *db.SystemGetDueForumReplyEmailWindowRow, token sql.NullString) (replyEmailSummary, error) {
	var summary replyEmailSummary
	subscriptions, err := n.Queries.ListSubscriptionsByUser(ctx, window.RecipientID)
	if err != nil {
		return summary, fmt.Errorf("check current forum email subscriptions: %w", err)
	}
	patterns := make(map[string]bool)
	for _, sub := range subscriptions {
		if sub.Method == "email" {
			patterns[sub.Pattern] = true
		}
	}
	if len(patterns) == 0 {
		return summary, nil
	}
	args := db.SystemListEligibleForumReplyEmailThreadsParams{
		RecipientID: window.RecipientID, WindowEnd: window.WindowEnd, LeaseToken: token, Limit: 50,
	}
	for {
		rows, err := n.Queries.SystemListEligibleForumReplyEmailThreads(ctx, args)
		if err != nil {
			return summary, fmt.Errorf("check current forum reply email access: %w", err)
		}
		for _, row := range rows {
			private := row.Handler == "private"
			path := fmt.Sprintf("%s/topic/%d/thread/%d", common.CanonicalForumBasePath(private, ""), row.TopicID, row.ThreadID)
			if !n.hasReplyEmailSubscription(path, patterns) {
				continue
			}
			if len(summary.Threads) == replyEmailThreadsPerMessage {
				summary.MoreThreads = true
				return summary, nil
			}
			label := "Private forum thread"
			if !private {
				label = strings.TrimSpace(row.Title.String)
				if label == "" {
					label = "Forum thread"
				}
			}
			summary.Threads = append(summary.Threads, replyEmailThread{
				Label: label, Link: strings.TrimRight(n.Config.BaseURL, "/") + path, Count: row.ReplyCount,
			})
		}
		if len(rows) < int(args.Limit) {
			return summary, nil
		}
		args.Offset += args.Limit
	}
}

func (n *Notifier) hasReplyEmailSubscription(path string, subscriptions map[string]bool) bool {
	for _, pattern := range expandPatternSeparators(n.forumReplyPatterns(common.ForumReplyNotificationIntent{CanonicalPath: path})) {
		if subscriptions[pattern] {
			return true
		}
	}
	return false
}

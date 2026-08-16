package forum

import (
	"context"
	"database/sql"
	"sort"
	"strings"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/db"
)

// ForkPreview is authorized, page-specific presentation data for one fork.
type ForkPreview struct {
	ThreadID         int32
	TopicID          int32
	FirstPoster      string
	FirstPosterID    int32
	FirstPostText    string
	FirstPostWritten time.Time
	LastPoster       string
	LastPosterID     int32
	LastAddition     time.Time
	CommentCount     int32
	IsUnread         bool
	Labels           []templates.TopicLabel
}

// ForkPreviewGroup contains the visible forks associated with one source comment.
type ForkPreviewGroup struct {
	Total   int
	Threads []*ForkPreview
}

// AuthorizedSourceReference contains only source relationship details the
// current viewer is authorized to see.
type AuthorizedSourceReference struct {
	ThreadID  int32
	TopicID   int32
	CommentID int32
}

func loadForksForViewer(ctx context.Context, cd *common.CoreData, sourceThreadID int32, perCommentLimit int) (map[int32]*ForkPreviewGroup, int, error) {
	rows, err := cd.Queries().GetReplyThreadsForLister(ctx, db.GetReplyThreadsForListerParams{
		ViewerID:        cd.UserID,
		ReplyToThreadID: sql.NullInt32{Int32: sourceThreadID, Valid: sourceThreadID != 0},
		ViewerMatchID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil {
		return nil, 0, err
	}
	if len(rows) == 0 {
		return map[int32]*ForkPreviewGroup{}, 0, nil
	}
	visibleSourceComments, err := cd.Queries().GetCommentsByThreadIdForUser(ctx, db.GetCommentsByThreadIdForUserParams{
		ViewerID: cd.UserID,
		ThreadID: sourceThreadID,
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	if err != nil {
		return nil, 0, err
	}
	visibleSourceCommentIDs := make(map[int32]struct{}, len(visibleSourceComments))
	for _, comment := range visibleSourceComments {
		visibleSourceCommentIDs[comment.Idcomments] = struct{}{}
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].IsUnread != rows[j].IsUnread {
			return rows[i].IsUnread > rows[j].IsUnread
		}
		return rows[i].Lastaddition.Time.After(rows[j].Lastaddition.Time)
	})

	groups := make(map[int32]*ForkPreviewGroup)
	for _, row := range rows {
		commentID := int32(0)
		if row.ReplyToCommentID.Valid {
			if _, visible := visibleSourceCommentIDs[row.ReplyToCommentID.Int32]; visible {
				commentID = row.ReplyToCommentID.Int32
			}
		}
		group := groups[commentID]
		if group == nil {
			group = &ForkPreviewGroup{}
			groups[commentID] = group
		}
		group.Total++
		if perCommentLimit > 0 && len(group.Threads) >= perCommentLimit {
			continue
		}
		preview := &ForkPreview{
			ThreadID:         row.Idforumthread,
			TopicID:          row.ForumtopicIdforumtopic,
			FirstPoster:      row.Firstpostusername.String,
			FirstPosterID:    row.Firstpostuserid.Int32,
			FirstPostText:    row.FirstPostText.String,
			FirstPostWritten: row.Firstpostwritten.Time,
			LastPoster:       row.LastPosterName.String,
			LastPosterID:     row.Lastposter,
			LastAddition:     row.Lastaddition.Time,
			CommentCount:     row.TotalComments.Int32,
			IsUnread:         row.IsUnread != 0,
		}
		if row.IsNew != 0 {
			preview.Labels = append(preview.Labels, templates.TopicLabel{Name: "new", Type: "private"})
		}
		if row.IsUnread != 0 {
			preview.Labels = append(preview.Labels, templates.TopicLabel{Name: "unread", Type: "private"})
		}
		appendForkLabels := func(labels sql.NullString, labelType string) {
			if !labels.Valid || labels.String == "" {
				return
			}
			for _, label := range strings.Split(labels.String, "\n") {
				preview.Labels = append(preview.Labels, templates.TopicLabel{Name: label, Type: labelType})
			}
		}
		appendForkLabels(row.PublicLabels, "public")
		appendForkLabels(row.AuthorLabels, "author")
		appendForkLabels(row.PrivateLabels, "private")
		group.Threads = append(group.Threads, preview)
	}
	return groups, len(rows), nil
}

func authorizedSourceReference(cd *common.CoreData, thread *db.GetThreadLastPosterAndPermsForUserRow) *AuthorizedSourceReference {
	if thread == nil || !thread.ReplyToThreadID.Valid {
		return nil
	}
	sourceThread, err := cd.ForumThreadByID(thread.ReplyToThreadID.Int32)
	if err != nil || sourceThread == nil {
		return nil
	}
	reference := &AuthorizedSourceReference{
		ThreadID: sourceThread.Idforumthread,
		TopicID:  sourceThread.ForumtopicIdforumtopic,
	}
	if !thread.ReplyToCommentID.Valid {
		return reference
	}
	sourceComment, err := cd.CommentByID(thread.ReplyToCommentID.Int32)
	if err == nil && sourceComment != nil && sourceComment.ForumthreadID == sourceThread.Idforumthread {
		reference.CommentID = sourceComment.Idcomments
	}
	return reference
}

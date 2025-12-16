package common

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/db"
)

const (
	// PrivateForumCategoryID identifies the hidden category for private topics.
	PrivateForumCategoryID int32 = 0
)

// PrivateTopic represents a private conversation with a computed title.
type PrivateTopic struct {
	*db.ListPrivateTopicsByUserIDRow
	DisplayTitle string
	Labels       []templates.TopicLabel
}

// PrivateForumTopics returns private forum topics visible to the current user.
func (cd *CoreData) PrivateForumTopics() ([]*PrivateTopic, error) {
	if cd == nil || cd.queries == nil {
		return nil, nil
	}
	if !cd.HasGrant("privateforum", "topic", "see", 0) {
		return nil, nil
	}
	return cd.privateForumTopics.Load(func() ([]*PrivateTopic, error) {
		tops, err := cd.queries.ListPrivateTopicsByUserID(cd.ctx, sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0})
		if err != nil {
			return nil, err
		}
		var pts []*PrivateTopic
		for _, t := range tops {
			parts, _ := cd.queries.ListPrivateTopicParticipantsByTopicIDForUser(cd.ctx, db.ListPrivateTopicParticipantsByTopicIDForUserParams{
				TopicID:  sql.NullInt32{Int32: t.Idforumtopic, Valid: true},
				ViewerID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
			})
			var names []string
			for _, p := range parts {
				if p.Idusers != cd.UserID {
					names = append(names, p.Username.String)
				}
			}
			title := strings.Join(names, ", ")
			if len(names) > 1 && t.Title.Valid && t.Title.String != "" {
				title = fmt.Sprintf("%s (%s)", title, t.Title.String)
			}
			var labels []templates.TopicLabel
			if pub, _, err := cd.ThreadPublicLabels(t.Idforumtopic); err == nil {
				for _, l := range pub {
					labels = append(labels, templates.TopicLabel{Name: l, Type: "public"})
				}
			} else {
				log.Printf("list public labels: %v", err)
			}
			pts = append(pts, &PrivateTopic{ListPrivateTopicsByUserIDRow: t, DisplayTitle: title, Labels: labels})
		}
		return pts, nil
	})
}

// PrivateTopics returns private forum topics or nil on error.
func (cd *CoreData) PrivateTopics() []*PrivateTopic {
	pts, _ := cd.PrivateForumTopics()
	return pts
}

func (cd *CoreData) GrantPrivateForumThread(ctx context.Context, newThreadID int32, userIDs []int32) error {
	for _, userID := range userIDs {
		for _, p := range []string{"see", "view", "post", "reply", "edit"} {
			if _, err := cd.queries.SystemCreateGrant(ctx, db.SystemCreateGrantParams{
				Section: "privateforum",
				Item:    sql.NullString{String: "thread", Valid: true},
				ItemID:  sql.NullInt32{Int32: newThreadID, Valid: true},
				Action:  p,
				UserID:  sql.NullInt32{Int32: userID, Valid: true},
			}); err != nil {
				return fmt.Errorf("granting see on thread %d to user %d: %w", newThreadID, userID, err)
			}
		}
	}
	return nil
}

// CreatePrivateTopicParams groups input for CreatePrivateTopic.
type CreatePrivateTopicParams struct {
	CreatorID      int32
	ParticipantIDs []int32
	Title          string
	Description    string
	PostBody       string
}

// CreatePrivateTopic creates a new private topic and assigns grants and the initial comment.
func (cd *CoreData) CreatePrivateTopic(p CreatePrivateTopicParams) (topicID int32, err error) {
	if cd == nil || cd.queries == nil {
		return 0, fmt.Errorf("no queries")
	}
	if !cd.HasGrant("privateforum", "topic", "create", 0) {
		log.Printf("private topic create denied: user=%d", p.CreatorID)
		return 0, fmt.Errorf("permission denied")
	}
	var usernames []string // TODO this should be fed in from the caller and if it is not provided we can fill it htis way
	for _, id := range p.ParticipantIDs {
		if u := cd.UserByID(id); u != nil {
			usernames = append(usernames, u.Username.String)
		} else {
			return 0, fmt.Errorf("unknown user %d", id)
		}
	}
	title := p.Title
	description := p.Description
	if title == "" {
		title = fmt.Sprintf("Private chat with %s", strings.Join(usernames, ", "))
		if description == "" {
			description = title
		}
	}
	tid, err := cd.queries.CreateForumTopicForPoster(cd.ctx, db.CreateForumTopicForPosterParams{
		PosterID:        p.CreatorID,
		ForumcategoryID: PrivateForumCategoryID,
		ForumLang:       sql.NullInt32{},
		Title:           sql.NullString{String: title, Valid: true},
		Description:     sql.NullString{String: description, Valid: true},
		Handler:         "private",
		Section:         "privateforum",
		GrantCategoryID: sql.NullInt32{Int32: PrivateForumCategoryID, Valid: true},
		GranteeID:       sql.NullInt32{Int32: p.CreatorID, Valid: p.CreatorID != 0},
	})
	if err != nil {
		return 0, fmt.Errorf("create topic %w", err)
	}
	if tid == 0 {
		return 0, fmt.Errorf("create topic returned 0")
	}
	topicID = int32(tid)
	for _, uid := range p.ParticipantIDs {
		for _, act := range []string{"see", "view", "post", "reply", "edit"} {
			if _, err := cd.queries.SystemCreateGrant(cd.ctx, db.SystemCreateGrantParams{ // TODO switch to cd.GrantForumTopic
				UserID:   sql.NullInt32{Int32: uid, Valid: true},
				RoleID:   sql.NullInt32{},
				Section:  "privateforum",
				Item:     sql.NullString{String: "topic", Valid: true},
				RuleType: "allow",
				ItemID:   sql.NullInt32{Int32: topicID, Valid: true},
				ItemRule: sql.NullString{},
				Action:   act,
				Extra:    sql.NullString{},
			}); err != nil {
				return 0, fmt.Errorf("create %s grant %w", act, err)
			}
		}
	}
	if p.PostBody == "" {
		return topicID, nil
	}
	thid, err := cd.queries.CreateForumThreadForPoster(cd.ctx, db.CreateForumThreadForPosterParams{
		ForumtopicID:  topicID,
		PosterID:      p.CreatorID,
		GrantParentID: sql.NullInt32{Int32: topicID, Valid: true},
		GranteeID:     sql.NullInt32{Int32: p.CreatorID, Valid: true},
	})
	if err != nil {
		return 0, fmt.Errorf("create thread: %w", err)
	}
	if thid == 0 {
		return 0, fmt.Errorf("create thread returned 0")
	}
	if err := cd.GrantPrivateForumThread(cd.ctx, int32(thid), p.ParticipantIDs); err != nil {
		return 0, fmt.Errorf("grant thread access: %w", err)
	}
	if _, err := cd.queries.CreateCommentInSectionForCommenter(cd.ctx, db.CreateCommentInSectionForCommenterParams{
		CommenterID:   sql.NullInt32{Int32: p.CreatorID, Valid: true},
		Section:       "privateforum",
		ItemType:      sql.NullString{String: "thread", Valid: true},
		ItemID:        sql.NullInt32{Int32: int32(thid), Valid: true},
		ForumthreadID: int32(thid),
		Text:          sql.NullString{String: p.PostBody, Valid: true},
		Written:       sql.NullTime{Time: time.Now(), Valid: true},
	}); err != nil {
		return 0, fmt.Errorf("create comment: %w", err)
	}
	return topicID, nil
}

package common

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

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

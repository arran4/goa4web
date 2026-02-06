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
	// PrivateTopicDefaultTitlePrefix is the prefix used for auto-generated private topic titles.
	PrivateTopicDefaultTitlePrefix = "Private chat with "
)

// GetPrivateTopicDisplayTitle returns the display title for a private topic.
// If the topic has a custom title (not starting with the default prefix), it is returned as is.
// Otherwise, it returns a comma-separated list of all participants.
func (cd *CoreData) GetPrivateTopicDisplayTitle(topicID int32, originalTitle string) string {
	if !strings.HasPrefix(originalTitle, PrivateTopicDefaultTitlePrefix) {
		return originalTitle
	}

	parts, err := cd.queries.AdminListPrivateTopicParticipantsByTopicID(cd.ctx, sql.NullInt32{Int32: topicID, Valid: true})
	if err != nil {
		log.Printf("list private participants: %v", err)
		return originalTitle
	}

	var names []string
	var allNames []string
	for _, p := range parts {
		if p.Username.Valid {
			allNames = append(allNames, p.Username.String)
			if cd.UserID == 0 || p.Idusers != cd.UserID {
				names = append(names, p.Username.String)
			}
		}
	}
	if len(names) == 0 {
		if len(allNames) > 0 {
			names = allNames
		} else {
			return originalTitle
		}
	}
	return strings.Join(names, ", ")
}

// GetPrivateTopicParticipants returns the list of participants (excluding viewer) for a private topic.
func (cd *CoreData) GetPrivateTopicParticipants(topicID int32) ([]string, error) {
	if cd.queries == nil {
		return nil, nil
	}
	parts, err := cd.queries.AdminListPrivateTopicParticipantsByTopicID(cd.ctx, sql.NullInt32{Int32: topicID, Valid: true})
	if err != nil {
		return nil, err
	}
	var names []string
	for _, p := range parts {
		if p.Username.Valid {
			if cd.UserID == 0 || p.Idusers != cd.UserID {
				names = append(names, p.Username.String)
			}
		}
	}
	return names, nil
}

// PrivateTopic represents a private conversation with a computed title.
type PrivateTopic struct {
	*db.ListPrivateTopicsByUserIDRow
	DisplayTitle string
	Labels       []templates.TopicLabel
	Participants []string
}

// PrivateForumTopics returns private forum topics visible to the current user.
func (cd *CoreData) PrivateForumTopics() ([]*PrivateTopic, error) {
	if cd == nil {
		return nil, nil
	}
	if !cd.HasGrant("privateforum", "topic", "see", 0) {
		return nil, nil
	}
	return cd.privateForumTopics.Load(func() ([]*PrivateTopic, error) {
		if cd.queries == nil {
			return nil, nil
		}
		tops, err := cd.queries.ListPrivateTopicsByUserID(cd.ctx, sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0})
		if err != nil {
			return nil, err
		}
		var pts []*PrivateTopic
		for _, t := range tops {
			title := t.Title.String
			var participants []string
			if t.Title.Valid {
				if strings.HasPrefix(t.Title.String, PrivateTopicDefaultTitlePrefix) {
					title = cd.GetPrivateTopicDisplayTitle(t.Idforumtopic, t.Title.String)
				} else {
					if p, err := cd.GetPrivateTopicParticipants(t.Idforumtopic); err == nil {
						participants = p
					} else {
						log.Printf("get participants: %v", err)
					}
				}
			}
			var labels []templates.TopicLabel

			rows, err := cd.queries.GetPrivateTopicThreadsAndLabels(cd.ctx, db.GetPrivateTopicThreadsAndLabelsParams{
				TopicID: t.Idforumtopic,
				UserID:  cd.UserID,
			})
			if err != nil {
				log.Printf("get topic threads and labels: %v", err)
			} else {
				type threadStatus struct {
					AuthorID int32
					Labels   map[string]bool
				}
				threads := make(map[int32]*threadStatus)
				for _, r := range rows {
					ts, ok := threads[r.Idforumthread]
					if !ok {
						ts = &threadStatus{
							AuthorID: r.AuthorID,
							Labels:   make(map[string]bool),
						}
						threads[r.Idforumthread] = ts
					}
					if r.Label.Valid {
						ts.Labels[r.Label.String] = r.Invert.Bool
					}
				}

				hasUnread := false
				hasNew := false

				for _, ts := range threads {
					// Check Unread: Exists unless explicitly marked read (invert=true)
					isRead := false
					if invert, ok := ts.Labels["unread"]; ok && invert {
						isRead = true
					}
					if !isRead {
						hasUnread = true
					}

					// Check New: Exists unless explicitly marked not new (invert=true) OR author is current user
					isNew := true
					if ts.AuthorID == cd.UserID {
						isNew = false
					} else if invert, ok := ts.Labels["new"]; ok && invert {
						isNew = false
					}

					if isNew {
						hasNew = true
					}

					if hasUnread && hasNew {
						break
					}
				}

				if hasUnread {
					labels = append(labels, templates.TopicLabel{Name: "unread", Type: "private"})
				}
				if hasNew {
					labels = append(labels, templates.TopicLabel{Name: "new", Type: "private"})
				}
			}

			if pub, owner, err := cd.PublicLabels("topic", t.Idforumtopic); err == nil {
				for _, l := range pub {
					labels = append(labels, templates.TopicLabel{Name: l, Type: "public"})
				}
				for _, l := range owner {
					labels = append(labels, templates.TopicLabel{Name: l, Type: "author"})
				}
			}

			if priv, err := cd.PrivateLabels("topic", t.Idforumtopic, 0); err == nil {
				for _, l := range priv {
					labels = append(labels, templates.TopicLabel{Name: l, Type: "private"})
				}
			}

			pts = append(pts, &PrivateTopic{ListPrivateTopicsByUserIDRow: t, DisplayTitle: title, Labels: labels, Participants: participants})
		}
		return pts, nil
	})
}

// PrivateTopics returns private forum topics or nil on error.
func (cd *CoreData) PrivateTopics() []*PrivateTopic {
	pts, _ := cd.PrivateForumTopics()
	return pts
}

// GrantPrivateForumTopic creates a grant for a private forum topic.
func (cd *CoreData) GrantPrivateForumTopic(topicID int32, uid, rid sql.NullInt32, action string) (int64, error) {
	if cd.queries == nil {
		return 0, fmt.Errorf("no queries")
	}
	return cd.queries.SystemCreateGrant(cd.ctx, db.SystemCreateGrantParams{
		UserID:   uid,
		RoleID:   rid,
		Section:  "privateforum",
		Item:     sql.NullString{String: "topic", Valid: true},
		RuleType: "allow",
		ItemID:   sql.NullInt32{Int32: topicID, Valid: true},
		ItemRule: sql.NullString{},
		Action:   action,
		Extra:    sql.NullString{},
	})
}

package common

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

<<<<<<< HEAD
	"github.com/arran4/goa4web/core/consts"
=======
>>>>>>> 585b27a2 (feat(forum): implement post appending within time window)
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/db"
)

const (
	// PrivateForumCategoryID identifies the hidden category for private topics.
	PrivateForumCategoryID int32 = 0
	// PrivateTopicDefaultTitlePrefix is the prefix used for auto-generated private topic titles.
	PrivateTopicDefaultTitlePrefix = "Private chat with "
)

// GetPrivateTopicDetails fetches the participant list from the database and returns the computed display title,
// the participant names (excluding the current user), and the total number of participants (including the current user).
func (cd *CoreData) GetPrivateTopicDetails(topicID int32, originalTitle string) (displayTitle string, participants []string, totalParticipants int, err error) {
	displayTitle = originalTitle
	if cd.queries == nil {
		return
	}
	parts, err := cd.queries.AdminListPrivateTopicParticipantsByTopicID(cd.ctx, sql.NullInt32{Int32: topicID, Valid: true})
	if err != nil {
		return
	}

	totalParticipants = len(parts)

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
	participants = names

	if strings.HasPrefix(originalTitle, PrivateTopicDefaultTitlePrefix) {
		if len(names) == 0 {
			if len(allNames) > 0 {
				names = allNames
			} else {
				displayTitle = originalTitle
				return
			}
		}
		displayTitle = strings.Join(names, ", ")
	}
	return
}

// GetPrivateTopicDisplayTitle returns the display title for a private topic.
// If the topic has a custom title (not starting with the default prefix), it is returned as is.
// Otherwise, it returns a comma-separated list of all participants.
func (cd *CoreData) GetPrivateTopicDisplayTitle(topicID int32, originalTitle string) string {
	displayTitle, _, _, err := cd.GetPrivateTopicDetails(topicID, originalTitle)
	if err != nil {
		log.Printf("list private participants: %v", err)
		return originalTitle
	}
	return displayTitle
}

// GetPrivateTopicParticipants returns the list of participants (excluding viewer) for a private topic.
func (cd *CoreData) GetPrivateTopicParticipants(topicID int32) ([]string, error) {
	_, participants, _, err := cd.GetPrivateTopicDetails(topicID, "")
	if err != nil {
		return nil, err
	}
	return participants, nil
}

// PrivateTopic represents a private conversation with a computed title.
type PrivateTopic struct {
	*db.ListPrivateTopicsByUserIDRow
	DisplayTitle       string
	Labels             []templates.TopicLabel
	Participants       []string
	ParticipantsString string
	TotalParticipants  int
}

// PrivateForumTopics returns private forum topics visible to the current user.
func (cd *CoreData) PrivateForumTopics() ([]*PrivateTopic, error) {
	if cd == nil {
		return nil, nil
	}
	if !cd.HasGrant("privateforum", "topic", "see", 0) {
		return nil, nil
	}
	return cd.cache.privateForumTopics.Load(func() ([]*PrivateTopic, error) {
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
			var totalParticipants int
			if t.Title.Valid {
				var err error
				title, participants, totalParticipants, err = cd.GetPrivateTopicDetails(t.Idforumtopic, t.Title.String)
				if err != nil {
					log.Printf("get private topic details: %v", err)
				}
			}
			var labels []templates.TopicLabel

<<<<<<< HEAD
			rows, err := cd.queries.GetPrivateTopicThreadsAndLabelsForUser(cd.ctx, db.GetPrivateTopicThreadsAndLabelsForUserParams{
				TopicID:       t.Idforumtopic,
				UserID:        cd.UserID,
				ViewerMatchID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
=======
			rows, err := cd.queries.GetPrivateTopicThreadsAndLabels(cd.ctx, db.GetPrivateTopicThreadsAndLabelsParams{
				TopicID: t.Idforumtopic,
				UserID:  cd.UserID,
>>>>>>> 585b27a2 (feat(forum): implement post appending within time window)
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

			pts = append(pts, &PrivateTopic{ListPrivateTopicsByUserIDRow: t, DisplayTitle: title, Labels: labels, Participants: participants, ParticipantsString: strings.Join(participants, ", "), TotalParticipants: totalParticipants})
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
<<<<<<< HEAD
		Section:  consts.PermissionSectionPrivateForum.String(),
		Item:     sql.NullString{String: consts.PermissionItemTopic.String(), Valid: true},
=======
		Section:  "privateforum",
		Item:     sql.NullString{String: "topic", Valid: true},
>>>>>>> 585b27a2 (feat(forum): implement post appending within time window)
		RuleType: "allow",
		ItemID:   sql.NullInt32{Int32: topicID, Valid: true},
		ItemRule: sql.NullString{},
		Action:   action,
		Extra:    sql.NullString{},
	})
}

// WithPrivateForumTopics preloads private forum topics for testing.
func WithPrivateForumTopics(topics []*PrivateTopic) CoreOption {
	return func(cd *CoreData) {
		cd.cache.privateForumTopics.Set(topics)
	}
}

// UnreadPrivateThreads fetches the unread private threads for the current user.
func (cd *CoreData) UnreadPrivateThreads(limit, offset int32, topicIDNull sql.NullInt32, topicIDVal int32) ([]*db.ListUnreadPrivateThreadsForUserRow, error) {
	if cd.queries == nil {
		return nil, nil
	}
	return cd.queries.ListUnreadPrivateThreadsForUser(cd.ctx, db.ListUnreadPrivateThreadsForUserParams{
		GranteeID:   cd.UserID,
		GrantUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		Limit:       limit,
		Offset:      offset,
	})
}

// UnreadPrivateThreadsCount fetches the total count of unread private threads for the current user.
func (cd *CoreData) UnreadPrivateThreadsCount(topicID int32) (int64, error) {
	if cd.queries == nil {
		return 0, nil
	}
	return cd.queries.CountUnreadPrivateThreadsForUser(cd.ctx, db.CountUnreadPrivateThreadsForUserParams{
		GranteeID:   cd.UserID,
		GrantUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
}

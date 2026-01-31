package user

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/subscriptions"
	"github.com/arran4/goa4web/internal/tasks"
)

func userSubscriptionsPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Subscriptions"

	dbSubs, err := cd.Queries().ListSubscriptionsByUser(r.Context(), cd.UserID)
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("list subscriptions: %w", err))
		return
	}

	groups := subscriptions.GetUserSubscriptions(dbSubs)

	var filteredGroups []*subscriptions.SubscriptionGroup
	for _, g := range groups {
		if g.Definition.IsAdminOnly && !cd.IsAdmin() {
			continue
		}
		// Also ensure default instance for param-less definitions
		if len(g.Instances) == 0 && !strings.Contains(g.Definition.Pattern, "{") {
			// Create empty instance
			g.Instances = append(g.Instances, &subscriptions.SubscriptionInstance{
				Original:   "", // Will use definition pattern
				Methods:    []string{},
				Parameters: []subscriptions.Parameter{},
			})
		}
		filteredGroups = append(filteredGroups, g)
	}

	for _, g := range filteredGroups {
		for _, inst := range g.Instances {
			for i, p := range inst.Parameters {
				if p.Resolved != "" {
					continue
				}
				if p.Key == "topicid" {
					if id, err := strconv.Atoi(p.Value); err == nil {
						topic, err := cd.Queries().GetForumTopicByIdForUser(r.Context(), db.GetForumTopicByIdForUserParams{
							ViewerID:      cd.UserID,
							Idforumtopic:  int32(id),
							ViewerMatchID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
						})
						if err == nil {
							title := topic.Title.String
							if topic.Handler == "private" {
								parts, _ := cd.Queries().ListPrivateTopicParticipantsByTopicIDForUser(r.Context(), db.ListPrivateTopicParticipantsByTopicIDForUserParams{
									TopicID:  sql.NullInt32{Int32: topic.Idforumtopic, Valid: true},
									ViewerID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
								})
								var names []string
								for _, p := range parts {
									if p.Idusers != cd.UserID {
										names = append(names, p.Username.String)
									}
								}
								namesTitle := strings.Join(names, ", ")
								if len(names) > 0 && (title == "" || strings.HasPrefix(title, "Private chat with")) {
									title = fmt.Sprintf("Private chat with %s", namesTitle)
								} else if len(names) > 0 {
									title = fmt.Sprintf("%s (%s)", namesTitle, title)
								}
								inst.Parameters[i].Link = fmt.Sprintf("/private/topic/%d", id)
							} else {
								inst.Parameters[i].Link = fmt.Sprintf("/forum/topic/%d", id)
							}
							inst.Parameters[i].Resolved = title
						}
					}
				} else if p.Key == "threadid" {
					if id, err := strconv.Atoi(p.Value); err == nil {
						// Ensure the user has permission to view the thread
						_, err := cd.Queries().GetThreadLastPosterAndPerms(r.Context(), db.GetThreadLastPosterAndPermsParams{
							ViewerID:      cd.UserID,
							ThreadID:      int32(id),
							ViewerMatchID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
						})
						if err == nil {
							thread, err := cd.Queries().AdminGetForumThreadById(r.Context(), int32(id))
							if err == nil {
								inst.Parameters[i].Resolved = thread.Title
								if thread.TopicHandler == "private" {
									inst.Parameters[i].Link = fmt.Sprintf("/private/topic/%d/thread/%d", thread.Idforumtopic, id)
								} else {
									inst.Parameters[i].Link = fmt.Sprintf("/forum/topic/%d/thread/%d", thread.Idforumtopic, id)
								}
							}
						}
					}
				}
			}
		}
	}

	data := struct {
		Groups []*subscriptions.SubscriptionGroup
	}{
		Groups: filteredGroups,
	}
	UserSubscriptionsPage.Handle(w, r, data)
}

const UserSubscriptionsPage tasks.Template = "user/subscriptions.gohtml"

package forum

import (
	"context"
	"sort"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

// TopicGrantGroup represents roles assigned to an action on a topic.
type TopicGrantGroup struct {
	Action    string
	Have      []string
	Disabled  []string
	Available []string
}

// buildTopicGrantGroups organises topic grants for editing.
func buildTopicGrantGroups(ctx context.Context, cd *common.CoreData, topicID int32) ([]TopicGrantGroup, error) {
	queries := cd.Queries()
	roles, err := cd.AllRoles()
	if err != nil {
		return nil, err
	}
	roleNames := map[int32]string{}
	for _, r := range roles {
		roleNames[r.ID] = r.Name
	}
	grants, err := queries.ListGrants(ctx)
	if err != nil {
		return nil, err
	}
	byAction := map[string]map[int32]*db.Grant{}
	for _, g := range grants {
		if g.Section != "forum" || !g.Item.Valid || g.Item.String != "topic" || !g.ItemID.Valid || g.ItemID.Int32 != topicID {
			continue
		}
		if g.UserID.Valid {
			continue // only role grants
		}
		rid := int32(0)
		if g.RoleID.Valid {
			rid = g.RoleID.Int32
		}
		if byAction[g.Action] == nil {
			byAction[g.Action] = map[int32]*db.Grant{}
		}
		byAction[g.Action][rid] = g
	}
	actions := []string{"see", "view", "reply", "post", "edit"}
	var groups []TopicGrantGroup
	for _, act := range actions {
		gg := TopicGrantGroup{Action: act}
		used := map[int32]bool{}
		if m, ok := byAction[act]; ok {
			for rid, g := range m {
				name, ok := roleNames[rid]
				if !ok {
					continue
				}
				used[rid] = true
				if g.Active {
					gg.Have = append(gg.Have, name)
				} else {
					gg.Disabled = append(gg.Disabled, name)
				}
			}
		}
		for rid, name := range roleNames {
			if !used[rid] {
				gg.Available = append(gg.Available, name)
			}
		}
		sort.Strings(gg.Have)
		sort.Strings(gg.Disabled)
		sort.Strings(gg.Available)
		groups = append(groups, gg)
	}
	return groups, nil
}

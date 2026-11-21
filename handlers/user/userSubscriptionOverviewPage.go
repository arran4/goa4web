package user

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/subscriptions"
	"github.com/arran4/goa4web/internal/tasks"
)

// UserSubscriptionOverviewPage renders saved follows and links to their settings.
const UserSubscriptionOverviewPage tasks.Template = "domains/user/subscriptionOverview.gohtml"

const (
	// followNewThreads labels a rule for new thread notifications.
	followNewThreads = "New threads"
	// followReplies labels a rule for reply notifications.
	followReplies = "Replies"
	// followFutureThreadReplies labels an automatic follow on newly created threads.
	followFutureThreadReplies = "Automatically follow replies to new threads"
)

type forumFollowScope struct {
	base     string
	topicID  int32
	threadID int32
}

type forumFollowChannels struct {
	Activity string
	InSite   bool
	Email    bool
	Paused   bool
}

type forumFollowTopic struct {
	Link, Title string
	Rules       []forumFollowChannels
	Threads     []forumFollowThread
}

type forumFollowThread struct {
	Link, Title string
	Rules       []forumFollowChannels
}

type forumFollowModule struct {
	Link, Title string
	Rules       []forumFollowChannels
	Topics      []forumFollowTopic
}

func userSubscriptionOverviewPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if !cd.CanLogIn() {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	switch r.URL.Query().Get("mode") {
	case "advanced", "admin", "user":
		userSubscriptionsPage(w, r)
		return
	}
	cd.PageTitle = "Subscriptions"
	rows, err := cd.UserSubscriptions()
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("list subscriptions: %w", err))
		return
	}
	pref, err := cd.Preference()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		handlers.RenderErrorPage(w, r, fmt.Errorf("load subscription preferences: %w", err))
		return
	}
	autoFollow := pref == nil || pref.AutoSubscribeReplies
	replyEmail := pref != nil && pref.Emailforumupdates.Valid && pref.Emailforumupdates.Bool
	modules, err := visibleForumFollows(cd, rows, replyEmail)
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("load followed forum scopes: %w", err))
		return
	}
	admin := cd.IsAdmin()
	data := struct {
		Modules    []forumFollowModule
		Other      []forumFollowChannels
		AutoFollow bool
		ReplyEmail bool
		Admin      bool
	}{Modules: modules, Other: otherSavedSubscriptions(rows, admin), AutoFollow: autoFollow, ReplyEmail: replyEmail, Admin: admin}
	_ = UserSubscriptionOverviewPage.Handle(w, r, data)
}

func otherSavedSubscriptions(rows []*db.ListSubscriptionsByUserRow, admin bool) []forumFollowChannels {
	byName := make(map[string]forumFollowChannels)
	for _, row := range rows {
		if _, _, shown := parseForumFollowRule(row.Pattern); shown || (row.Method != "internal" && row.Method != "email") {
			continue
		}
		def, params := subscriptions.MatchDefinition(row.Pattern)
		if def == nil || def.Mandatory || def.Legacy || len(params) != 0 || (def.IsAdminOnly && !admin) {
			continue
		}
		channels := byName[def.Name]
		channels.Activity = def.Name
		channels.InSite = channels.InSite || row.Method == "internal"
		channels.Email = channels.Email || row.Method == "email"
		byName[def.Name] = channels
	}
	result := make([]forumFollowChannels, 0, len(byName))
	for _, channels := range byName {
		result = append(result, channels)
	}
	slices.SortFunc(result, func(a, b forumFollowChannels) int { return strings.Compare(a.Activity, b.Activity) })
	return result
}

func visibleForumFollows(cd *common.CoreData, rows []*db.ListSubscriptionsByUserRow, replyEmail bool) ([]forumFollowModule, error) {
	rules := forumFollowRules(rows, replyEmail)
	var result []forumFollowModule
	for _, module := range []forumFollowModule{{Link: "/forum", Title: "Public forum"}, {Link: "/private", Title: "Private forum"}} {
		module.Rules = rules[forumFollowScope{base: module.Link}]
		byTopic := make(map[int32][]int32)
		for scope := range rules {
			if scope.base != module.Link || scope.topicID == 0 {
				continue
			}
			if scope.threadID != 0 {
				byTopic[scope.topicID] = append(byTopic[scope.topicID], scope.threadID)
			} else if _, exists := byTopic[scope.topicID]; !exists {
				byTopic[scope.topicID] = nil
			}
		}
		var topicIDs []int32
		for topicID := range byTopic {
			topicIDs = append(topicIDs, topicID)
		}
		slices.Sort(topicIDs)
		for _, topicID := range topicIDs {
			topic, err := visibleFollowedTopic(cd, module.Link, topicID, byTopic[topicID], rules)
			if err != nil {
				return nil, err
			}
			if topic != nil {
				module.Topics = append(module.Topics, *topic)
			}
		}
		if len(module.Rules) != 0 || len(module.Topics) != 0 {
			result = append(result, module)
		}
	}
	return result, nil
}

func visibleFollowedTopic(cd *common.CoreData, base string, topicID int32, threadIDs []int32, rules map[forumFollowScope][]forumFollowChannels) (*forumFollowTopic, error) {
	record, err := cd.ForumTopicByID(topicID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if record == nil || record.DeletedAt.Valid || (record.Handler == "private") != (base == "/private") {
		return nil, nil
	}
	topic := &forumFollowTopic{
		Link:  fmt.Sprintf("%s/topic/%d", base, topicID),
		Rules: rules[forumFollowScope{base: base, topicID: topicID}],
	}
	slices.Sort(threadIDs)
	for _, threadID := range threadIDs {
		thread, err := cd.ForumThreadByID(threadID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		if thread == nil || thread.ForumtopicIdforumtopic != topicID || thread.DeletedAt.Valid {
			continue
		}
		opener, err := cd.CommentByID(thread.Firstpost)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		if opener == nil || opener.DeletedAt.Valid || opener.ForumthreadID != threadID {
			continue
		}
		title := a4code.SnipTextWords(opener.Text.String, 10)
		if title == "" {
			title = fmt.Sprintf("Thread %d", threadID)
		}
		topic.Threads = append(topic.Threads, forumFollowThread{
			Link: fmt.Sprintf("%s/thread/%d", topic.Link, threadID), Title: title,
			Rules: rules[forumFollowScope{base: base, topicID: topicID, threadID: threadID}],
		})
	}
	if len(topic.Rules) == 0 && len(topic.Threads) == 0 {
		return nil, nil
	}
	topic.Title = strings.TrimSpace(record.Title.String)
	if base == "/private" {
		topic.Title = cd.GetPrivateTopicDisplayTitle(topicID, topic.Title)
	}
	if topic.Title == "" {
		topic.Title = fmt.Sprintf("Topic %d", topicID)
	}
	return topic, nil
}

func forumFollowRules(rows []*db.ListSubscriptionsByUserRow, replyEmail bool) map[forumFollowScope][]forumFollowChannels {
	byScope := make(map[forumFollowScope]map[string]forumFollowChannels)
	for _, row := range rows {
		scope, activity, ok := parseForumFollowRule(row.Pattern)
		if !ok {
			continue
		}
		method := row.Method
		if suffix, auto := strings.CutPrefix(method, "autosub_"); auto && activity == followNewThreads {
			activity, method = followFutureThreadReplies, suffix
		}
		if method != "internal" && method != "email" {
			continue
		}
		if byScope[scope] == nil {
			byScope[scope] = make(map[string]forumFollowChannels)
		}
		channels := byScope[scope][activity]
		channels.Activity = activity
		channels.InSite = channels.InSite || method == "internal"
		channels.Email = channels.Email || method == "email"
		channels.Paused = channels.Email && activity == followReplies && !replyEmail
		byScope[scope][activity] = channels
	}
	result := make(map[forumFollowScope][]forumFollowChannels, len(byScope))
	for scope, rules := range byScope {
		for _, name := range []string{followNewThreads, followReplies, followFutureThreadReplies} {
			if channels, ok := rules[name]; ok {
				result[scope] = append(result[scope], channels)
			}
		}
	}
	return result
}

func parseForumFollowRule(pattern string) (forumFollowScope, string, bool) {
	action, path, _ := strings.Cut(pattern, ":")
	var activity string
	switch action {
	case "create thread":
		activity = followNewThreads
	case "reply":
		activity = followReplies
	default:
		return forumFollowScope{}, "", false
	}
	path = strings.TrimSuffix(strings.TrimSuffix(path, "/*"), "/")
	path, ok := strings.CutPrefix(path, "/")
	parts := strings.Split(path, "/")
	if !ok || (parts[0] != "forum" && parts[0] != "private") {
		return forumFollowScope{}, "", false
	}
	scope := forumFollowScope{base: "/" + parts[0]}
	if len(parts) == 1 || (len(parts) == 2 && parts[1] == "topic") {
		return scope, activity, true
	}
	if len(parts) < 3 || parts[1] != "topic" {
		return forumFollowScope{}, "", false
	}
	topicID, err := strconv.ParseInt(parts[2], 10, 32)
	if err != nil || topicID <= 0 || strconv.FormatInt(topicID, 10) != parts[2] {
		return forumFollowScope{}, "", false
	}
	scope.topicID = int32(topicID)
	if len(parts) == 3 || (len(parts) == 4 && parts[3] == "thread") {
		return scope, activity, true
	}
	if activity != followReplies || len(parts) != 5 || parts[3] != "thread" {
		return forumFollowScope{}, "", false
	}
	threadID, err := strconv.ParseInt(parts[4], 10, 32)
	if err != nil || threadID <= 0 || strconv.FormatInt(threadID, 10) != parts[4] {
		return forumFollowScope{}, "", false
	}
	scope.threadID = int32(threadID)
	return scope, activity, true
}

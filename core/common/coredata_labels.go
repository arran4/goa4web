package common

import (
	"sort"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/db"
)

// PublicLabels returns public and owner labels for an item sorted alphabetically.
func (cd *CoreData) PublicLabels(item string, itemID int32) (public, owner []string, err error) {
	if cd.queries == nil {
		return nil, nil, nil
	}
	rows, err := cd.queries.ListContentPublicLabels(cd.ctx, db.ListContentPublicLabelsParams{Item: item, ItemID: itemID})
	if err != nil {
		return nil, nil, err
	}
	for _, r := range rows {
		public = append(public, r.Label)
	}
	sort.Strings(public)

	ownerRows, err := cd.queries.ListContentLabelStatus(cd.ctx, db.ListContentLabelStatusParams{Item: item, ItemID: itemID})
	if err != nil {
		return nil, nil, err
	}
	for _, r := range ownerRows {
		owner = append(owner, r.Label)
	}
	sort.Strings(owner)
	return public, owner, nil
}

// AddPublicLabel adds a public label to an item.
func (cd *CoreData) AddPublicLabel(item string, itemID int32, label string) error {
	if cd.queries == nil {
		return nil
	}
	return cd.queries.AddContentPublicLabel(cd.ctx, db.AddContentPublicLabelParams{Item: item, ItemID: itemID, Label: label})
}

// RemovePublicLabel removes a public label from an item.
func (cd *CoreData) RemovePublicLabel(item string, itemID int32, label string) error {
	if cd.queries == nil {
		return nil
	}
	return cd.queries.RemoveContentPublicLabel(cd.ctx, db.RemoveContentPublicLabelParams{Item: item, ItemID: itemID, Label: label})
}

// AddAuthorLabel adds an owner-only label to an item.
func (cd *CoreData) AddAuthorLabel(item string, itemID int32, label string) error {
	if cd.queries == nil {
		return nil
	}
	return cd.queries.AddContentLabelStatus(cd.ctx, db.AddContentLabelStatusParams{Item: item, ItemID: itemID, Label: label})
}

// RemoveAuthorLabel removes an owner-only label from an item.
func (cd *CoreData) RemoveAuthorLabel(item string, itemID int32, label string) error {
	if cd.queries == nil {
		return nil
	}
	return cd.queries.RemoveContentLabelStatus(cd.ctx, db.RemoveContentLabelStatusParams{Item: item, ItemID: itemID, Label: label})
}

// SetAuthorLabels replaces all author-only labels on an item with the provided list.
func (cd *CoreData) SetAuthorLabels(item string, itemID int32, labels []string) error {
	if cd.queries == nil {
		return nil
	}
	_, current, err := cd.PublicLabels(item, itemID)
	if err != nil {
		return err
	}
	have := make(map[string]struct{}, len(current))
	for _, l := range current {
		have[l] = struct{}{}
	}
	want := make(map[string]struct{}, len(labels))
	for _, l := range labels {
		want[l] = struct{}{}
	}
	for l := range want {
		if _, ok := have[l]; !ok {
			if err := cd.AddAuthorLabel(item, itemID, l); err != nil {
				return err
			}
		}
	}
	for l := range have {
		if _, ok := want[l]; !ok {
			if err := cd.RemoveAuthorLabel(item, itemID, l); err != nil {
				return err
			}
		}
	}
	return nil
}

// SetPublicLabels replaces all public labels on an item with the provided list.
func (cd *CoreData) SetPublicLabels(item string, itemID int32, labels []string) error {
	if cd.queries == nil {
		return nil
	}
	current, _, err := cd.PublicLabels(item, itemID)
	if err != nil {
		return err
	}
	have := make(map[string]struct{}, len(current))
	for _, l := range current {
		have[l] = struct{}{}
	}
	want := make(map[string]struct{}, len(labels))
	for _, l := range labels {
		want[l] = struct{}{}
	}
	for l := range want {
		if _, ok := have[l]; !ok {
			if err := cd.AddPublicLabel(item, itemID, l); err != nil {
				return err
			}
		}
	}
	for l := range have {
		if _, ok := want[l]; !ok {
			if err := cd.RemovePublicLabel(item, itemID, l); err != nil {
				return err
			}
		}
	}
	return nil
}

// PrivateLabels returns private labels for an item sorted alphabetically.
func (cd *CoreData) PrivateLabels(item string, itemID int32, authorID int32) ([]string, error) {
	if cd.queries == nil {
		return nil, nil
	}
	rows, err := cd.queries.ListContentPrivateLabels(cd.ctx, db.ListContentPrivateLabelsParams{Item: item, ItemID: itemID, UserID: cd.UserID})
	if err != nil {
		return nil, err
	}
	var (
		userLabels []string
		inverted   = make(map[string]bool)
	)
	for _, r := range rows {
		if r.Invert {
			inverted[r.Label] = true
			continue
		}
		userLabels = append(userLabels, r.Label)
	}
	sort.Strings(userLabels)
	labelSet := make(map[string]struct{}, len(userLabels))
	for _, label := range userLabels {
		labelSet[label] = struct{}{}
	}
	labels := make([]string, 0, len(userLabels)+2)
	// Only threads, news articles, links, image board posts, blog entries,
	// and writing articles receive default status labels.
	if cd.UserID != 0 {
		switch item {
		case "thread", "news", "link", "imagebbs", "blog", "writing":
			if !inverted["new"] && authorID != cd.UserID {
				if _, ok := labelSet["new"]; !ok {
					labels = append(labels, "new")
				}
			}
			if !inverted["unread"] {
				if _, ok := labelSet["unread"]; !ok {
					labels = append(labels, "unread")
				}
			}
		}
	}
	labels = append(labels, userLabels...)
	return labels, nil
}

// ClearPrivateLabelStatus removes stored unread inversions for an item across all users.
func (cd *CoreData) ClearPrivateLabelStatus(item string, itemID int32) error {
	if cd.queries == nil {
		return nil
	}
	return cd.queries.SystemClearContentPrivateLabel(cd.ctx, db.SystemClearContentPrivateLabelParams{Item: item, ItemID: itemID, Label: "unread"})
}

// ClearUnreadForOthers removes the inverted unread label for an item from all
// users except the current one.
func (cd *CoreData) ClearUnreadForOthers(item string, itemID int32) error {
	if cd.queries == nil {
		return nil
	}
	return cd.queries.ClearUnreadContentPrivateLabelExceptUser(cd.ctx, db.ClearUnreadContentPrivateLabelExceptUserParams{Item: item, ItemID: itemID, UserID: cd.UserID})
}

// SetPrivateLabelStatus updates the special new/unread flags for an item.
func (cd *CoreData) SetPrivateLabelStatus(item string, itemID int32, newLabel, unreadLabel bool) error {
	if cd.queries == nil {
		return nil
	}
	if newLabel {
		if err := cd.queries.RemoveContentPrivateLabel(cd.ctx, db.RemoveContentPrivateLabelParams{Item: item, ItemID: itemID, UserID: cd.UserID, Label: "new"}); err != nil {
			return err
		}
	} else {
		if err := cd.queries.AddContentPrivateLabel(cd.ctx, db.AddContentPrivateLabelParams{Item: item, ItemID: itemID, UserID: cd.UserID, Label: "new", Invert: true}); err != nil {
			return err
		}
	}
	if unreadLabel {
		if err := cd.queries.RemoveContentPrivateLabel(cd.ctx, db.RemoveContentPrivateLabelParams{Item: item, ItemID: itemID, UserID: cd.UserID, Label: "unread"}); err != nil {
			return err
		}
	} else {
		if err := cd.queries.AddContentPrivateLabel(cd.ctx, db.AddContentPrivateLabelParams{Item: item, ItemID: itemID, UserID: cd.UserID, Label: "unread", Invert: true}); err != nil {
			return err
		}
	}
	return nil
}

// AddPrivateLabel adds a private label to an item for the current user.
func (cd *CoreData) AddPrivateLabel(item string, itemID int32, label string) error {
	if cd.queries == nil {
		return nil
	}
	return cd.queries.AddContentPrivateLabel(cd.ctx, db.AddContentPrivateLabelParams{Item: item, ItemID: itemID, UserID: cd.UserID, Label: label, Invert: false})
}

// RemovePrivateLabel removes a private label from an item for the current user.
func (cd *CoreData) RemovePrivateLabel(item string, itemID int32, label string) error {
	if cd.queries == nil {
		return nil
	}
	return cd.queries.RemoveContentPrivateLabel(cd.ctx, db.RemoveContentPrivateLabelParams{Item: item, ItemID: itemID, UserID: cd.UserID, Label: label})
}

// SetPrivateLabels replaces all private labels for the current user on an item with the provided list.
func (cd *CoreData) SetPrivateLabels(item string, itemID int32, labels []string) error {
	if cd.queries == nil {
		return nil
	}
	rows, err := cd.queries.ListContentPrivateLabels(cd.ctx, db.ListContentPrivateLabelsParams{Item: item, ItemID: itemID, UserID: cd.UserID})
	if err != nil {
		return err
	}
	have := make(map[string]struct{}, len(rows))
	for _, r := range rows {
		if r.Invert || r.Label == "new" || r.Label == "unread" {
			continue
		}
		have[r.Label] = struct{}{}
	}
	want := make(map[string]struct{}, len(labels))
	for _, l := range labels {
		want[l] = struct{}{}
	}
	for l := range want {
		if _, ok := have[l]; !ok {
			if err := cd.AddPrivateLabel(item, itemID, l); err != nil {
				return err
			}
		}
	}
	for l := range have {
		if _, ok := want[l]; !ok {
			if err := cd.RemovePrivateLabel(item, itemID, l); err != nil {
				return err
			}
		}
	}
	return nil
}

// Convenience wrappers for common items.

func (cd *CoreData) TopicPublicLabels(topicID int32) (public, owner []string, err error) {
	return cd.PublicLabels("topic", topicID)
}

func (cd *CoreData) AddTopicPublicLabel(topicID int32, label string) error {
	return cd.AddPublicLabel("topic", topicID, label)
}

func (cd *CoreData) RemoveTopicPublicLabel(topicID int32, label string) error {
	return cd.RemovePublicLabel("topic", topicID, label)
}

func (cd *CoreData) SetTopicPublicLabels(topicID int32, labels []string) error {
	return cd.SetPublicLabels("topic", topicID, labels)
}

func (cd *CoreData) ThreadPublicLabels(threadID int32) (public, owner []string, err error) {
	return cd.PublicLabels("thread", threadID)
}

func (cd *CoreData) AddThreadPublicLabel(threadID int32, label string) error {
	return cd.AddPublicLabel("thread", threadID, label)
}

func (cd *CoreData) RemoveThreadPublicLabel(threadID int32, label string) error {
	return cd.RemovePublicLabel("thread", threadID, label)
}

func (cd *CoreData) AddThreadAuthorLabel(threadID int32, label string) error {
	return cd.AddAuthorLabel("thread", threadID, label)
}

func (cd *CoreData) RemoveThreadAuthorLabel(threadID int32, label string) error {
	return cd.RemoveAuthorLabel("thread", threadID, label)
}

func (cd *CoreData) SetThreadAuthorLabels(threadID int32, labels []string) error {
	return cd.SetAuthorLabels("thread", threadID, labels)
}

func (cd *CoreData) SetThreadPublicLabels(threadID int32, labels []string) error {
	return cd.SetPublicLabels("thread", threadID, labels)
}

func (cd *CoreData) ThreadPrivateLabels(threadID int32, authorID int32) ([]string, error) {
	return cd.PrivateLabels("thread", threadID, authorID)
}

func (cd *CoreData) ClearThreadPrivateLabelStatus(threadID int32) error {
	return cd.ClearPrivateLabelStatus("thread", threadID)
}

func (cd *CoreData) ClearThreadUnreadForOthers(threadID int32) error {
	return cd.ClearUnreadForOthers("thread", threadID)
}

func (cd *CoreData) SetThreadPrivateLabelStatus(threadID int32, newLabel, unreadLabel bool) error {
	return cd.SetPrivateLabelStatus("thread", threadID, newLabel, unreadLabel)
}

func (cd *CoreData) AddThreadPrivateLabel(threadID int32, label string) error {
	return cd.AddPrivateLabel("thread", threadID, label)
}

func (cd *CoreData) RemoveThreadPrivateLabel(threadID int32, label string) error {
	return cd.RemovePrivateLabel("thread", threadID, label)
}

func (cd *CoreData) SetThreadPrivateLabels(threadID int32, labels []string) error {
	return cd.SetPrivateLabels("thread", threadID, labels)
}

// Writings

// WritingAuthorLabels returns author labels for a writing.
func (cd *CoreData) WritingAuthorLabels(writingID int32) ([]string, error) {
	_, owner, err := cd.PublicLabels("writing", writingID)
	return owner, err
}

// AddWritingAuthorLabel adds an author-only label to a writing.
func (cd *CoreData) AddWritingAuthorLabel(writingID int32, label string) error {
	return cd.AddAuthorLabel("writing", writingID, label)
}

// RemoveWritingAuthorLabel removes an author-only label from a writing.
func (cd *CoreData) RemoveWritingAuthorLabel(writingID int32, label string) error {
	return cd.RemoveAuthorLabel("writing", writingID, label)
}

// SetWritingAuthorLabels replaces all author labels on a writing with the provided list.
func (cd *CoreData) SetWritingAuthorLabels(writingID int32, labels []string) error {
	return cd.SetAuthorLabels("writing", writingID, labels)
}

func (cd *CoreData) WritingPrivateLabels(writingID int32, authorID int32) ([]string, error) {
	return cd.PrivateLabels("writing", writingID, authorID)
}

func (cd *CoreData) SetWritingPrivateLabels(writingID int32, labels []string) error {
	return cd.SetPrivateLabels("writing", writingID, labels)
}

func (cd *CoreData) ClearWritingUnreadForOthers(writingID int32) error {
	return cd.ClearUnreadForOthers("writing", writingID)
}

// WritingLabels returns author and private labels for a writing.
func (cd *CoreData) WritingLabels(writingID int32, authorID int32) []templates.TopicLabel {
	var labels []templates.TopicLabel
	if als, err := cd.WritingAuthorLabels(writingID); err == nil {
		for _, l := range als {
			labels = append(labels, templates.TopicLabel{Name: l, Type: "author"})
		}
	}
	if pls, err := cd.WritingPrivateLabels(writingID, authorID); err == nil {
		for _, l := range pls {
			labels = append(labels, templates.TopicLabel{Name: l, Type: "private"})
		}
	}
	return labels
}

// News

// NewsAuthorLabels returns author labels for a news item.
func (cd *CoreData) NewsAuthorLabels(newsID int32) ([]string, error) {
	_, owner, err := cd.PublicLabels("news", newsID)
	return owner, err
}

// AddNewsAuthorLabel adds an author-only label to a news item.
func (cd *CoreData) AddNewsAuthorLabel(newsID int32, label string) error {
	return cd.AddAuthorLabel("news", newsID, label)
}

// RemoveNewsAuthorLabel removes an author-only label from a news item.
func (cd *CoreData) RemoveNewsAuthorLabel(newsID int32, label string) error {
	return cd.RemoveAuthorLabel("news", newsID, label)
}

// SetNewsAuthorLabels replaces all author labels on a news item with the provided list.
func (cd *CoreData) SetNewsAuthorLabels(newsID int32, labels []string) error {
	return cd.SetAuthorLabels("news", newsID, labels)
}

func (cd *CoreData) NewsPrivateLabels(newsID int32, authorID int32) ([]string, error) {
	return cd.PrivateLabels("news", newsID, authorID)
}

func (cd *CoreData) SetNewsPrivateLabels(newsID int32, labels []string) error {
	return cd.SetPrivateLabels("news", newsID, labels)
}

// NewsLabels returns author and private labels for a news item.
func (cd *CoreData) NewsLabels(newsID int32, authorID int32) []templates.TopicLabel {
	var labels []templates.TopicLabel
	if als, err := cd.NewsAuthorLabels(newsID); err == nil {
		for _, l := range als {
			labels = append(labels, templates.TopicLabel{Name: l, Type: "author"})
		}
	}
	if pls, err := cd.NewsPrivateLabels(newsID, authorID); err == nil {
		for _, l := range pls {
			labels = append(labels, templates.TopicLabel{Name: l, Type: "private"})
		}
	}
	return labels
}

// Blogs

// BlogAuthorLabels returns author labels for a blog post.
func (cd *CoreData) BlogAuthorLabels(blogID int32) ([]string, error) {
	_, owner, err := cd.PublicLabels("blog", blogID)
	return owner, err
}

// AddBlogAuthorLabel adds an author-only label to a blog post.
func (cd *CoreData) AddBlogAuthorLabel(blogID int32, label string) error {
	return cd.AddAuthorLabel("blog", blogID, label)
}

// RemoveBlogAuthorLabel removes an author-only label from a blog post.
func (cd *CoreData) RemoveBlogAuthorLabel(blogID int32, label string) error {
	return cd.RemoveAuthorLabel("blog", blogID, label)
}

// SetBlogAuthorLabels replaces all author labels on a blog post with the provided list.
func (cd *CoreData) SetBlogAuthorLabels(blogID int32, labels []string) error {
	return cd.SetAuthorLabels("blog", blogID, labels)
}

func (cd *CoreData) BlogPrivateLabels(blogID int32, authorID int32) ([]string, error) {
	return cd.PrivateLabels("blog", blogID, authorID)
}

func (cd *CoreData) SetBlogPrivateLabels(blogID int32, labels []string) error {
	return cd.SetPrivateLabels("blog", blogID, labels)
}

// BlogLabels returns author and private labels for a blog post.
func (cd *CoreData) BlogLabels(blogID int32, authorID int32) []templates.TopicLabel {
	var labels []templates.TopicLabel
	if als, err := cd.BlogAuthorLabels(blogID); err == nil {
		for _, l := range als {
			labels = append(labels, templates.TopicLabel{Name: l, Type: "author"})
		}
	}
	if pls, err := cd.BlogPrivateLabels(blogID, authorID); err == nil {
		for _, l := range pls {
			labels = append(labels, templates.TopicLabel{Name: l, Type: "private"})
		}
	}
	return labels
}

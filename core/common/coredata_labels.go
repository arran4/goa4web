package common

import (
	"sort"

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
func (cd *CoreData) PrivateLabels(item string, itemID int32) ([]string, error) {
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
		if r.Label == "new" || r.Label == "unread" {
			continue
		}
		userLabels = append(userLabels, r.Label)
	}
	sort.Strings(userLabels)
	labels := make([]string, 0, len(userLabels)+2)
	if !inverted["new"] {
		labels = append(labels, "new")
	}
	if !inverted["unread"] {
		labels = append(labels, "unread")
	}
	labels = append(labels, userLabels...)
	return labels, nil
}

// ClearPrivateLabelStatus removes stored new/unread inversions for an item across all users.
func (cd *CoreData) ClearPrivateLabelStatus(item string, itemID int32) error {
	if cd.queries == nil {
		return nil
	}
	if err := cd.queries.SystemClearContentPrivateLabel(cd.ctx, db.SystemClearContentPrivateLabelParams{Item: item, ItemID: itemID, Label: "new"}); err != nil {
		return err
	}
	return cd.queries.SystemClearContentPrivateLabel(cd.ctx, db.SystemClearContentPrivateLabelParams{Item: item, ItemID: itemID, Label: "unread"})
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

func (cd *CoreData) SetThreadPublicLabels(threadID int32, labels []string) error {
	return cd.SetPublicLabels("thread", threadID, labels)
}

func (cd *CoreData) ThreadPrivateLabels(threadID int32) ([]string, error) {
	return cd.PrivateLabels("thread", threadID)
}

func (cd *CoreData) ClearThreadPrivateLabelStatus(threadID int32) error {
	return cd.ClearPrivateLabelStatus("thread", threadID)
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

func (cd *CoreData) WritingPublicLabels(writingID int32) (public, owner []string, err error) {
	return cd.PublicLabels("writing", writingID)
}

func (cd *CoreData) AddWritingPublicLabel(writingID int32, label string) error {
	return cd.AddPublicLabel("writing", writingID, label)
}

func (cd *CoreData) RemoveWritingPublicLabel(writingID int32, label string) error {
	return cd.RemovePublicLabel("writing", writingID, label)
}

func (cd *CoreData) SetWritingPublicLabels(writingID int32, labels []string) error {
	return cd.SetPublicLabels("writing", writingID, labels)
}

func (cd *CoreData) WritingPrivateLabels(writingID int32) ([]string, error) {
	return cd.PrivateLabels("writing", writingID)
}

func (cd *CoreData) SetWritingPrivateLabels(writingID int32, labels []string) error {
	return cd.SetPrivateLabels("writing", writingID, labels)
}

// News

func (cd *CoreData) NewsPublicLabels(newsID int32) (public, owner []string, err error) {
	return cd.PublicLabels("news", newsID)
}

func (cd *CoreData) AddNewsPublicLabel(newsID int32, label string) error {
	return cd.AddPublicLabel("news", newsID, label)
}

func (cd *CoreData) RemoveNewsPublicLabel(newsID int32, label string) error {
	return cd.RemovePublicLabel("news", newsID, label)
}

func (cd *CoreData) SetNewsPublicLabels(newsID int32, labels []string) error {
	return cd.SetPublicLabels("news", newsID, labels)
}

func (cd *CoreData) NewsPrivateLabels(newsID int32) ([]string, error) {
	return cd.PrivateLabels("news", newsID)
}

func (cd *CoreData) SetNewsPrivateLabels(newsID int32, labels []string) error {
	return cd.SetPrivateLabels("news", newsID, labels)
}

// Blogs

func (cd *CoreData) BlogPublicLabels(blogID int32) (public, owner []string, err error) {
	return cd.PublicLabels("blog", blogID)
}

func (cd *CoreData) AddBlogPublicLabel(blogID int32, label string) error {
	return cd.AddPublicLabel("blog", blogID, label)
}

func (cd *CoreData) RemoveBlogPublicLabel(blogID int32, label string) error {
	return cd.RemovePublicLabel("blog", blogID, label)
}

func (cd *CoreData) SetBlogPublicLabels(blogID int32, labels []string) error {
	return cd.SetPublicLabels("blog", blogID, labels)
}

func (cd *CoreData) BlogPrivateLabels(blogID int32) ([]string, error) {
	return cd.PrivateLabels("blog", blogID)
}

func (cd *CoreData) SetBlogPrivateLabels(blogID int32, labels []string) error {
	return cd.SetPrivateLabels("blog", blogID, labels)
}

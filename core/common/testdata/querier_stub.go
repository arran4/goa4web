package testdata

import "github.com/arran4/goa4web/internal/db"

// VisibleThreadLabels returns sample label rows for thread visibility tests.
func VisibleThreadLabels(userID int32, labels ...string) []*db.ListContentPrivateLabelsRow {
	rows := make([]*db.ListContentPrivateLabelsRow, 0, len(labels))
	for _, l := range labels {
		rows = append(rows, &db.ListContentPrivateLabelsRow{
			Item:   "thread",
			ItemID: 1,
			UserID: userID,
			Label:  l,
			Invert: false,
		})
	}
	return rows
}

// SampleSubscriptions returns subscription rows for a given user.
func SampleSubscriptions(userID int32, patterns ...string) []*db.ListSubscriptionsByUserRow {
	rows := make([]*db.ListSubscriptionsByUserRow, 0, len(patterns))
	for idx, p := range patterns {
		rows = append(rows, &db.ListSubscriptionsByUserRow{ID: int32(idx + 1), Pattern: p, Method: "internal"})
	}
	return rows
}

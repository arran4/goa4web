package db

// EmailsByUserID groups verified email rows by the owning user.
func EmailsByUserID(rows []*GetVerifiedUserEmailsRow) map[int32][]string {
	emails := make(map[int32][]string, len(rows))
	for _, row := range rows {
		emails[row.UserID] = append(emails[row.UserID], row.Email)
	}
	return emails
}

// PrimaryEmail returns the highest priority email from the provided slice.
func PrimaryEmail(emails []string) string {
	if len(emails) == 0 {
		return ""
	}
	return emails[0]
}

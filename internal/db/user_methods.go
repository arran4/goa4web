package db

// GetLogin satisfies the gobookmarks.User interface
func (u *User) GetLogin() string {
	if u.Username.Valid {
		return u.Username.String
	}
	return ""
}

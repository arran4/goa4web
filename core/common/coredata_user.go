package common

import (
	"database/sql"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

// UserSettings returns preference settings for a given user.
func (cd *CoreData) UserSettings(userID int32) (*db.Preference, error) {
	if cd == nil || cd.queries == nil {
		return nil, nil
	}
	return cd.queries.GetPreferenceForLister(cd.ctx, userID)
}

// UserLanguages fetches languages selected by the user.
func (cd *CoreData) UserLanguages(userID int32) ([]*db.UserLanguage, error) {
	if cd == nil || cd.queries == nil {
		return nil, nil
	}
	return cd.queries.GetUserLanguages(cd.ctx, userID)
}

// UserEmails retrieves all emails associated with a user for the current lister.
func (cd *CoreData) UserEmails(userID int32) ([]*db.UserEmail, error) {
	if cd == nil || cd.queries == nil {
		return nil, nil
	}
	return cd.queries.ListUserEmailsForLister(cd.ctx, db.ListUserEmailsForListerParams{UserID: userID, ListerID: cd.UserID})
}

// AddUserEmail associates an email with the user.
func (cd *CoreData) AddUserEmail(userID int32, email, code string, expire time.Time) error {
	if cd == nil || cd.queries == nil {
		return nil
	}
	return cd.queries.InsertUserEmail(cd.ctx, db.InsertUserEmailParams{
		UserID:                userID,
		Email:                 email,
		VerifiedAt:            sql.NullTime{},
		LastVerificationCode:  sql.NullString{String: code, Valid: code != ""},
		VerificationExpiresAt: sql.NullTime{Time: expire, Valid: true},
		NotificationPriority:  0,
	})
}

// SaveEmail updates email notification preferences for the user.
func (cd *CoreData) SaveEmail(userID int32, updates, auto bool) error {
	if cd == nil || cd.queries == nil {
		return nil
	}
	_, err := cd.queries.GetPreferenceForLister(cd.ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return cd.queries.InsertEmailPreferenceForLister(cd.ctx, db.InsertEmailPreferenceForListerParams{
				EmailForumUpdates:    sql.NullBool{Bool: updates, Valid: true},
				AutoSubscribeReplies: auto,
				ListerID:             userID,
			})
		}
		return err
	}
	if err := cd.queries.UpdateEmailForumUpdatesForLister(cd.ctx, db.UpdateEmailForumUpdatesForListerParams{
		EmailForumUpdates: sql.NullBool{Bool: updates, Valid: true},
		ListerID:          userID,
	}); err != nil {
		return err
	}
	return cd.queries.UpdateAutoSubscribeRepliesForLister(cd.ctx, db.UpdateAutoSubscribeRepliesForListerParams{
		AutoSubscribeReplies: auto,
		ListerID:             userID,
	})
}

// DeleteEmail removes an email belonging to the user.
func (cd *CoreData) DeleteEmail(userID, id int32) error {
	if cd == nil || cd.queries == nil {
		return nil
	}
	return cd.queries.DeleteUserEmailForOwner(cd.ctx, db.DeleteUserEmailForOwnerParams{ID: id, OwnerID: userID})
}

// AddEmail promotes an email to receive notifications.
func (cd *CoreData) AddEmail(userID, emailID int32) error {
	if cd == nil || cd.queries == nil {
		return nil
	}
	val, err := cd.queries.GetMaxNotificationPriority(cd.ctx, userID)
	if err != nil {
		return err
	}
	var max int32
	switch v := val.(type) {
	case int64:
		max = int32(v)
	case int32:
		max = v
	}
	return cd.queries.SetNotificationPriorityForLister(cd.ctx, db.SetNotificationPriorityForListerParams{
		ListerID:             userID,
		NotificationPriority: max + 1,
		ID:                   emailID,
	})
}

// UserGallery returns images uploaded by the user for the current lister.
func (cd *CoreData) UserGallery(userID int32, limit, offset int32) ([]*db.UploadedImage, error) {
	if cd == nil || cd.queries == nil {
		return nil, nil
	}
	return cd.queries.ListUploadedImagesByUserForLister(cd.ctx, db.ListUploadedImagesByUserForListerParams{
		ListerID:      cd.UserID,
		UserID:        userID,
		ListerMatchID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		Limit:         limit,
		Offset:        offset,
	})
}

// PublicProfile fetches a user's public profile information.
func (cd *CoreData) PublicProfile(userID int32) (*db.SystemGetUserByIDRow, error) {
	if cd == nil || cd.queries == nil {
		return nil, nil
	}
	return cd.queries.SystemGetUserByID(cd.ctx, userID)
}

// PagedUsers returns users with their roles. Pagination parameters are
// currently ignored as the underlying query does not yet support them.
func (cd *CoreData) PagedUsers(limit, offset int32) ([]*db.ListUsersWithRolesRow, error) {
	if cd == nil || cd.queries == nil {
		return nil, nil
	}
	return cd.queries.ListUsersWithRoles(cd.ctx)
}

// UserNotifications returns notifications for a user.
func (cd *CoreData) UserNotifications(userID int32, limit, offset int32) ([]*db.Notification, error) {
	if cd == nil || cd.queries == nil {
		return nil, nil
	}
	return cd.queries.ListNotificationsForLister(cd.ctx, db.ListNotificationsForListerParams{ListerID: userID, Limit: limit, Offset: offset})
}

// DeleteSubscription removes a subscription for a user.
func (cd *CoreData) DeleteSubscription(userID, subID int32) error {
	if cd == nil || cd.queries == nil {
		return nil
	}
	return cd.queries.DeleteSubscriptionByIDForSubscriber(cd.ctx, db.DeleteSubscriptionByIDForSubscriberParams{SubscriberID: userID, ID: subID})
}

// UpdateSubscriptions updates subscription settings for the user.
func (cd *CoreData) UpdateSubscriptions(userID, subID int32, pattern, method string) error {
	if cd == nil || cd.queries == nil {
		return nil
	}
	return cd.queries.UpdateSubscriptionByIDForSubscriber(cd.ctx, db.UpdateSubscriptionByIDForSubscriberParams{
		Pattern:      pattern,
		Method:       method,
		SubscriberID: userID,
		ID:           subID,
	})
}

// SetUserLanguage sets the default language preference for the user.
func (cd *CoreData) SetUserLanguage(userID, languageID int32) error {
	if cd == nil || cd.queries == nil {
		return nil
	}
	pref, err := cd.queries.GetPreferenceForLister(cd.ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return cd.queries.InsertPreferenceForLister(cd.ctx, db.InsertPreferenceForListerParams{
				LanguageID: sql.NullInt32{Int32: languageID, Valid: languageID != 0},
				ListerID:   userID,
				PageSize:   int32(cd.Config.PageSizeDefault),
				Timezone:   sql.NullString{},
			})
		}
		return err
	}
	return cd.queries.UpdatePreferenceForLister(cd.ctx, db.UpdatePreferenceForListerParams{
		LanguageID: sql.NullInt32{Int32: languageID, Valid: languageID != 0},
		ListerID:   userID,
		PageSize:   pref.PageSize,
		Timezone:   pref.Timezone,
	})
}

// SetUserLanguages replaces a user's language selections.
func (cd *CoreData) SetUserLanguages(userID int32, langs []int32) error {
	if cd == nil || cd.queries == nil {
		return nil
	}
	if err := cd.queries.DeleteUserLanguagesForUser(cd.ctx, userID); err != nil {
		return err
	}
	for _, l := range langs {
		if err := cd.queries.InsertUserLang(cd.ctx, db.InsertUserLangParams{UsersIdusers: userID, LanguageID: l}); err != nil {
			return err
		}
	}
	return nil
}

// SetTimezone updates the user's timezone preference.
func (cd *CoreData) SetTimezone(userID int32, tz string) error {
	if cd == nil || cd.queries == nil {
		return nil
	}
	pref, err := cd.queries.GetPreferenceForLister(cd.ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return cd.queries.InsertPreferenceForLister(cd.ctx, db.InsertPreferenceForListerParams{
				LanguageID: sql.NullInt32{},
				ListerID:   userID,
				PageSize:   int32(cd.Config.PageSizeDefault),
				Timezone:   sql.NullString{String: tz, Valid: tz != ""},
			})
		}
		return err
	}
	return cd.queries.UpdatePreferenceForLister(cd.ctx, db.UpdatePreferenceForListerParams{
		LanguageID: pref.LanguageID,
		ListerID:   userID,
		PageSize:   pref.PageSize,
		Timezone:   sql.NullString{String: tz, Valid: tz != ""},
	})
}

// SaveProfile toggles the public profile setting for a user.
func (cd *CoreData) SaveProfile(userID int32, enable bool) error {
	if cd == nil || cd.queries == nil {
		return nil
	}
	var ts sql.NullTime
	if enable {
		ts = sql.NullTime{Time: time.Now(), Valid: true}
	}
	return cd.queries.UpdatePublicProfileEnabledAtForUser(cd.ctx, db.UpdatePublicProfileEnabledAtForUserParams{
		EnabledAt: ts,
		UserID:    userID,
		GranteeID: sql.NullInt32{Int32: userID, Valid: userID != 0},
	})
}

// UpdatePermissions updates a permission's role name.
func (cd *CoreData) UpdatePermissions(id int32, role string) error {
	if cd == nil || cd.queries == nil {
		return nil
	}
	return cd.queries.AdminUpdateUserRole(cd.ctx, db.AdminUpdateUserRoleParams{IduserRoles: id, Name: role})
}

// AllowPermission grants a permission to a user.
func (cd *CoreData) AllowPermission(userID, topicID int32, role string, level int32, inviteMax int32) error {
	if cd == nil || cd.queries == nil {
		return nil
	}
	// No direct query implemented; placeholder for future expansion.
	return nil
}

// DisallowPermission revokes a permission from a user.
func (cd *CoreData) DisallowPermission(id int32) error {
	if cd == nil || cd.queries == nil {
		return nil
	}
	return cd.queries.AdminDeleteUserRole(cd.ctx, id)
}

// SaveNotificationDigestPreferences updates notification digest settings for the user.
func (cd *CoreData) SaveNotificationDigestPreferences(userID int32, dailyHour *int, markRead bool, weeklyDay, weeklyHour, monthlyDay, monthlyHour *int) error {
	if cd == nil || cd.queries == nil {
		return nil
	}
	var sqlDailyHour, sqlWeeklyDay, sqlWeeklyHour, sqlMonthlyDay, sqlMonthlyHour sql.NullInt32
	if dailyHour != nil {
		sqlDailyHour = sql.NullInt32{Int32: int32(*dailyHour), Valid: true}
	}
	if weeklyDay != nil {
		sqlWeeklyDay = sql.NullInt32{Int32: int32(*weeklyDay), Valid: true}
	}
	if weeklyHour != nil {
		sqlWeeklyHour = sql.NullInt32{Int32: int32(*weeklyHour), Valid: true}
	}
	if monthlyDay != nil {
		sqlMonthlyDay = sql.NullInt32{Int32: int32(*monthlyDay), Valid: true}
	}
	if monthlyHour != nil {
		sqlMonthlyHour = sql.NullInt32{Int32: int32(*monthlyHour), Valid: true}
	}

	_, err := cd.queries.GetPreferenceForLister(cd.ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			if err := cd.queries.InsertEmailPreferenceForLister(cd.ctx, db.InsertEmailPreferenceForListerParams{
				ListerID: userID,
			}); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return cd.queries.UpdateNotificationDigestPreferences(cd.ctx, db.UpdateNotificationDigestPreferencesParams{
		DailyDigestHour:     sqlDailyHour,
		DailyDigestMarkRead: markRead,
		WeeklyDigestDay:     sqlWeeklyDay,
		WeeklyDigestHour:    sqlWeeklyHour,
		MonthlyDigestDay:    sqlMonthlyDay,
		MonthlyDigestHour:   sqlMonthlyHour,
		ListerID:            userID,
	})
}

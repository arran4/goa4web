package admin

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/gorilla/mux"
)

func getTargetUserID(r *http.Request) (int32, error) {
	vars := mux.Vars(r)
	uidStr := vars["user"]
	uid, err := strconv.Atoi(uidStr)
	if err != nil {
		return 0, err
	}
	return int32(uid), nil
}

func generateVerificationCode() string {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		log.Printf("rand read: %v", err)
		return ""
	}
	return hex.EncodeToString(buf[:])
}

// AdminAddEmailTask adds a new email to a user.
type AdminAddEmailTask struct {
	tasks.TaskString
}

var adminAddEmailTask = &AdminAddEmailTask{TaskString: TaskAddEmail}

var _ tasks.Task = (*AdminAddEmailTask)(nil)
var _ notif.DirectEmailNotificationTemplateProvider = (*AdminAddEmailTask)(nil)
var _ tasks.EmailTemplatesRequired = (*AdminAddEmailTask)(nil)

func (t AdminAddEmailTask) Action(w http.ResponseWriter, r *http.Request) any {
	targetUID, err := getTargetUserID(r)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}

	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	emailAddr := r.FormValue("new_email")
	if emailAddr == "" {
		return handlers.RefreshDirectHandler{TargetURL: fmt.Sprintf("/admin/user/%d", targetUID)}
	}
	if _, err := mail.ParseAddress(emailAddr); err != nil {
		return handlers.RefreshDirectHandler{TargetURL: fmt.Sprintf("/admin/user/%d?error=invalid+email", targetUID)}
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	if ue, err := queries.GetUserEmailByEmail(r.Context(), emailAddr); err == nil && ue.VerifiedAt.Valid {
		return handlers.RefreshDirectHandler{TargetURL: fmt.Sprintf("/admin/user/%d?error=email+exists", targetUID)}
	}

	code := generateVerificationCode()
	expiryHours := cd.Config.EmailVerificationExpiryHours
	if expiryHours <= 0 {
		expiryHours = 24
	}
	expire := time.Now().Add(time.Duration(expiryHours) * time.Hour)

	// Use AdminAddUserEmail or InsertUserEmail. InsertUserEmail is what addEmailTask uses.
	// But AdminAddUserEmail exists.
	// InsertUserEmailParams: UserID, Email, VerifiedAt, LastVerificationCode, VerificationExpiresAt, NotificationPriority
	// AdminAddUserEmailParams: UserID, Email, VerifiedAt, NotificationPriority
	// AdminAddUserEmail doesn't set verification code.
	// So I should use InsertUserEmail to set the code.
	if err := queries.InsertUserEmail(r.Context(), db.InsertUserEmailParams{
		UserID:                targetUID,
		Email:                 emailAddr,
		VerifiedAt:            sql.NullTime{}, // Not verified yet
		LastVerificationCode:  sql.NullString{String: code, Valid: true},
		VerificationExpiresAt: sql.NullTime{Time: expire, Valid: true},
		NotificationPriority:  0, // Default priority
	}); err != nil {
		log.Printf("insert user email: %v", err)
		return fmt.Errorf("insert user email fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	// Prepare event for email sending
	path := "/usr/email/verify?code=" + code
	cfg := cd.Config
	page := "http://" + r.Host + path
	if cfg.BaseURL != "" {
		page = strings.TrimRight(cfg.BaseURL, "/") + path
	}
	evt := cd.Event()
	if evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		evt.Data["page"] = page
		evt.Data["email"] = emailAddr
		evt.Data["URL"] = page
		evt.Data["VerificationCode"] = code
		evt.Data["Token"] = code
		// Fetch target user for template context
		if user, err := queries.SystemGetUserByID(r.Context(), targetUID); err == nil {
			evt.Data["Username"] = user.Username.String
		}
		evt.Data["ExpiresAt"] = expire
	}

	return handlers.RefreshDirectHandler{TargetURL: fmt.Sprintf("/admin/user/%d", targetUID)}
}

func (AdminAddEmailTask) DirectEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateVerify.EmailTemplates(), true
}

func (AdminAddEmailTask) RequiredTemplates() []tasks.Template {
	return EmailTemplateVerify.RequiredTemplates()
}

func (AdminAddEmailTask) DirectEmailAddress(evt eventbus.TaskEvent) (string, error) {
	if evt.Data != nil {
		if email, ok := evt.Data["email"].(string); ok {
			return email, nil
		}
	}
	return "", fmt.Errorf("email not provided")
}

// AdminDeleteEmailTask deletes a user's email.
type AdminDeleteEmailTask struct {
	tasks.TaskString
}

var adminDeleteEmailTask = &AdminDeleteEmailTask{TaskString: TaskDeleteEmail}

var _ tasks.Task = (*AdminDeleteEmailTask)(nil)

func (t AdminDeleteEmailTask) Action(w http.ResponseWriter, r *http.Request) any {
	targetUID, err := getTargetUserID(r)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}

	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	emailIDStr := r.FormValue("email_id")
	emailID, err := strconv.Atoi(emailIDStr)
	if err != nil {
		return fmt.Errorf("invalid email id: %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()

	// Verify email belongs to user
	ue, err := queries.AdminGetUserEmailByID(r.Context(), int32(emailID))
	if err != nil || ue.UserID != targetUID {
		return handlers.RefreshDirectHandler{TargetURL: fmt.Sprintf("/admin/user/%d", targetUID)}
	}

	if err := queries.AdminDeleteUserEmail(r.Context(), int32(emailID)); err != nil {
		return fmt.Errorf("delete user email fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return handlers.RefreshDirectHandler{TargetURL: fmt.Sprintf("/admin/user/%d", targetUID)}
}

// AdminVerifyEmailTask verifies a user's email.
type AdminVerifyEmailTask struct {
	tasks.TaskString
}

var adminVerifyEmailTask = &AdminVerifyEmailTask{TaskString: TaskVerifyEmail}

var _ tasks.Task = (*AdminVerifyEmailTask)(nil)

func (t AdminVerifyEmailTask) Action(w http.ResponseWriter, r *http.Request) any {
	targetUID, err := getTargetUserID(r)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}

	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	emailIDStr := r.FormValue("email_id")
	emailID, err := strconv.Atoi(emailIDStr)
	if err != nil {
		return fmt.Errorf("invalid email id: %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()

	// Verify email belongs to user
	ue, err := queries.AdminGetUserEmailByID(r.Context(), int32(emailID))
	if err != nil || ue.UserID != targetUID {
		return handlers.RefreshDirectHandler{TargetURL: fmt.Sprintf("/admin/user/%d", targetUID)}
	}

	if err := queries.AdminUpdateUserEmailDetails(r.Context(), db.AdminUpdateUserEmailDetailsParams{
		Email:                ue.Email,
		VerifiedAt:           sql.NullTime{Time: time.Now(), Valid: true},
		NotificationPriority: ue.NotificationPriority,
		ID:                   ue.ID,
	}); err != nil {
		return fmt.Errorf("verify user email fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return handlers.RefreshDirectHandler{TargetURL: fmt.Sprintf("/admin/user/%d", targetUID)}
}

// AdminUnverifyEmailTask unverifies a user's email.
type AdminUnverifyEmailTask struct {
	tasks.TaskString
}

var adminUnverifyEmailTask = &AdminUnverifyEmailTask{TaskString: TaskUnverifyEmail}

var _ tasks.Task = (*AdminUnverifyEmailTask)(nil)

func (t AdminUnverifyEmailTask) Action(w http.ResponseWriter, r *http.Request) any {
	targetUID, err := getTargetUserID(r)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}

	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	emailIDStr := r.FormValue("email_id")
	emailID, err := strconv.Atoi(emailIDStr)
	if err != nil {
		return fmt.Errorf("invalid email id: %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()

	// Verify email belongs to user
	ue, err := queries.AdminGetUserEmailByID(r.Context(), int32(emailID))
	if err != nil || ue.UserID != targetUID {
		return handlers.RefreshDirectHandler{TargetURL: fmt.Sprintf("/admin/user/%d", targetUID)}
	}

	if err := queries.AdminUpdateUserEmailDetails(r.Context(), db.AdminUpdateUserEmailDetailsParams{
		Email:                ue.Email,
		VerifiedAt:           sql.NullTime{Valid: false},
		NotificationPriority: ue.NotificationPriority,
		ID:                   ue.ID,
	}); err != nil {
		return fmt.Errorf("unverify user email fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	return handlers.RefreshDirectHandler{TargetURL: fmt.Sprintf("/admin/user/%d", targetUID)}
}

// AdminResendVerificationEmailTask resends the verification link for an unverified user email address.
type AdminResendVerificationEmailTask struct {
	tasks.TaskString
}

var adminResendVerificationEmailTask = &AdminResendVerificationEmailTask{TaskString: TaskResendVerification}

var _ tasks.Task = (*AdminResendVerificationEmailTask)(nil)
var _ notif.DirectEmailNotificationTemplateProvider = (*AdminResendVerificationEmailTask)(nil)
var _ tasks.EmailTemplatesRequired = (*AdminResendVerificationEmailTask)(nil)

func (t AdminResendVerificationEmailTask) Action(w http.ResponseWriter, r *http.Request) any {
	targetUID, err := getTargetUserID(r)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}

	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	emailIDStr := r.FormValue("email_id")
	emailID, err := strconv.Atoi(emailIDStr)
	if err != nil {
		return fmt.Errorf("invalid email id: %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()

	ue, err := queries.AdminGetUserEmailByID(r.Context(), int32(emailID))
	if err != nil || ue.UserID != targetUID {
		return handlers.RefreshDirectHandler{TargetURL: fmt.Sprintf("/admin/user/%d", targetUID)}
	}

	code := generateVerificationCode()
	expiryHours := cd.Config.EmailVerificationExpiryHours
	if expiryHours <= 0 {
		expiryHours = 24
	}
	expire := time.Now().Add(time.Duration(expiryHours) * time.Hour)

	if err := queries.SystemUpdateVerificationCode(r.Context(), db.SystemUpdateVerificationCodeParams{
		LastVerificationCode:  sql.NullString{String: code, Valid: code != ""},
		VerificationExpiresAt: sql.NullTime{Time: expire, Valid: true},
		ID:                    int32(emailID),
	}); err != nil {
		log.Printf("set verification code: %v", err)
	}

	path := "/usr/email/verify?code=" + code
	cfg := cd.Config
	page := "http://" + r.Host + path
	if cfg.BaseURL != "" {
		page = strings.TrimRight(cfg.BaseURL, "/") + path
	}
	evt := cd.Event()
	if evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		evt.Data["page"] = page
		evt.Data["email"] = ue.Email
		evt.Data["URL"] = page
		evt.Data["VerificationCode"] = code
		evt.Data["Token"] = code
		if user, err := queries.SystemGetUserByID(r.Context(), targetUID); err == nil {
			evt.Data["Username"] = user.Username.String
		}
		evt.Data["ExpiresAt"] = expire
	}
	return handlers.RefreshDirectHandler{TargetURL: fmt.Sprintf("/admin/user/%d", targetUID)}
}

func (AdminResendVerificationEmailTask) DirectEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateVerify.EmailTemplates(), true
}

func (AdminResendVerificationEmailTask) RequiredTemplates() []tasks.Template {
	return EmailTemplateVerify.RequiredTemplates()
}

func (AdminResendVerificationEmailTask) DirectEmailAddress(evt eventbus.TaskEvent) (string, error) {
	if evt.Data != nil {
		if email, ok := evt.Data["email"].(string); ok {
			return email, nil
		}
	}
	return "", fmt.Errorf("email not provided")
}

package user

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

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
)

// AddEmailTask handles user email verification requests and sends
// notifications directly to the specified address.
type AddEmailTask struct{ tasks.TaskString }

var addEmailTask = &AddEmailTask{TaskString: tasks.TaskString(TaskAdd)}

var _ tasks.Task = (*AddEmailTask)(nil)
var _ notif.DirectEmailNotificationTemplateProvider = (*AddEmailTask)(nil)

func (AddEmailTask) Action(w http.ResponseWriter, r *http.Request) any {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	uid, _ := session.Values["UID"].(int32)
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	emailAddr := r.FormValue("new_email")
	if emailAddr == "" {
		return handlers.RefreshDirectHandler{TargetURL: "/usr/email"}
	}
	if _, err := mail.ParseAddress(emailAddr); err != nil {
		return handlers.RefreshDirectHandler{TargetURL: "/usr/email?error=invalid+email"}
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if ue, err := queries.GetUserEmailByEmail(r.Context(), emailAddr); err == nil && ue.VerifiedAt.Valid {
		return handlers.RefreshDirectHandler{TargetURL: "/usr/email?error=email+exists"}
	}
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		log.Printf("rand read: %v", err)
	}
	code := hex.EncodeToString(buf[:])
	expire := time.Now().Add(24 * time.Hour)
	if err := queries.InsertUserEmail(r.Context(), db.InsertUserEmailParams{UserID: uid, Email: emailAddr, VerifiedAt: sql.NullTime{}, LastVerificationCode: sql.NullString{String: code, Valid: true}, VerificationExpiresAt: sql.NullTime{Time: expire, Valid: true}, NotificationPriority: 0}); err != nil {
		log.Printf("insert user email: %v", err)
		return fmt.Errorf("insert user email fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	path := "/usr/email/verify?code=" + code
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cfg := cd.Config
	page := "http://" + r.Host + path
	if cfg.HTTPHostname != "" {
		page = strings.TrimRight(cfg.HTTPHostname, "/") + path
	}
	evt := cd.Event()
	evt.Data["page"] = page
	evt.Data["email"] = emailAddr
	evt.Data["URL"] = page
	if user, err := cd.CurrentUser(); err == nil && user != nil {
		evt.Data["Username"] = user.Username.String
	}
	return handlers.RefreshDirectHandler{TargetURL: "/usr/email"}
}

func (AddEmailTask) Resend(w http.ResponseWriter, r *http.Request) any {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	uid, _ := session.Values["UID"].(int32)
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	ue, err := queries.GetUserEmailByID(r.Context(), int32(id))
	if err != nil || ue.UserID != uid {
		return handlers.RefreshDirectHandler{TargetURL: "/usr/email"}
	}
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		log.Printf("rand read: %v", err)
	}
	code := hex.EncodeToString(buf[:])
	expire := time.Now().Add(24 * time.Hour)
	if err := queries.SetVerificationCode(r.Context(), db.SetVerificationCodeParams{LastVerificationCode: sql.NullString{String: code, Valid: true}, VerificationExpiresAt: sql.NullTime{Time: expire, Valid: true}, ID: int32(id)}); err != nil {
		log.Printf("set verification code: %v", err)
	}
	path := "/usr/email/verify?code=" + code
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cfg := cd.Config
	page := "http://" + r.Host + path
	if cfg.HTTPHostname != "" {
		page = strings.TrimRight(cfg.HTTPHostname, "/") + path
	}
	evt := cd.Event()
	evt.Data["page"] = page
	evt.Data["email"] = ue.Email
	evt.Data["URL"] = page
	if user, err := cd.CurrentUser(); err == nil && user != nil {
		evt.Data["Username"] = user.Username.String
	}
	return handlers.RefreshDirectHandler{TargetURL: "/usr/email"}
}

func (AddEmailTask) Notify(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/usr/email", http.StatusSeeOther)
		return
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	val, _ := queries.GetMaxNotificationPriority(r.Context(), uid)
	var maxPr int32
	switch v := val.(type) {
	case int64:
		maxPr = int32(v)
	case int32:
		maxPr = v
	}
	if err := queries.SetNotificationPriority(r.Context(), db.SetNotificationPriorityParams{NotificationPriority: maxPr + 1, ID: int32(id)}); err != nil {
		log.Printf("set notification priority: %v", err)
	}
	http.Redirect(w, r, "/usr/email", http.StatusSeeOther)
}

func (AddEmailTask) DirectEmailTemplate() *notif.EmailTemplates {
	return notif.NewEmailTemplates("verifyEmail")
}

func (AddEmailTask) DirectEmailAddress(evt eventbus.TaskEvent) (string, error) {
	if evt.Data != nil {
		if email, ok := evt.Data["email"].(string); ok {
			return email, nil
		}
	}
	return "", fmt.Errorf("email not provided")
}

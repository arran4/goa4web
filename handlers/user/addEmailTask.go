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
type AddEmailTask struct {
	tasks.TaskString
	codeGenerator func() (string, error)
}

var addEmailTask = &AddEmailTask{TaskString: tasks.TaskString(TaskAdd)}

var _ tasks.Task = (*AddEmailTask)(nil)
var _ notif.DirectEmailNotificationTemplateProvider = (*AddEmailTask)(nil)
var _ tasks.EmailTemplatesRequired = (*AddEmailTask)(nil)

func (t AddEmailTask) generateVerificationCode() string {
	if t.codeGenerator != nil {
		if code, err := t.codeGenerator(); err == nil {
			return code
		}
	}
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		log.Printf("rand read: %v", err)
		return ""
	}
	return hex.EncodeToString(buf[:])
}

func (t AddEmailTask) Action(w http.ResponseWriter, r *http.Request) any {
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
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	if ue, err := queries.GetUserEmailByEmail(r.Context(), emailAddr); err == nil && ue.VerifiedAt.Valid {
		return handlers.RefreshDirectHandler{TargetURL: "/usr/email?error=email+exists"}
	}
	code := t.generateVerificationCode()
	expiryHours := cd.Config.EmailVerificationExpiryHours
	if expiryHours <= 0 {
		expiryHours = 24
	}
	expire := time.Now().Add(time.Duration(expiryHours) * time.Hour)
	if err := cd.AddUserEmail(uid, emailAddr, code, expire); err != nil {
		log.Printf("insert user email: %v", err)
		return fmt.Errorf("insert user email fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	path := "/usr/email/verify?code=" + code
	cfg := cd.Config
	page := "http://" + r.Host + path
	if cfg.HTTPHostname != "" {
		page = strings.TrimRight(cfg.HTTPHostname, "/") + path
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
		if user, err := cd.CurrentUser(); err == nil && user != nil {
			evt.Data["Username"] = user.Username.String
		}
	}
	evt.Data["ExpiresAt"] = expire
	return handlers.RefreshDirectHandler{TargetURL: "/usr/email"}
}

func (t AddEmailTask) Resend(w http.ResponseWriter, r *http.Request) any {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	uid, _ := session.Values["UID"].(int32)
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	id, _ := strconv.Atoi(r.FormValue("id"))
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	ue, err := queries.GetUserEmailByID(r.Context(), int32(id))
	if err != nil || ue.UserID != uid {
		return handlers.RefreshDirectHandler{TargetURL: "/usr/email"}
	}
	code := t.generateVerificationCode()
	expiryHours := cd.Config.EmailVerificationExpiryHours
	if expiryHours <= 0 {
		expiryHours = 24
	}
	expire := time.Now().Add(time.Duration(expiryHours) * time.Hour)
	if err := queries.SetVerificationCodeForLister(r.Context(), db.SetVerificationCodeForListerParams{ListerID: uid, LastVerificationCode: sql.NullString{String: code, Valid: code != ""}, VerificationExpiresAt: sql.NullTime{Time: expire, Valid: true}, ID: int32(id)}); err != nil {
		log.Printf("set verification code: %v", err)
	}
	path := "/usr/email/verify?code=" + code
	cfg := cd.Config
	page := "http://" + r.Host + path
	if cfg.HTTPHostname != "" {
		page = strings.TrimRight(cfg.HTTPHostname, "/") + path
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
		if user, err := cd.CurrentUser(); err == nil && user != nil {
			evt.Data["Username"] = user.Username.String
		}
	}
	evt.Data["ExpiresAt"] = expire
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
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if err := cd.AddEmail(uid, int32(id)); err != nil {
		log.Printf("set notification priority: %v", err)
	}
	http.Redirect(w, r, "/usr/email", http.StatusSeeOther)
}

func (AddEmailTask) DirectEmailTemplate(evt eventbus.TaskEvent) (templates *notif.EmailTemplates, send bool) {
	return EmailTemplateVerify.EmailTemplates(), true
}

func (AddEmailTask) EmailTemplatesRequired() []tasks.Page {
	return EmailTemplateVerify.RequiredPages()
}

func (AddEmailTask) DirectEmailAddress(evt eventbus.TaskEvent) (string, error) {
	if evt.Data != nil {
		if email, ok := evt.Data["email"].(string); ok {
			return email, nil
		}
	}
	return "", fmt.Errorf("email not provided")
}

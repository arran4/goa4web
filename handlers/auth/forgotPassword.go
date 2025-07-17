package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"log"
	"net/http"

	"github.com/arran4/goa4web/internal/tasks"

	common "github.com/arran4/goa4web/core/common"
	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
)

type ForgotPasswordTask struct {
	tasks.TaskString
}

var _ tasks.Task = (*ForgotPasswordTask)(nil)
var _ notif.SelfNotificationTemplateProvider = (*ForgotPasswordTask)(nil)
var _ notif.AdminEmailTemplateProvider = (*ForgotPasswordTask)(nil)

// ForgotPasswordTask handles password reset requests.
var forgotPasswordTask = &ForgotPasswordTask{
	TaskString: TaskUserResetPassword,
}

func (f ForgotPasswordTask) AdminEmailTemplate() *notif.EmailTemplates {
	return &notif.EmailTemplates{
		Text: "adminNotificationUserRequestPasswordResetEmail.gotxt",
		HTML: "adminNotificationUserRequestPasswordResetEmail.gohtml",
	}
}

func (f ForgotPasswordTask) AdminInternalNotificationTemplate() *string {
	v := "adminNotificationUserRequestPasswordResetEmail.gotxt"
	return &v
}

func (f ForgotPasswordTask) SelfEmailTemplate() *notif.EmailTemplates {
	return &notif.EmailTemplates{
		Text: "userRequestPasswordResetEmail.gotxt",
		HTML: "userRequestPasswordResetEmail.gohtml",
	}
}

func (f ForgotPasswordTask) SelfInternalNotificationTemplate() *string {
	s := "password_reset.gotxt"
	return &s
}

func (ForgotPasswordTask) Page(w http.ResponseWriter, r *http.Request) {
	handlers.TemplateHandler(w, r, "forgotPasswordPage.gohtml", r.Context().Value(common.KeyCoreData))
}

func (ForgotPasswordTask) Action(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	username := r.PostFormValue("username")
	pw := r.PostFormValue("password")
	if username == "" || pw == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	row, err := queries.GetUserByUsername(r.Context(), sql.NullString{String: username, Valid: true})
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if row.Email == "" {
		http.Error(w, "no verified email", http.StatusBadRequest)
		return
	}
	hash, alg, err := HashPassword(pw)
	if err != nil {
		http.Error(w, "hash error", http.StatusInternalServerError)
		return
	}
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		http.Error(w, "rand", http.StatusInternalServerError)
		return
	}
	code := hex.EncodeToString(buf[:])
	if err := queries.CreatePasswordReset(r.Context(), db.CreatePasswordResetParams{UserID: row.Idusers, Passwd: hash, PasswdAlgorithm: alg, VerificationCode: code}); err != nil {
		log.Printf("create reset: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if row.Email != "" {
		if cd, ok := r.Context().Value(common.KeyCoreData).(*common.CoreData); ok {
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				evt.Data["reset"] = notif.PasswordResetInfo{Username: row.Username.String, Code: code}
			}
		}
		// OLD _ = emailutil.CreateEmailTemplateAndQueue(r.Context(), queries, row.Idusers, row.Email, page, handlers.TaskUserResetPassword, code)
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

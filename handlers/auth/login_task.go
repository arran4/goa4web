package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

// LoginTask handles rendering and processing of the login form.
type LoginTask struct {
	tasks.TaskString
}

// loginTask handles login requests.
var loginTask = &LoginTask{TaskString: TaskLogin}

// ensure LoginTask conforms to tasks.Task
var _ tasks.Task = (*LoginTask)(nil)

// Page serves the username/password login form.
func (LoginTask) Page(w http.ResponseWriter, r *http.Request) {
	renderLoginForm(w, r, r.URL.Query().Get("error"))
}

// Action processes the submitted login form.
func (LoginTask) Action(w http.ResponseWriter, r *http.Request) any {
	if config.AppRuntimeConfig.LogFlags&config.LogFlagAuth != 0 {
		sess, _ := core.GetSession(r)
		log.Printf("login attempt for %s session=%s", r.PostFormValue("username"), handlers.HashSessionID(sess.ID))
	}

	username := r.PostFormValue("username")
	password := r.PostFormValue("password")

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	row, err := queries.Login(r.Context(), sql.NullString{String: username, Valid: true})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if err := queries.InsertLoginAttempt(r.Context(), db.InsertLoginAttemptParams{Username: username, IpAddress: strings.Split(r.RemoteAddr, ":")[0]}); err != nil {
				log.Printf("insert login attempt: %v", err)
			}
			return loginFormHandler{msg: "No such user"}
		}
		return fmt.Errorf("login query fail %w", err)
	}

	if !VerifyPassword(password, row.Passwd.String, row.PasswdAlgorithm.String) {
		expiry := time.Now().Add(-time.Duration(config.AppRuntimeConfig.PasswordResetExpiryHours) * time.Hour)
		reset, err := queries.GetPasswordResetByUser(r.Context(), db.GetPasswordResetByUserParams{UserID: row.Idusers, CreatedAt: expiry})
		code := r.FormValue("code")
		if err == nil && VerifyPassword(password, reset.Passwd, reset.PasswdAlgorithm) {
			if code != "" && code == reset.VerificationCode {
				_ = queries.MarkPasswordResetVerified(r.Context(), reset.ID)
				_ = queries.InsertPassword(r.Context(), db.InsertPasswordParams{UsersIdusers: reset.UserID, Passwd: reset.Passwd, PasswdAlgorithm: sql.NullString{String: reset.PasswdAlgorithm, Valid: true}})
			} else {
				type Data struct {
					*common.CoreData
					ID int32
				}
				data := Data{
					CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
					ID:       reset.ID,
				}
				return handlers.TemplateWithDataHandler("passwordVerifyPage.gohtml", data)
			}
		} else {
			if err := queries.InsertLoginAttempt(r.Context(), db.InsertLoginAttemptParams{Username: username, IpAddress: strings.Split(r.RemoteAddr, ":")[0]}); err != nil {
				log.Printf("insert login attempt: %v", err)
			}
			return loginFormHandler{msg: "Invalid password"}
		}
	}

	if _, err := queries.UserHasLoginRole(r.Context(), row.Idusers); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return loginFormHandler{msg: "approval is pending"}
		}
		return fmt.Errorf("user role %w", err)
	}

	if row.PasswdAlgorithm.String == "" || row.PasswdAlgorithm.String == "md5" {
		newHash, newAlg, err := HashPassword(password)
		if err == nil {
			if err := queries.InsertPassword(r.Context(), db.InsertPasswordParams{UsersIdusers: row.Idusers, Passwd: newHash, PasswdAlgorithm: sql.NullString{String: newAlg, Valid: true}}); err != nil {
				log.Printf("insert password: %v", err)
			}
		}
	}

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	session.Values["UID"] = int32(row.Idusers)
	session.Values["LoginTime"] = time.Now().Unix()
	session.Values["ExpiryTime"] = time.Now().AddDate(1, 0, 0).Unix()

	backURL := r.FormValue("back")
	backMethod := r.FormValue("method")
	backData := r.FormValue("data")

	if err := session.Save(r, w); err != nil {
		return fmt.Errorf("session save %w", err)
	}

	if config.AppRuntimeConfig.LogFlags&config.LogFlagAuth != 0 {
		log.Printf("login success uid=%d session=%s", row.Idusers, handlers.HashSessionID(session.ID))
	}

	if backURL != "" {
		if backMethod == "" || backMethod == http.MethodGet {
			return handlers.RedirectHandler(backURL)
		}
		vals, err := url.ParseQuery(backData)
		if err != nil {
			return fmt.Errorf("parse back data %w", err)
		}
		return redirectBackPageHandler{BackURL: backURL, Method: backMethod, Values: vals}
	}

	return handlers.RefreshDirectHandler{TargetURL: "/"}
}

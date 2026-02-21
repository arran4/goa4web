package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/arran4/goa4web/core/consts"

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
var _ tasks.TemplatesRequired = (*LoginTask)(nil)

const (
	templateLoginPage          = "loginPage.gohtml"
	templatePasswordVerifyPage = "passwordVerifyPage.gohtml"
)

// Page serves the username/password login form.
func (LoginTask) Page(w http.ResponseWriter, r *http.Request) {
	renderLoginForm(w, r, r.URL.Query().Get("error"), r.URL.Query().Get("notice"))
}

// Action processes the submitted login form.
func (LoginTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd.Config.LogFlags&config.LogFlagAuth != 0 {
		sess, _ := core.GetSession(r)
		log.Printf("login attempt for %s session=%s", r.PostFormValue("username"), handlers.HashSessionID(sess.ID))
	}

	username := r.PostFormValue("username")
	password := r.PostFormValue("password")

	queries := cd.Queries()

	cfg := cd.Config
	ip := strings.Split(r.RemoteAddr, ":")[0]
	if cfg.LoginAttemptThreshold > 0 {
		since := time.Now().Add(-time.Duration(cfg.LoginAttemptWindow) * time.Minute)
		cnt, err := queries.SystemCountRecentLoginAttempts(r.Context(), db.SystemCountRecentLoginAttemptsParams{Username: username, IpAddress: ip, CreatedAt: since})
		if err != nil {
			log.Printf("count login attempts: %v", err)
		} else if cnt >= int64(cfg.LoginAttemptThreshold) {
			return loginFormHandler{msg: "Too many failed attempts"}
		}
	}

	row, err := cd.UserCredentials(username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if err := queries.SystemInsertLoginAttempt(r.Context(), db.SystemInsertLoginAttemptParams{Username: username, IpAddress: strings.Split(r.RemoteAddr, ":")[0]}); err != nil {
				log.Printf("insert login attempt: %v", err)
			}
			return loginFormHandler{msg: "Invalid username or password"}
		}
		return fmt.Errorf("LoginTask.Action: user credentials query: %w", err)
	}

	if !VerifyPassword(password, row.Passwd.String, row.PasswdAlgorithm.String) {
		expiry := time.Now().Add(-time.Duration(cd.Config.PasswordResetExpiryHours) * time.Hour)
		reset, err := queries.GetPasswordResetByUser(r.Context(), db.GetPasswordResetByUserParams{UserID: row.Idusers, CreatedAt: expiry})
		if err == nil && VerifyPassword(password, reset.Passwd, reset.PasswdAlgorithm) {
			code := r.FormValue("code")
			if code != "" {
				if err := cd.VerifyPasswordReset(code, password); err != nil {
					if err := queries.SystemInsertLoginAttempt(r.Context(), db.SystemInsertLoginAttemptParams{Username: username, IpAddress: strings.Split(r.RemoteAddr, ":")[0]}); err != nil {
						log.Printf("insert login attempt: %v", err)
					}
					return loginFormHandler{msg: "Invalid username or password"}
				}
			} else {
				type Data struct {
					ID int32
				}
				cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
				cd.PageTitle = "Verify Password"
				data := Data{ID: reset.ID}
				return handlers.TemplateWithDataHandler(templatePasswordVerifyPage, data)
			}
		} else {
			if err := queries.SystemInsertLoginAttempt(r.Context(), db.SystemInsertLoginAttemptParams{Username: username, IpAddress: strings.Split(r.RemoteAddr, ":")[0]}); err != nil {
				log.Printf("insert login attempt: %v", err)
			}
			return loginFormHandler{msg: "Invalid username or password"}
		}
	}

	if _, err := queries.GetLoginRoleForUser(r.Context(), row.Idusers); err != nil {
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

	session := cd.GetSession()
	session.Values["UID"] = int32(row.Idusers)
	session.Values["LoginTime"] = time.Now().Unix()
	session.Values["ExpiryTime"] = time.Now().AddDate(1, 0, 0).Unix()

	backURL, _ := r.Context().Value(consts.KeyCoreData).(*common.CoreData).SanitizeBackURL(r, r.FormValue("back"))
	backMethod := r.FormValue("method")
	backData := r.FormValue("data")

	if err := session.Save(r, w); err != nil {
		return fmt.Errorf("session save %w", err)
	}

	if cd.Config.LogFlags&config.LogFlagAuth != 0 {
		log.Printf("login success uid=%d session=%s", row.Idusers, handlers.HashSessionID(session.ID))
	}

	if backURL != "" {
		if backMethod == "" || backMethod == http.MethodGet {
			return handlers.RefreshDirectHandler{TargetURL: backURL}
		}
		var vals url.Values
		if backData != "" {
			if dec, err := cd.DecryptData(backData); err == nil {
				vals, _ = url.ParseQuery(dec)
			} else {
				log.Printf("decrypt back data: %v", err)
			}
		}
		return redirectBackPageHandler{BackURL: backURL, Method: backMethod, Values: vals}
	}

	return handlers.RefreshDirectHandler{TargetURL: "/"}
}

// RequiredTemplates declares the templates used by this task's pages.
func (LoginTask) RequiredTemplates() []tasks.Template {
	return []tasks.Template{
		tasks.Template(templateLoginPage),
		tasks.Template(templatePasswordVerifyPage),
	}
}

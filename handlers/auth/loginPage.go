package auth

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	db "github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	common "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
	handlers "github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

// LoginTask handles rendering and processing of the login form.
type LoginTask struct {
	tasks.TaskString
}

var loginTask = &LoginTask{TaskString: TaskLogin}

// ensure LoginTask conforms to tasks.Task
var _ tasks.Task = (*LoginTask)(nil)

// VerifyPasswordTask verifies reset codes during login.
type VerifyPasswordTask struct {
	tasks.TaskString
}

var verifyPasswordTask = &VerifyPasswordTask{TaskString: TaskPasswordVerify}

// ensure VerifyPasswordTask conforms to tasks.Task
var _ tasks.Task = (*VerifyPasswordTask)(nil)

func renderLoginForm(w http.ResponseWriter, r *http.Request, errMsg string) {
	type Data struct {
		*common.CoreData
		Error string
	}
	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Error:    errMsg,
	}
	handlers.TemplateHandler(w, r, "loginPage.gohtml", data)
}

// LoginUserPassPage serves the username/password login form.
func (LoginTask) Page(w http.ResponseWriter, r *http.Request) {
	renderLoginForm(w, r, r.URL.Query().Get("error"))
}

// LoginActionPage processes the submitted login form.
func (LoginTask) Action(w http.ResponseWriter, r *http.Request) {
	if config.AppRuntimeConfig.LogFlags&config.LogFlagAuth != 0 {
		sess, _ := core.GetSession(r)
		log.Printf("login attempt for %s session=%s", r.PostFormValue("username"), sess.ID)
	}
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	row, err := queries.Login(r.Context(), sql.NullString{String: username, Valid: true})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			log.Printf("No rows Error: %s", err)
			_ = queries.InsertLoginAttempt(r.Context(), db.InsertLoginAttemptParams{
				Username:  username,
				IpAddress: strings.Split(r.RemoteAddr, ":")[0],
			})
			renderLoginForm(w, r, "No such user")
			return
		default:
			log.Printf("query Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	if !VerifyPassword(password, row.Passwd.String, row.PasswdAlgorithm.String) {
		reset, err := queries.GetPasswordResetByUser(r.Context(), row.Idusers)
		if err == nil && VerifyPassword(password, reset.Passwd, reset.PasswdAlgorithm) {
			session, ok := core.GetSessionOrFail(w, r)
			if !ok {
				return
			}
			session.Values["PendingResetID"] = reset.ID
			_ = session.Save(r, w)
			handlers.TemplateHandler(w, r, "passwordVerifyPage.gohtml", struct{ *common.CoreData }{r.Context().Value(consts.KeyCoreData).(*common.CoreData)})
			return
		}
		_ = queries.InsertLoginAttempt(r.Context(), db.InsertLoginAttemptParams{
			Username:  username,
			IpAddress: strings.Split(r.RemoteAddr, ":")[0],
		})
		renderLoginForm(w, r, "Invalid password")
		return
	}

	if _, err := queries.UserHasRole(r.Context(), db.UserHasRoleParams{UsersIdusers: row.Idusers, Name: "user"}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			renderLoginForm(w, r, "approval is pending")
		} else {
			log.Printf("user role: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	if row.PasswdAlgorithm.String == "" || row.PasswdAlgorithm.String == "md5" {
		newHash, newAlg, err := HashPassword(password)
		if err == nil {
			_ = queries.InsertPassword(r.Context(), db.InsertPasswordParams{UsersIdusers: row.Idusers, Passwd: newHash, PasswdAlgorithm: sql.NullString{String: newAlg, Valid: true}})
		}
	}

	user := &db.User{Idusers: row.Idusers}

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	session.Values["UID"] = int32(user.Idusers)
	session.Values["LoginTime"] = time.Now().Unix()
	session.Values["ExpiryTime"] = time.Now().AddDate(1, 0, 0).Unix()

	backURL, _ := session.Values["BackURL"].(string)
	backMethod, _ := session.Values["BackMethod"].(string)
	backData, _ := session.Values["BackData"].(string)
	delete(session.Values, "BackURL")
	delete(session.Values, "BackMethod")
	delete(session.Values, "BackData")

	if err := session.Save(r, w); err != nil {
		log.Printf("session.Save Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if config.AppRuntimeConfig.LogFlags&config.LogFlagAuth != 0 {
		log.Printf("login success uid=%d session=%s", user.Idusers, session.ID)
	}

	if backURL != "" {
		if backMethod == "" || backMethod == http.MethodGet {
			http.Redirect(w, r, backURL, http.StatusTemporaryRedirect)
			return
		}
		vals, _ := url.ParseQuery(backData)
		type Data struct {
			*common.CoreData
			BackURL string
			Method  string
			Values  url.Values
		}
		data := Data{
			CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
			BackURL:  backURL,
			Method:   backMethod,
			Values:   vals,
		}
		cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
		if err := templates.GetCompiledSiteTemplates(cd.Funcs(r)).ExecuteTemplate(w, "redirectBackPage.gohtml", data); err != nil {
			log.Printf("Template Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (VerifyPasswordTask) Action(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	id, _ := session.Values["PendingResetID"].(int32)
	if id == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	code := r.FormValue("code")
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	reset, err := queries.GetPasswordResetByCode(r.Context(), code)
	if err != nil || reset.ID != id {
		http.Error(w, "invalid code", http.StatusUnauthorized)
		return
	}
	_ = queries.MarkPasswordResetVerified(r.Context(), reset.ID)
	_ = queries.InsertPassword(r.Context(), db.InsertPasswordParams{UsersIdusers: reset.UserID, Passwd: reset.Passwd, PasswdAlgorithm: sql.NullString{String: reset.PasswdAlgorithm, Valid: true}})
	delete(session.Values, "PendingResetID")
	_ = session.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

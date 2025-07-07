package auth

import (
	"database/sql"
	"errors"
	db "github.com/arran4/goa4web/internal/db"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/arran4/goa4web/core"
	corecommon "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
	common "github.com/arran4/goa4web/handlers/common"
	"github.com/arran4/goa4web/runtimeconfig"
)

// LoginUserPassPage serves the username/password login form.
func LoginUserPassPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*corecommon.CoreData
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*corecommon.CoreData),
	}

	if err := templates.RenderTemplate(w, "loginPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// LoginActionPage processes the submitted login form.
func LoginActionPage(w http.ResponseWriter, r *http.Request) {
	if runtimeconfig.AppRuntimeConfig.LogFlags&runtimeconfig.LogFlagAuth != 0 {
		log.Printf("login attempt for %s", r.PostFormValue("username"))
	}
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")

	//sum := md5.Sum([]byte(password))
	//
	//hashedPassword := hex.EncodeToString(sum[:])

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

	row, err := queries.Login(r.Context(), sql.NullString{String: username, Valid: true})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			log.Printf("No rows Error: %s", err)
			_ = queries.InsertLoginAttempt(r.Context(), db.InsertLoginAttemptParams{
				Username:  username,
				IpAddress: strings.Split(r.RemoteAddr, ":")[0],
			})
			http.Error(w, "No such user", http.StatusNotFound)
			return
		default:
			log.Printf("query Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	if !verifyPassword(password, row.Passwd.String, row.PasswdAlgorithm.String) {
		_ = queries.InsertLoginAttempt(r.Context(), db.InsertLoginAttemptParams{
			Username:  username,
			IpAddress: strings.Split(r.RemoteAddr, ":")[0],
		})
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	if row.PasswdAlgorithm.String == "" || row.PasswdAlgorithm.String == "md5" {
		newHash, newAlg, err := hashPassword(password)
		if err == nil {
			_ = queries.InsertPassword(r.Context(), db.InsertPasswordParams{UsersIdusers: row.Idusers, Passwd: newHash, PasswdAlgorithm: sql.NullString{String: newAlg, Valid: true}})
		}
	}

	user := &db.User{Idusers: row.Idusers, Email: row.Email}

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

	if runtimeconfig.AppRuntimeConfig.LogFlags&runtimeconfig.LogFlagAuth != 0 {
		log.Printf("login success uid=%d", user.Idusers)
	}

	if backURL != "" {
		if backMethod == "" || backMethod == http.MethodGet {
			http.Redirect(w, r, backURL, http.StatusTemporaryRedirect)
			return
		}
		vals, _ := url.ParseQuery(backData)
		type Data struct {
			*corecommon.CoreData
			BackURL string
			Method  string
			Values  url.Values
		}
		data := Data{
			CoreData: r.Context().Value(common.KeyCoreData).(*corecommon.CoreData),
			BackURL:  backURL,
			Method:   backMethod,
			Values:   vals,
		}
		if err := templates.GetCompiledTemplates(corecommon.NewFuncs(r)).ExecuteTemplate(w, "redirectBackPage.gohtml", data); err != nil {
			log.Printf("Template Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

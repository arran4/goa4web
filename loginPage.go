package goa4web

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/templates"
)

func loginUserPassPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	if err := templates.RenderTemplate(w, "loginPage.gohtml", data, common.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func loginActionPage(w http.ResponseWriter, r *http.Request) {
	log.Printf("login attempt for %s", r.PostFormValue("username"))
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")

	//sum := md5.Sum([]byte(password))
	//
	//hashedPassword := hex.EncodeToString(sum[:])

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	var (
		uid    int32
		email  sql.NullString
		hashed sql.NullString
		alg    sql.NullString
	)
	err := queries.DB().QueryRowContext(r.Context(),
		"SELECT idusers, email, passwd, IFNULL(passwd_algorithm,''), username FROM users WHERE username = ?",
		username,
	).Scan(&uid, &email, &hashed, &alg, new(string))
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			log.Printf("No rows Error: %s", err)
			_ = queries.InsertLoginAttempt(r.Context(), InsertLoginAttemptParams{
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

	if !verifyPassword(password, hashed.String, alg.String) {
		_ = queries.InsertLoginAttempt(r.Context(), InsertLoginAttemptParams{
			Username:  username,
			IpAddress: strings.Split(r.RemoteAddr, ":")[0],
		})
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	if alg.String == "" || alg.String == "md5" {
		newHash, newAlg, err := hashPassword(password)
		if err == nil {
			_, _ = queries.DB().ExecContext(r.Context(),
				"UPDATE users SET passwd=?, passwd_algorithm=? WHERE idusers=?",
				newHash, newAlg, uid,
			)
		}
	}

	user := &User{Idusers: uid, Email: email}

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

	log.Printf("login success uid=%d", user.Idusers)

	if backURL != "" {
		if backMethod == "" || backMethod == http.MethodGet {
			http.Redirect(w, r, backURL, http.StatusTemporaryRedirect)
			return
		}
		vals, _ := url.ParseQuery(backData)
		type Data struct {
			*CoreData
			BackURL string
			Method  string
			Values  url.Values
		}
		data := Data{
			CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
			BackURL:  backURL,
			Method:   backMethod,
			Values:   vals,
		}
		if err := templates.GetCompiledTemplates(common.NewFuncs(r)).ExecuteTemplate(w, "redirectBackPage.gohtml", data); err != nil {
			log.Printf("Template Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

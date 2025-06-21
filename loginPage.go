package main

import (
	"database/sql"
	"errors"
	"github.com/gorilla/csrf"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"time"
)

func loginUserPassPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		CSRFField template.HTML
	}

	data := Data{
		CoreData:  r.Context().Value(ContextValues("coreData")).(*CoreData),
		CSRFField: csrf.TemplateField(r),
	}

	if err := renderTemplate(w, r, "loginPage.gohtml", data); err != nil {
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

	user, err := queries.Login(r.Context(), LoginParams{
		Username: sql.NullString{String: username, Valid: true},
		MD5:      password,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			log.Printf("No rows Error: %s", err)
			http.Error(w, "No such user", http.StatusNotFound)
			return
		default:
			log.Printf("getCommentsByThreadIdForUser Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	session, ok := GetSessionOrFail(w, r)
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
		if err := getCompiledTemplates(NewFuncs(r)).ExecuteTemplate(w, "redirectBackPage.gohtml", data); err != nil {
			log.Printf("Template Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

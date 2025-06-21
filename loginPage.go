package main

import (
	"database/sql"
	"errors"
	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
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

	if err := getCompiledTemplates(NewFuncs(r)).ExecuteTemplate(w, "loginPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func loginActionPage(w http.ResponseWriter, r *http.Request) {
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

	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	session.Values["UID"] = int32(user.Idusers)
	session.Values["LoginTime"] = time.Now().Unix()
	session.Values["ExpiryTime"] = time.Now().AddDate(1, 0, 0).Unix()

	returnPath, _ := session.Values["return_path"].(string)
	returnMethod, _ := session.Values["return_method"].(string)
	returnForm, _ := session.Values["return_form"].(string)

	delete(session.Values, "return_path")
	delete(session.Values, "return_method")
	delete(session.Values, "return_form")

	if err := session.Save(r, w); err != nil {
		log.Printf("session.Save Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if returnPath != "" {
		if returnMethod == "" || returnMethod == http.MethodGet || returnMethod == http.MethodHead {
			http.Redirect(w, r, returnPath, http.StatusTemporaryRedirect)
		} else {
			type Data struct {
				*CoreData
				Path   string
				Method string
				Form   url.Values
			}
			vals, _ := url.ParseQuery(returnForm)
			data := Data{
				CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
				Path:     returnPath,
				Method:   returnMethod,
				Form:     vals,
			}
			if err := getCompiledTemplates(NewFuncs(r)).ExecuteTemplate(w, "goBackPage.gohtml", data); err != nil {
				log.Printf("Template Error: %s", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}
		return
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

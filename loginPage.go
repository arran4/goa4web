package main

import (
	"database/sql"
	"errors"
	"github.com/gorilla/csrf"
	"html/template"
	"log"
	"net/http"
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

	session, _ := GetSession(r)
	session.Values["UID"] = int32(user.Idusers)
	session.Values["LoginTime"] = time.Now().Unix()
	session.Values["ExpiryTime"] = time.Now().AddDate(1, 0, 0).Unix()

	if err := session.Save(r, w); err != nil {
		log.Printf("session.Save Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

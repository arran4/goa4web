package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"time"
)

func loginUserPassPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	if err := getCompiledTemplates().ExecuteTemplate(w, "loginPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	delete(session.Values, "UID")
	delete(session.Values, "LoginTime")
	delete(session.Values, "ExpiryTime")

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func loginActionPage(w http.ResponseWriter, r *http.Request) {
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")

	sum := md5.Sum([]byte(password))

	hashedPassword := hex.EncodeToString(sum[:])

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	user, err := queries.Login(r.Context(), LoginParams{
		Username: sql.NullString{String: username, Valid: true},
		MD5:      hashedPassword,
	})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	session.Values["UID"] = int32(user.Idusers)
	session.Values["LoginTime"] = time.Now().Unix()
	session.Values["ExpiryTime"] = time.Now().AddDate(1, 0, 0).Unix()

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

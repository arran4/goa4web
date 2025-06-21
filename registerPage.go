package main

import (
	"database/sql"
	"errors"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"strings"
	"time"
)

func registerPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
	}

	renderTemplate(w, r, "registerPage.gohtml", data)
}
func registerActionPage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm Error: %s", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	uVals, uOK := r.PostForm["username"]
	pVals, pOK := r.PostForm["password"]
	eVals, eOK := r.PostForm["email"]
	if !uOK || len(uVals) == 0 || uVals[0] == "" ||
		!pOK || len(pVals) == 0 || pVals[0] == "" ||
		!eOK || len(eVals) == 0 || eVals[0] == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}
	username := uVals[0]
	password := pVals[0]
	email := eVals[0]
	if !strings.Contains(email, "@") {
		http.Error(w, "invalid email", http.StatusBadRequest)
		return
	}
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	if _, err := queries.UserByUsername(r.Context(), sql.NullString{
		String: username,
		Valid:  true,
	}); errors.Is(err, sql.ErrNoRows) {
	} else if err != nil {
		log.Printf("UserByUsername Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	} else {
		http.Error(w, "User already exists", http.StatusForbidden)
		return
	}

	if _, err := queries.UserByEmail(r.Context(), sql.NullString{
		String: email,
		Valid:  true,
	}); errors.Is(err, sql.ErrNoRows) {
	} else if err != nil {
		log.Printf("UserByUsername Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	} else {
		http.Error(w, "User already exists", http.StatusForbidden)
		return
	}
	//sum := md5.Sum([]byte(password))

	//hashedPassword := hex.EncodeToString(sum[:])

	result, err := queries.InsertUser(r.Context(), InsertUserParams{
		Username: sql.NullString{
			Valid:  true,
			String: username,
		},
		MD5: password,
		Email: sql.NullString{
			Valid:  true,
			String: email,
		},
	})
	if err != nil {
		log.Printf("InsertUser Error: %s", err)
		http.Error(w, "Can't create user", http.StatusForbidden)
		return
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		log.Printf("LastInsertId Error: %s", err)
		http.Error(w, "Session error", http.StatusForbidden)
		return
	}

	session := r.Context().Value(ContextValues("session")).(*sessions.Session)
	session.Values["UID"] = int32(lastInsertID)
	session.Values["LoginTime"] = time.Now().Unix()
	session.Values["ExpiryTime"] = time.Now().AddDate(1, 0, 0).Unix()

	if err := session.Save(r, w); err != nil {
		log.Printf("session.Save Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)

}

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

// GetSession returns the session from the request context if present,
// otherwise it retrieves the session from the session store.
func GetSession(r *http.Request) (*sessions.Session, error) {
	if sessVal := r.Context().Value(ContextValues("session")); sessVal != nil {
		sess, ok := sessVal.(*sessions.Session)
		if !ok {
			return nil, fmt.Errorf("invalid session in context")
		}
		return sess, nil
	}
	sess, err := store.Get(r, sessionName)
	if err != nil {
		log.Printf("get session: %v", err)
	}
	return sess, err
}

// clearSession removes the session cookie so the user is effectively logged out.
func clearSession(w http.ResponseWriter, r *http.Request) {
	sess, _ := store.New(r, sessionName)
	sess.Options.MaxAge = -1
	if err := sess.Save(r, w); err != nil {
		log.Printf("clear session: %v", err)
	}
}

// GetSessionOrFail wraps GetSession and writes a 500 response if retrieving the
// session fails. It returns the session and a boolean indicating success.
func GetSessionOrFail(w http.ResponseWriter, r *http.Request) (*sessions.Session, bool) {
	sess, err := GetSession(r)
	if err != nil {
		log.Printf("session error: %v", err)
		clearSession(w, r)
		http.Redirect(w, r, "/login", http.StatusFound)
		return nil, false
	}
	return sess, true
}

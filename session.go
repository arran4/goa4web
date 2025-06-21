package main

import (
	"fmt"
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
	return store.Get(r, sessionName)
}

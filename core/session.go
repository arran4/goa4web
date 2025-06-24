package core

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

const sessionName = "my-session"

var store *sessions.CookieStore

// SetSessionStore configures the session store used by GetSession.
func SetSessionStore(s *sessions.CookieStore) {
	store = s
}

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

func clearSession(w http.ResponseWriter, r *http.Request) {
	sess, _ := store.New(r, sessionName)
	sess.Options.MaxAge = -1
	if err := sess.Save(r, w); err != nil {
		log.Printf("clear session: %v", err)
	}
}

// sessionError logs the error and clears the session cookie.
func sessionError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("session error: %v", err)
	clearSession(w, r)
}

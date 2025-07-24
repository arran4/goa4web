package core

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

// SessionManager manages web session retrieval and error handling.
type SessionManager struct {
	// Name is the cookie name used for storing session data.
	Name string
	// Store provides the backing cookie store implementation.
	Store *sessions.CookieStore
}

// GetSession returns the session from the request context if present,
// otherwise it retrieves the session from the session store.
func (sm *SessionManager) GetSession(r *http.Request) (*sessions.Session, error) {
	if sessVal := r.Context().Value(ContextValues("session")); sessVal != nil {
		sess, ok := sessVal.(*sessions.Session)
		if !ok {
			return nil, fmt.Errorf("invalid session in context")
		}
		return sess, nil
	}
	sess, err := sm.Store.Get(r, sm.Name)
	if err != nil {
		log.Printf("get session: %v", err)
	}
	return sess, err
}

// clearSession removes the session cookie so the user is effectively logged out.
func (sm *SessionManager) clearSession(w http.ResponseWriter, r *http.Request) {
	sess, _ := sm.Store.New(r, sm.Name)
	sess.Options.MaxAge = -1
	if err := sess.Save(r, w); err != nil {
		log.Printf("clear session: %v", err)
	}
}

// GetSessionOrFail wraps GetSession and writes a 500 response if retrieving the
// session fails. It returns the session and a boolean indicating success.
func (sm *SessionManager) GetSessionOrFail(w http.ResponseWriter, r *http.Request) (*sessions.Session, bool) {
	sess, err := sm.GetSession(r)
	if err != nil {
		sm.SessionErrorRedirect(w, r, err)
		return nil, false
	}
	return sess, true
}

// SessionErrorRedirect clears the session and redirects to the login page when
// an error occurs retrieving the session.
func (sm *SessionManager) SessionErrorRedirect(w http.ResponseWriter, r *http.Request, err error) {
	sm.SessionError(w, r, err)
	http.Redirect(w, r, "/login", http.StatusFound)
}

// SessionError logs the error and clears the session cookie.
func (sm *SessionManager) SessionError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("session error: %v", err)
	sm.clearSession(w, r)
}

// ContextValues represents context key names used across the application.
type ContextValues string

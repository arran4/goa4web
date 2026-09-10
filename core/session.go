package core

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/sessions"
)

var SessionName string
var Store *sessions.CookieStore

func GetSession(r *http.Request) (*sessions.Session, error) {
	if sessVal := r.Context().Value(ContextValues("session")); sessVal != nil {
		sess, ok := sessVal.(*sessions.Session)
		if !ok {
			return nil, fmt.Errorf("invalid session in context")
		}
		return sess, nil
	}
	sess, err := Store.Get(r, SessionName)
	if err != nil {
		log.Printf("get session: %v", err)
	}
	return sess, err
}

func clearSession(w http.ResponseWriter, r *http.Request) {
	sess, _ := Store.New(r, SessionName)
	sess.Options.MaxAge = -1
	if err := sess.Save(r, w); err != nil {
		log.Printf("clear session: %v", err)
	}
}

func GetSessionOrFail(w http.ResponseWriter, r *http.Request) (*sessions.Session, bool) {
	sess, err := GetSession(r)
	if err != nil {
		SessionErrorRedirect(w, r, err)
		return nil, false
	}
	return sess, true
}

func SessionErrorRedirect(w http.ResponseWriter, r *http.Request, err error) {
	SessionError(w, r, err)
	RedirectToLogin(w, r, nil)
}

func SessionError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("session error: %v", err)
	clearSession(w, r)
}

func DisableCaching(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Cloudflare-CDN-Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}

func RedirectToLogin(w http.ResponseWriter, r *http.Request, session *sessions.Session) {
	DisableCaching(w)
	if session != nil {
		if err := session.Save(r, w); err != nil {
			log.Printf("save session: %v", err)
		}
	}

	back := r.URL.RequestURI()
	if r.URL.Path == "/login" {
		back = safeLoginContinuation(r, r.URL.Query().Get("back"))
	}

	newVals := url.Values{}
	if back != "" {
		newVals.Set("back", back)
	}
	target := "/login"
	if query := newVals.Encode(); query != "" {
		target += "?" + query
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}

func safeLoginContinuation(r *http.Request, raw string) string {
	if raw == "" {
		return ""
	}
	back, err := url.Parse(raw)
	if err != nil || back.IsAbs() || back.Host != "" {
		return ""
	}
	resolved := r.URL.ResolveReference(back)
	if resolved.Path == "/login" || strings.HasPrefix(resolved.Path, "/login/") {
		return ""
	}
	return raw
}

type ContextValues string

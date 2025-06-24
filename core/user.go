package core

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"net/url"
	"time"

	db "github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/sessions"
)

// UserAdderMiddleware loads the current user's information from the database
// and attaches it along with preferences and permissions to the request context.
func UserAdderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// Get the session.
		session, err := GetSession(request)
		if err != nil {
			sessionError(writer, request, err)
		}

		queries := request.Context().Value(ContextValues("queries")).(*db.Queries)
		var (
			user        *db.User
			permissions []*db.Permission
			preference  *db.Preference
			languages   []*db.Userlang
			uid         int32
		)
		if uidi, ok := session.Values["UID"]; ok {
			if v, ok := uidi.(int32); ok {
				uid = v
			}

			if expi, ok := session.Values["ExpiryTime"]; ok {
				var exp int64
				switch t := expi.(type) {
				case int64:
					exp = t
				case int:
					exp = int64(t)
				case float64:
					exp = int64(t)
				}
				if exp != 0 && time.Now().Unix() > exp {
					delete(session.Values, "UID")
					delete(session.Values, "LoginTime")
					delete(session.Values, "ExpiryTime")
					RedirectToLogin(writer, request, session)
					return
				}
			}

			if uid != 0 {
				if user, err = queries.GetUserById(request.Context(), uid); err != nil {
					switch {
					case errors.Is(err, sql.ErrNoRows):
					default:
						log.Printf("Error: GetUserById: %s", err)
						http.Redirect(writer, request, "?error="+err.Error(), http.StatusTemporaryRedirect)
						return
					}
				} else {
					permissions, _ = queries.GetPermissionsByUserID(request.Context(), uid)
					preference, _ = queries.GetPreferenceByUserID(request.Context(), uid)
					languages, _ = queries.GetUserLanguages(request.Context(), uid)
				}
			}
		}

		if session.ID != "" {
			if uid != 0 {
				_ = queries.InsertSession(request.Context(), db.InsertSessionParams{SessionID: session.ID, UsersIdusers: uid})
			} else {
				_ = queries.DeleteSessionByID(request.Context(), session.ID)
			}
		}

		ctx := context.WithValue(request.Context(), ContextValues("session"), session)
		ctx = context.WithValue(ctx, ContextValues("user"), user)
		ctx = context.WithValue(ctx, ContextValues("permissions"), permissions)
		ctx = context.WithValue(ctx, ContextValues("preference"), preference)
		ctx = context.WithValue(ctx, ContextValues("languages"), languages)
		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}

// RedirectToLogin stores the current request so the user can be redirected back
// after logging in, then issues the redirect to /login.
func RedirectToLogin(w http.ResponseWriter, r *http.Request, session *sessions.Session) {
	if session != nil {
		backURL := r.URL.RequestURI()
		session.Values["BackURL"] = backURL
		if r.Method != http.MethodGet {
			if err := r.ParseForm(); err == nil {
				session.Values["BackMethod"] = r.Method
				session.Values["BackData"] = r.Form.Encode()
			}
		} else {
			delete(session.Values, "BackMethod")
			delete(session.Values, "BackData")
		}
		_ = session.Save(r, w)
	}
	http.Redirect(w, r, "/login?back="+url.QueryEscape(r.URL.RequestURI()), http.StatusTemporaryRedirect)
}

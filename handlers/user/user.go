package user

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
)

func UserAdderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// Get the session.
		session, err := core.GetSession(request)
		if err != nil {
			core.SessionError(writer, request, err)
		}

		queries := request.Context().Value(common.KeyQueries).(*db.Queries)
		var (
			user        *db.User
			permissions []*db.Permission
			preference  *db.Preference
			languages   []*db.UserLanguage
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
					redirectToLogin(writer, request, session)
					return
				}
			}

			if uid != 0 {
				var row *db.GetUserByIdRow
				if row, err = queries.GetUserById(request.Context(), uid); err != nil {
					switch {
					case errors.Is(err, sql.ErrNoRows):
					default:
						log.Printf("Error: GetUserById: %s", err)
						http.Redirect(writer, request, "?error="+err.Error(), http.StatusTemporaryRedirect)
						return
					}
				} else {
					user = &db.User{Idusers: row.Idusers, Email: row.Email, Username: row.Username}
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

		ctx := context.WithValue(request.Context(), common.KeySession, session)
		ctx = context.WithValue(ctx, common.KeyUser, user)
		ctx = context.WithValue(ctx, common.KeyPermissions, permissions)
		ctx = context.WithValue(ctx, common.KeyPreference, preference)
		ctx = context.WithValue(ctx, common.KeyLanguages, languages)
		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}

func redirectToLogin(w http.ResponseWriter, r *http.Request, session *sessions.Session) {
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

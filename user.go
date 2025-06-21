package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"time"
)

func UserAdderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// Get the session. If the session cannot be loaded (for example due to
		// invalid cookie data), start a fresh session and redirect the user to
		// the login page.
		session, err := store.Get(request, sessionName)
		if err != nil {
			log.Printf("invalid session: %v", err)
			session, _ = store.New(request, sessionName)
			// remember where the user was going
			session.Values["return_path"] = request.URL.RequestURI()
			session.Values["return_method"] = request.Method
			if request.Method != http.MethodGet && request.Method != http.MethodHead {
				request.ParseForm()
				session.Values["return_form"] = request.PostForm.Encode()
			}
			delete(session.Values, "UID")
			delete(session.Values, "LoginTime")
			delete(session.Values, "ExpiryTime")
			_ = session.Save(request, writer)
			http.Redirect(writer, request, "/login", http.StatusTemporaryRedirect)
			return
		}

		queries := request.Context().Value(ContextValues("queries")).(*Queries)
		var (
			user        *User
			permissions []*Permission
			preference  *Preference
			languages   []*Userlang
		)
		if uidi, ok := session.Values["UID"]; ok {
			var uid int32
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
					session.Values["return_path"] = request.URL.RequestURI()
					session.Values["return_method"] = request.Method
					if request.Method != http.MethodGet && request.Method != http.MethodHead {
						request.ParseForm()
						session.Values["return_form"] = request.PostForm.Encode()
					}
					delete(session.Values, "UID")
					delete(session.Values, "LoginTime")
					delete(session.Values, "ExpiryTime")
					_ = session.Save(request, writer)
					http.Redirect(writer, request, "/login", http.StatusTemporaryRedirect)
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

		ctx := context.WithValue(request.Context(), ContextValues("session"), session)
		ctx = context.WithValue(ctx, ContextValues("user"), user)
		ctx = context.WithValue(ctx, ContextValues("permissions"), permissions)
		ctx = context.WithValue(ctx, ContextValues("preference"), preference)
		ctx = context.WithValue(ctx, ContextValues("languages"), languages)
		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}

package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"
)

func UserAdderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// Get the session.
		session, err := store.Get(request, sessionName)
		if err != nil {
			http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		queries := request.Context().Value(ContextValues("queries")).(*Queries)

		task := request.FormValue("task")
		username := request.FormValue("username")
		password := request.FormValue("password")
		if task == "Login" && username != "" || password != "" {
			UID, err := loginUser(request.Context(), queries, username, password)
			if err != nil {
				http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if UID != 0 {
				session.Values["UID"] = UID
				session.Values["LoginTime"] = time.Now().Unix()
				session.Values["ExpiryTime"] = time.Now().AddDate(1, 0, 0).Unix()
			}
		}

		// TODO once login works.
		session.Values["UID"] = int32(1)
		session.Values["LoginTime"] = time.Now().Unix()
		session.Values["ExpiryTime"] = time.Now().AddDate(1, 0, 0).Unix()

		// TODO session.Values["ExpiryTime"]

		// TODO inject user security levels / scopes into session and context
		/*

			if (section == NULL)
				section = "all";
			a4string query("SELECT level FROM permissions "
					"WHERE users_idusers=\"%d\" AND (section=\"%s\" OR section=\"all\")",
					cont.user.UID, section);
			a4mysqlResult *result = cont.sql.query(query.raw());
			enum authlevels returnthis = auth_reader;
			if (result->hasRow())
			{
				char *tmp = result->getColumn(0);
				if (!strcasecmp(tmp, "reader"))
					returnthis = auth_reader;
				else if (!strcasecmp(tmp, "writer"))
					returnthis = auth_writer;
				else if (!strcasecmp(tmp, "moderator"))
					returnthis = auth_moderator;
				else if (!strcasecmp(tmp, "administrator"))
					returnthis = auth_administrator;
			}
			delete result;
			return returnthis;


		*/
		// TODO inject user preferences into session and context
		/*

				a4string query("SELECT language_idlanguage FROM preferences WHERE users_idusers=\"%d\"", user.UID);
			a4mysqlResult *result = sql.query(query.raw());
			if (result->hasRow())
			{
				defaultLang = atoiornull(result->getColumn(0));
				while (result->nextRow());
			}
			delete result;
			query.set("SELECT language_idlanguage FROM userlang WHERE users_idusers=\"%d\"", user.UID);
			result = sql.query(query.raw());
			if (result->hasRow())
			{
				setlangMod++;
				langMod.push("language_idlanguage IN (");
				langMod.pushf("0");
				do
					langMod.pushf(", %s", result->getColumn(0));
				while (result->nextRow());
				langMod.push(") ");
			}
			delete result;

		*/

		ctx := context.WithValue(request.Context(), ContextValues("session"), session)
		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}

func loginUser(ctx context.Context, queries *Queries, username, password string) (int32, error) {
	userID, err := queries.Login(ctx, LoginParams{
		Username: sql.NullString{
			String: username,
			Valid:  true,
		},
		MD5: password,
	})
	if err != nil {
		return -1, fmt.Errorf("loginUser: %w", err)
	}
	return userID, nil
}

//func verifyUser(ctx context.Context, w http.ResponseWriter, queries *Queries, sid string) (int, error) {
//	if sid == "" {
//		t := time.Now().AddDate(1, 0, 0)
//		var err error
//		sid, err = generateSID(ctx, queries)
//		if err != nil {
//			return 0, fmt.Errorf("verifyUser: %w", err)
//		}
//		http.SetCookie(w, &http.Cookie{
//			Name:    "SID",
//			Name:   sid,
//			Expires: t,
//		})
//	}
//
//	user, err := queries.SelectUserBySID(ctx, sql.NullString{
//		String: sid,
//		Valid:  true,
//	})
//	if err != nil {
//		return 0, fmt.Errorf("verifyUser: %w", err)
//	}
//
//	if user.Logintime.Time.AddDate(1, 0, 0).After(time.Now()) {
//		return int(user.UsersIdusers), nil
//	}
//
//	return 0, nil
//}

func registerUser(ctx context.Context, queries *Queries, username, password, email string) (int, error) {
	userID, err := queries.Login(ctx, LoginParams{
		Username: sql.NullString{
			String: username,
			Valid:  true,
		},
		MD5: password,
	})
	if userID != 0 || err == nil {
		return -3, fmt.Errorf("registerUser: %w", err)
	}

	result, err := queries.InsertUser(ctx, InsertUserParams{
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
		return -3, err // SQL error.
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return -3, fmt.Errorf("registerUser: %w", err) // SQL error.
	}

	return int(lastInsertID), nil
}

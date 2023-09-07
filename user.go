package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
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
		var user *User
		if uidi, ok := session.Values["UID"]; !ok {
		} else if uid, ok := uidi.(int32); !ok {
		} else if user, err = queries.GetUserById(request.Context(), uid); err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
			default:
				log.Printf("Error: GetUserById: %s", err)
				http.Redirect(writer, request, "?error="+err.Error(), http.StatusTemporaryRedirect)
			}
		}

		// TODO session.Values["ExpiryTime"]
		//	if user.Logintime.Time.AddDate(1, 0, 0).After(time.Now()) {
		//if err := session.Save(r, w); err != nil {
		//delete(session.Values, "UID")
		//delete(session.Values, "LoginTime")
		//delete(session.Values, "ExpiryTime")
		//	log.Printf("session.Save Error: %s", err)
		//	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		//	return
		//	}
		//}

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
		ctx = context.WithValue(ctx, ContextValues("user"), user)
		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}

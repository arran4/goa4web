package auth

import (
	"database/sql"
	"errors"
	db "github.com/arran4/goa4web/internal/db"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/arran4/goa4web/core"
	corecommon "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
	common "github.com/arran4/goa4web/handlers/common"
)

// RegisterPage renders the user registration form.
func RegisterPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*corecommon.CoreData
	}

	data := Data{
		CoreData: r.Context().Value(common.KeyCoreData).(*corecommon.CoreData),
	}

	if err := templates.RenderTemplate(w, "registerPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// RegisterActionPage handles user creation from the registration form.
func RegisterActionPage(w http.ResponseWriter, r *http.Request) {
	log.Printf("registration attempt %s", r.PostFormValue("username"))
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm Error: %s", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	uVals, uOK := r.PostForm["username"]
	pVals, pOK := r.PostForm["password"]
	eVals, eOK := r.PostForm["email"]
	if !uOK || len(uVals) == 0 || uVals[0] == "" ||
		!pOK || len(pVals) == 0 || pVals[0] == "" ||
		!eOK || len(eVals) == 0 || eVals[0] == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}
	username := uVals[0]
	password := pVals[0]
	email := eVals[0]
	if !strings.Contains(email, "@") {
		http.Error(w, "invalid email", http.StatusBadRequest)
		return
	}
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

	if _, err := queries.UserByUsername(r.Context(), sql.NullString{
		String: username,
		Valid:  true,
	}); errors.Is(err, sql.ErrNoRows) {
	} else if err != nil {
		log.Printf("UserByUsername Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	} else {
		http.Error(w, "User already exists", http.StatusForbidden)
		return
	}

	if _, err := queries.UserByEmail(r.Context(), sql.NullString{
		String: email,
		Valid:  true,
	}); errors.Is(err, sql.ErrNoRows) {
	} else if err != nil {
		log.Printf("UserByUsername Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	} else {
		http.Error(w, "User already exists", http.StatusForbidden)
		return
	}
	//sum := md5.Sum([]byte(password))

	//hashedPassword := hex.EncodeToString(sum[:])

	hash, alg, err := hashPassword(password)
	if err != nil {
		log.Printf("hashPassword Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	result, err := queries.DB().ExecContext(r.Context(),
		"INSERT INTO users (username, passwd, passwd_algorithm, email) VALUES (?, ?, ?, ?)",
		username, hash, alg, email,
	)
	if err != nil {
		log.Printf("InsertUser Error: %s", err)
		http.Error(w, "Can't create user", http.StatusForbidden)
		return
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		log.Printf("LastInsertId Error: %s", err)
		http.Error(w, "Session error", http.StatusForbidden)
		return
	}

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	session.Values["UID"] = int32(lastInsertID)
	session.Values["LoginTime"] = time.Now().Unix()
	session.Values["ExpiryTime"] = time.Now().AddDate(1, 0, 0).Unix()

	if err := session.Save(r, w); err != nil {
		log.Printf("session.Save Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Printf("registration success uid=%d", lastInsertID)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)

}

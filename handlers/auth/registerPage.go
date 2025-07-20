package auth

import (
	"database/sql"
	"errors"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"strings"

	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/notifications"

	"github.com/arran4/goa4web/config"
	common "github.com/arran4/goa4web/core/common"
	handlers "github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
)

// RegisterTask encapsulates rendering and processing of the registration form.
type RegisterTask struct {
	tasks.TaskString
}

// registerTask handles user registration.
var registerTask = &RegisterTask{TaskString: TaskRegister}

// ensure RegisterTask satisfies tasks.Task
var _ tasks.Task = (*RegisterTask)(nil)

// RegisterPage renders the user registration form.
func (RegisterTask) Page(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData)
	handlers.TemplateHandler(w, r, "registerPage.gohtml", cd)
}

// RegisterActionPage handles user creation from the registration form.
func (RegisterTask) Action(w http.ResponseWriter, r *http.Request) {
	if config.AppRuntimeConfig.LogFlags&config.LogFlagAuth != 0 {
		log.Printf("registration attempt %s", r.PostFormValue("username"))
	}
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
	queries := r.Context().Value(consts.KeyQueries).(*db.Queries)

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

	if _, err := queries.UserByEmail(r.Context(), email); errors.Is(err, sql.ErrNoRows) {
	} else if err != nil {
		log.Printf("UserByUsername Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	} else {
		http.Error(w, "User already exists", http.StatusForbidden)
		return
	}

	hash, alg, err := HashPassword(password)
	if err != nil {
		log.Printf("hashPassword Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	result, err := queries.DB().ExecContext(r.Context(),
		"INSERT INTO users (username) VALUES (?)",
		username,
	)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "UNIQUE constraint failed") {
			http.Error(w, "User already exists", http.StatusForbidden)
			return
		}
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
	if err := queries.InsertUserEmail(r.Context(), db.InsertUserEmailParams{UserID: int32(lastInsertID), Email: email, VerifiedAt: sql.NullTime{}, LastVerificationCode: sql.NullString{}}); err != nil {
		log.Printf("InsertUserEmail Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := queries.InsertPassword(r.Context(), db.InsertPasswordParams{UsersIdusers: int32(lastInsertID), Passwd: hash, PasswdAlgorithm: sql.NullString{String: alg, Valid: true}}); err != nil {
		log.Printf("InsertPassword Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["signup"] = notifications.SignupInfo{Username: username}
		}
	}

	if config.AppRuntimeConfig.LogFlags&config.LogFlagAuth != 0 {
		log.Printf("registration success uid=%d", lastInsertID)
	}

	renderLoginForm(w, r, "approval is pending")

}

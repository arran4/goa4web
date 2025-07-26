package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"net/mail"
	"strings"

	"github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
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
func (RegisterTask) Action(w http.ResponseWriter, r *http.Request) any {
	if config.AppRuntimeConfig.LogFlags&config.LogFlagAuth != 0 {
		log.Printf("registration attempt %s", r.PostFormValue("username"))
	}
	if err := handlers.ValidateForm(r, []string{"username", "password", "email"}, []string{"username", "password", "email"}); err != nil {
		return fmt.Errorf("validation fail %w", err)
	}
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	email := r.PostFormValue("email")
	if _, err := mail.ParseAddress(email); err != nil {
		return handlers.ErrRedirectOnSamePageHandler(errors.New("invalid email"))
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	if _, err := queries.UserByUsername(r.Context(), sql.NullString{
		String: username,
		Valid:  true,
	}); errors.Is(err, sql.ErrNoRows) {
	} else if err != nil {
		log.Printf("UserByUsername Error: %s", err)
		return fmt.Errorf("user by username %w", err)
	} else {
		return handlers.ErrRedirectOnSamePageHandler(errors.New("user exists"))
	}

	if _, err := queries.UserByEmail(r.Context(), email); errors.Is(err, sql.ErrNoRows) {
	} else if err != nil {
		log.Printf("UserByUsername Error: %s", err)
		return fmt.Errorf("user by email %w", err)
	} else {
		return handlers.ErrRedirectOnSamePageHandler(errors.New("user exists"))
	}

	hash, alg, err := HashPassword(password)
	if err != nil {
		log.Printf("hashPassword Error: %s", err)
		return fmt.Errorf("hash password %w", err)
	}
	result, err := queries.DB().ExecContext(r.Context(),
		"INSERT INTO users (username) VALUES (?)",
		username,
	)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return handlers.ErrRedirectOnSamePageHandler(err)
		}
		log.Printf("InsertUser Error: %s", err)
		return fmt.Errorf("insert user %w", err)
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		log.Printf("LastInsertId Error: %s", err)
		return fmt.Errorf("last insert id %w", err)
	}
	if err := queries.InsertUserEmail(r.Context(), db.InsertUserEmailParams{UserID: int32(lastInsertID), Email: email, VerifiedAt: sql.NullTime{}, LastVerificationCode: sql.NullString{}}); err != nil {
		log.Printf("InsertUserEmail Error: %s", err)
		return fmt.Errorf("insert user email %w", err)
	}
	if err := queries.InsertPassword(r.Context(), db.InsertPasswordParams{UsersIdusers: int32(lastInsertID), Passwd: hash, PasswdAlgorithm: sql.NullString{String: alg, Valid: true}}); err != nil {
		log.Printf("InsertPassword Error: %s", err)
		return fmt.Errorf("insert password %w", err)
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["Username"] = username
		}
	}

	if config.AppRuntimeConfig.LogFlags&config.LogFlagAuth != 0 {
		log.Printf("registration success uid=%d", lastInsertID)
	}

	return loginFormHandler{msg: "approval is pending"}
}

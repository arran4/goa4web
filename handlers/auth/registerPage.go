package auth

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"net/url"
	"strings"

	"github.com/arran4/goa4web/core/consts"

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
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Register"
	type Data struct {
		Back string
	}
	back, _ := cd.SanitizeBackURL(r, r.FormValue("back"))
	data := Data{
		Back: back,
	}
	_ = RegisterPageTmpl.Handle(w, r, data)
}

const RegisterPageTmpl tasks.Template = "pages/auth/registerPage.gohtml"

// RegisterActionPage handles user creation from the registration form.
func (RegisterTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd.Config.LogFlags&config.LogFlagAuth != 0 {
		log.Printf("registration attempt %s", r.PostFormValue("username"))
	}
	if err := handlers.ValidateForm(r, []string{"username", "password", "email", "back"}, []string{"username", "password", "email"}); err != nil {
		return fmt.Errorf("validation fail %w", err)
	}
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	email := r.PostFormValue("email")
	if _, err := mail.ParseAddress(email); err != nil {
		return handlers.ErrRedirectOnSamePageHandler(errors.New("invalid email"))
	}
	if userExists, err := cd.UserExists(username, email); err != nil {
		log.Printf("UserExists Error: %s", err)
		return fmt.Errorf("user exists %w", err)
	} else if userExists {
		return handlers.ErrRedirectOnSamePageHandler(errors.New("user exists"))
	}

	hash, alg, err := HashPassword(password)
	if err != nil {
		log.Printf("hashPassword Error: %s", err)
		return fmt.Errorf("hash password %w", err)
	}
	id, err := cd.CreateUserWithEmail(username, email, hash, alg)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return handlers.ErrRedirectOnSamePageHandler(err)
		}
		log.Printf("CreateUserWithEmail Error: %s", err)
		return fmt.Errorf("create user with email %w", err)
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data["Username"] = username
		}
	}

	if cd.Config.LogFlags&config.LogFlagAuth != 0 {
		log.Printf("registration success uid=%d", id)
	}

	query := url.Values{"notice": {"approval is pending"}}
	if back, _ := cd.SanitizeBackURL(r, r.PostFormValue("back")); back != "" {
		query.Set("back", back)
	}
	target := "/login?" + query.Encode()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.DisableCaching(w)
		http.Redirect(w, r, target, http.StatusSeeOther)
	})
}

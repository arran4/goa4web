package handlers

import (
	"net/http"

	"github.com/arran4/goa4web/core/common"
)

// ValidateForm ensures that only the allowed form keys are present and that all required keys exist with a non-empty value.
// It parses the request form if needed. Allowed keys should include required keys.
func ValidateForm(r *http.Request, allowed, required []string) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	allowedSet := make(map[string]struct{}, len(allowed)+2)
	// Always allow CSRF and task fields which are automatically added by forms.
	allowedSet[CSRFField] = struct{}{}
	allowedSet[TaskField] = struct{}{}
	for _, k := range allowed {
		allowedSet[k] = struct{}{}
	}
	for k := range r.PostForm {
		if _, ok := allowedSet[k]; !ok {
			return common.UserError{ErrorMessage: "invalid form"}
		}
	}
	for _, k := range required {
		if v := r.PostFormValue(k); v == "" {
			return common.UserError{ErrorMessage: "missing " + k}
		}
	}
	return nil
}

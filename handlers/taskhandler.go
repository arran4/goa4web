package handlers

import (
	"errors"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/tasks"
	"log"
	"net/http"
	"net/url"
)

// TaskHandler wraps t.Action to record the task on the request event and handle the
// returned result
func TaskHandler(t tasks.Task) func(http.ResponseWriter, *http.Request) {
	if nt, ok := t.(tasks.NamedTask); ok {
		tasks.Register(nt)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if v := r.Context().Value(consts.KeyCoreData).(*common.CoreData); v != nil {
			v.SetEventTask(t)
		}
		result := t.Action(w, r)
		switch result := result.(type) {
		case http.HandlerFunc:
			result(w, r)
		case http.Handler:
			result.ServeHTTP(w, r)
		case error:
			var ue interface {
				error
				UserErrorMessage() string
			}
			if errors.As(result, &ue) {
				if msg := ue.UserErrorMessage(); msg != "" {
					r.URL.RawQuery = "error=" + url.QueryEscape(msg)
				} else {
					r.URL.RawQuery = "error=" + url.QueryEscape(result.Error())
				}
				TaskErrorAcknowledgementPage(w, r)
				return
			}
			log.Printf("task action: %v", result)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		default:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
}

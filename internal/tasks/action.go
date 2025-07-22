package tasks

import (
	"errors"
	"log"
	"net/http"
	"net/url"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

// eventTaskSetter defines the minimal interface needed to record the task on the event.
// Implemented by coredata.CoreData.
type eventTaskSetter interface {
	SetEventTask(Task)
}

// Action wraps t.Action to record the task on the request event.
func Action(t Task) func(http.ResponseWriter, *http.Request) {
	if nt, ok := t.(NamedTask); ok {
		Register(nt)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if v := r.Context().Value(consts.KeyCoreData); v != nil {
			if s, ok := v.(eventTaskSetter); ok {
				s.SetEventTask(t)
			}
		}
		t.Action(w, r)
	}
}

// ActionV2 wraps t.Action to record the task on the request event and handle the
// returned ActionResult and error.
func ActionV2(t ActionResultV2) func(http.ResponseWriter, *http.Request) {
	if nt, ok := t.(NamedTask); ok {
		Register(nt)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if v := r.Context().Value(consts.KeyCoreData); v != nil {
			if s, ok := v.(eventTaskSetter); ok {
				s.SetEventTask(t)
			}
		}
		result, err := t.Action(w, r)
		if err != nil {
			var ue *common.UserError
			if errors.As(err, &ue) {
				if msg := ue.ErrorMessage; msg != "" {
					r.URL.RawQuery = "error=" + url.QueryEscape(msg)
				} else {
					r.URL.RawQuery = "error=" + url.QueryEscape(err.Error())
				}
				handlers.TaskErrorAcknowledgementPage(w, r)
				return
			}
			log.Printf("task action: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if result != nil {
			result.Action(w, r)
		}
	}
}

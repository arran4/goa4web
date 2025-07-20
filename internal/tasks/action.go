package tasks

import (
	"net/http"

	"github.com/arran4/goa4web/core/consts"
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

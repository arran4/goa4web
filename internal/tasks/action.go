package tasks

import (
	"log"
	"net/http"

	"github.com/arran4/goa4web/core/consts"
)

// eventTaskSetter defines the minimal interface needed to record the task on the event.
// Implemented by coredata.CoreData.
type eventTaskSetter interface {
	SetEventTask(Task)
}

// Action wraps t.Action to record the task on the request event.
// TODO refactor out preferring handlers.ActionHandler
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
		if result := t.Action(w, r); result != nil {
			log.Panicf("Action returned %v (Migrate to ActionHandler)", result)
		}
	}
}

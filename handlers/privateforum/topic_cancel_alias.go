package privateforum

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
)

// TopicCancelAlias provides a backwards-compatible alias for
// GET /private/topic/{topic}/cancel by redirecting to the existing
// thread cancel confirmation page at /private/topic/{topic}/thread/cancel.
func TopicCancelAlias(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vars := mux.Vars(r)
	topic := vars["topic"]
	base := cd.ForumBasePath
	if base == "" {
		base = "/private"
	}
	target := fmt.Sprintf("%s/topic/%s/thread/cancel", base, topic)
	http.Redirect(w, r, target, http.StatusSeeOther)
}

package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/tasks"
)

// TaskHandler wraps t.Action to record the task on the request event and handle the
// returned result
func TaskHandler(t tasks.Task) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if v := r.Context().Value(consts.KeyCoreData).(*common.CoreData); v != nil {
			v.SetEventTask(t)
		}
		result := t.Action(w, r)

		if v := r.Context().Value(consts.KeyCoreData).(*common.CoreData); v != nil {
			if pt, ok := result.(tasks.HasPageTitle); ok {
				v.PageTitle = pt.PageTitle()
			}
			if hb, ok := result.(tasks.HasBreadcrumb); ok {
				v.SetCurrentPage(hb)
			}
		}

		switch result := result.(type) {
		case RedirectHandler:
			// Use 303 See Other so POST actions redirect to a GET of the target resource.
			// 307 would preserve the HTTP method and often breaks when the target only supports GET.
			http.Redirect(w, r, string(result), http.StatusSeeOther)
		case RefreshDirectHandler:
			cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
			cd.AutoRefresh = result.Content()
			TaskDoneAutoRefreshPageTmpl.Handle(w, r, result)
		case TextByteWriter:
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			if _, err := w.Write([]byte(result)); err != nil {
				log.Printf("write response: %v", err)
			}
		case http.HandlerFunc:
			result(w, r)
		case http.Handler:
			result.ServeHTTP(w, r)
		case SessionFetchFail:
			loginRedirect(w, r)
		case *SessionFetchFail:
			loginRedirect(w, r)
		case nil:
			TaskDoneAutoRefreshPage(w, r)
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
			RenderErrorPage(w, r, result)
			return
		default:
			RenderErrorPage(w, r, fmt.Errorf("%s", http.StatusText(http.StatusInternalServerError)))
		}
	}
}

func loginRedirect(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	vals := url.Values{}
	vals.Set("back", r.URL.RequestURI())
	if r.Method != http.MethodGet {
		if err := r.ParseForm(); err == nil {
			vals.Set("method", r.Method)
			if enc, err := cd.EncryptData(r.Form.Encode()); err == nil {
				vals.Set("data", enc)
			}
		}
	}
	http.Redirect(w, r, "/login?"+vals.Encode(), http.StatusTemporaryRedirect)
}

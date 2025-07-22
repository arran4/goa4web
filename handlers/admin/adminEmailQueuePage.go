package admin

import (
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"net/mail"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/tasks"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/workers/emailqueue"
)

// ResendQueueTask triggers sending queued emails immediately.
type ResendQueueTask struct{ tasks.TaskString }

var resendQueueTask = &ResendQueueTask{TaskString: TaskResend}

// ensure ResendQueueTask satisfies the tasks.Task interface
var _ tasks.Task = (*ResendQueueTask)(nil)

// DeleteQueueTask removes queued emails without sending.
type DeleteQueueTask struct{ tasks.TaskString }

var deleteQueueTask = &DeleteQueueTask{TaskString: TaskDelete}

// ensure DeleteQueueTask satisfies the tasks.Task interface
var _ tasks.Task = (*DeleteQueueTask)(nil)

func AdminEmailQueuePage(w http.ResponseWriter, r *http.Request) {
	type EmailItem struct {
		*db.ListUnsentPendingEmailsRow
		Email   string
		Subject string
	}
	type Data struct {
		*common.CoreData
		Emails []EmailItem
	}
	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
	}
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	rows, err := queries.ListUnsentPendingEmails(r.Context())
	if err != nil {
		log.Printf("list pending emails: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	ids := make([]int32, 0, len(rows))
	for _, e := range rows {
		if e.ToUserID.Valid {
			ids = append(ids, e.ToUserID.Int32)
		}
	}
	users := make(map[int32]*db.GetUserByIdRow)
	for _, id := range ids {
		if u, err := queries.GetUserById(r.Context(), id); err == nil {
			users[id] = u
		}
	}
	for _, e := range rows {
		emailStr := ""
		if e.ToUserID.Valid {
			if u, ok := users[e.ToUserID.Int32]; ok && u.Email.Valid && u.Email.String != "" {
				emailStr = u.Email.String
			}
		}
		subj := ""
		if m, err := mail.ReadMessage(strings.NewReader(e.Body)); err == nil {
			subj = m.Header.Get("Subject")
		}
		data.Emails = append(data.Emails, EmailItem{e, emailStr, subj})
	}
	handlers.TemplateHandler(w, r, "emailQueuePage.gohtml", data)
}

func (ResendQueueTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	provider := email.ProviderFromConfig(config.AppRuntimeConfig)
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm: %v", err)
	}
	var emails []*db.GetPendingEmailByIDRow
	var ids []int32
	for _, idStr := range r.Form["id"] {
		id, _ := strconv.Atoi(idStr)
		e, err := queries.GetPendingEmailByID(r.Context(), int32(id))
		if err != nil {
			log.Printf("get email: %v", err)
			continue
		}
		emails = append(emails, e)
		if e.ToUserID.Valid {
			ids = append(ids, e.ToUserID.Int32)
		}
	}
	users := make(map[int32]*db.GetUserByIdRow)
	for _, id := range ids {
		if u, err := queries.GetUserById(r.Context(), id); err == nil {
			users[id] = u
		}
	}
	for _, e := range emails {
		addr, err := emailqueue.ResolveQueuedEmailAddress(r.Context(), queries, &db.FetchPendingEmailsRow{ID: e.ID, ToUserID: e.ToUserID, Body: e.Body, ErrorCount: e.ErrorCount, DirectEmail: e.DirectEmail})
		if err != nil {
			log.Printf("resolve address: %v", err)
			continue
		}
		if provider != nil {
			if err := provider.Send(r.Context(), addr, []byte(e.Body)); err != nil {
				log.Printf("send email: %v", err)
				continue
			}
		}
		if err := queries.MarkEmailSent(r.Context(), e.ID); err != nil {
			log.Printf("mark sent: %v", err)
		}
	}
	handlers.TaskDoneAutoRefreshPage(w, r)
}

func (DeleteQueueTask) Action(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm: %v", err)
	}
	for _, idStr := range r.Form["id"] {
		id, _ := strconv.Atoi(idStr)
		if err := queries.DeletePendingEmail(r.Context(), int32(id)); err != nil {
			log.Printf("delete email: %v", err)
		}
	}
	handlers.TaskDoneAutoRefreshPage(w, r)
}

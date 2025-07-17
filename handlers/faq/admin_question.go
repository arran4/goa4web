package faq

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"

	"github.com/arran4/goa4web/core"
)

func AdminQuestionsPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*corecorecommon.CoreData
		Categories []*db.FaqCategory
		Rows       []*db.Faq
	}

	data := Data{
		CoreData: r.Context().Value(corecommon.KeyCoreData).(*corecorecommon.CoreData),
	}

	queries := r.Context().Value(corecommon.KeyQueries).(*db.Queries)

	catrows, err := queries.GetAllFAQCategories(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	data.Categories = catrows

	rows, err := queries.GetAllFAQQuestions(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
	data.Rows = rows

	common.TemplateHandler(w, r, "adminQuestionPage.gohtml", data)
}

func QuestionsDeleteActionPage(w http.ResponseWriter, r *http.Request) {
	faq, err := strconv.Atoi(r.PostFormValue("faq"))
	if err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(corecommon.KeyQueries).(*db.Queries)

	if err := queries.DeleteFAQ(r.Context(), int32(faq)); err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	common.TaskDoneAutoRefreshPage(w, r)
}

func QuestionsEditActionPage(w http.ResponseWriter, r *http.Request) {
	question := r.PostFormValue("question")
	answer := r.PostFormValue("answer")
	category, err := strconv.Atoi(r.PostFormValue("category"))
	if err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	faq, err := strconv.Atoi(r.PostFormValue("faq"))
	if err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(corecommon.KeyQueries).(*db.Queries)

	if err := queries.UpdateFAQQuestionAnswer(r.Context(), db.UpdateFAQQuestionAnswerParams{
		Answer:                       sql.NullString{Valid: true, String: answer},
		Question:                     sql.NullString{Valid: true, String: question},
		FaqcategoriesIdfaqcategories: int32(category),
		Idfaq:                        int32(faq),
	}); err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	common.TaskDoneAutoRefreshPage(w, r)
}

func QuestionsCreateActionPage(w http.ResponseWriter, r *http.Request) {
	question := r.PostFormValue("question")
	answer := r.PostFormValue("answer")
	category, err := strconv.Atoi(r.PostFormValue("category"))
	if err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}
	queries := r.Context().Value(corecommon.KeyQueries).(*db.Queries)
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	if _, err := queries.DB().ExecContext(r.Context(),
		"INSERT INTO faq (question, answer, faqCategories_idfaqCategories, users_idusers, language_idlanguage) VALUES (?, ?, ?, ?, ?)",
		sql.NullString{String: question, Valid: true},
		sql.NullString{String: answer, Valid: true},
		int32(category), uid, 1,
	); err != nil {
		log.Printf("Error: %s", err)
		http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	common.TaskDoneAutoRefreshPage(w, r)
}

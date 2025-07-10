package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"log"
	"net/http"

	corecommon "github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
	common "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/utils/emailutil"
)

func ForgotPasswordPage(w http.ResponseWriter, r *http.Request) {
	data := struct{ *corecommon.CoreData }{CoreData: r.Context().Value(common.KeyCoreData).(*corecommon.CoreData)}
	if err := templates.RenderTemplate(w, "forgotPasswordPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func ForgotPasswordActionPage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	username := r.PostFormValue("username")
	pw := r.PostFormValue("password")
	if username == "" || pw == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}
	queries := r.Context().Value(common.KeyQueries).(*db.Queries)
	row, err := queries.GetUserByUsername(r.Context(), sql.NullString{String: username, Valid: true})
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if row.Email == "" {
		http.Error(w, "no verified email", http.StatusBadRequest)
		return
	}
	hash, alg, err := HashPassword(pw)
	if err != nil {
		http.Error(w, "hash error", http.StatusInternalServerError)
		return
	}
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		http.Error(w, "rand", http.StatusInternalServerError)
		return
	}
	code := hex.EncodeToString(buf[:])
	if err := queries.CreatePasswordReset(r.Context(), db.CreatePasswordResetParams{UserID: row.Idusers, Passwd: hash, PasswdAlgorithm: alg, VerificationCode: code}); err != nil {
		log.Printf("create reset: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if row.Email != "" {
		page := r.URL.Scheme + "://" + r.Host + "/login"
		_ = emailutil.CreateEmailTemplateAndQueue(r.Context(), queries, row.Idusers, row.Email, page, common.TaskUserResetPassword, code)
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"log"
	"net/http"

	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	notif "github.com/arran4/goa4web/internal/notifications"
)

func ForgotPasswordPage(w http.ResponseWriter, r *http.Request) {
	hcommon.TemplateHandler(w, r, "forgotPasswordPage.gohtml", r.Context().Value(hcommon.KeyCoreData))
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
	queries := r.Context().Value(hcommon.KeyQueries).(*db.Queries)
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
		if cd, ok := r.Context().Value(common.KeyCoreData).(*corecommon.CoreData); ok {
			if evt := cd.Event(); evt != nil {
				if evt.Data == nil {
					evt.Data = map[string]any{}
				}
				evt.Data["reset"] = notif.PasswordResetInfo{Username: row.Username.String, Code: code}
			}
		}
		// OLD _ = emailutil.CreateEmailTemplateAndQueue(r.Context(), queries, row.Idusers, row.Email, page, hcommon.TaskUserResetPassword, code)
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

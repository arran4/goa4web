package user

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/handlers/auth"
	"github.com/arran4/goa4web/internal/db"
)

func cloneValues(v url.Values) url.Values {
	c := make(url.Values, len(v))
	for k, vals := range v {
		c[k] = append([]string(nil), vals...)
	}
	return c
}

func adminUsersPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Rows     []*db.UserFilteredRow
		Search   string
		Role     string
		Status   string
		NextLink string
		PrevLink string
		PageSize int
		Roles    []*db.Role
		Comments map[int32]*db.AdminUserComment
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	data := Data{
		CoreData: cd,
		Search:   r.URL.Query().Get("search"),
		Role:     r.URL.Query().Get("role"),
		Status:   r.URL.Query().Get("status"),
		PageSize: cd.PageSize(),
		Comments: map[int32]*db.AdminUserComment{},
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	if roles, err := data.AllRoles(); err == nil {
		data.Roles = roles
	}

	pageSize := data.PageSize
	var rows []*db.UserFilteredRow
	var err error
	if data.Search != "" {
		rows, err = queries.SearchUsersFiltered(r.Context(), db.SearchUsersFilteredParams{
			Query:  data.Search,
			Role:   data.Role,
			Status: data.Status,
			Limit:  int32(pageSize + 1),
			Offset: int32(offset),
		})
	} else {
		rows, err = queries.ListUsersFiltered(r.Context(), db.ListUsersFilteredParams{
			Role:   data.Role,
			Status: data.Status,
			Limit:  int32(pageSize + 1),
			Offset: int32(offset),
		})
	}
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	hasMore := len(rows) > pageSize
	if hasMore {
		rows = rows[:pageSize]
	}
	data.Rows = rows
	for _, u := range rows {
		if c, err := queries.LatestAdminUserComment(r.Context(), u.Idusers); err == nil {
			data.Comments[u.Idusers] = c
		}
	}

	params := url.Values{}
	if data.Search != "" {
		params.Set("search", data.Search)
	}
	if data.Role != "" {
		params.Set("role", data.Role)
	}
	if data.Status != "" {
		params.Set("status", data.Status)
	}
	if hasMore {
		nextVals := cloneValues(params)
		nextVals.Set("offset", strconv.Itoa(offset+pageSize))
		data.NextLink = "/admin/users?" + nextVals.Encode()
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: fmt.Sprintf("Next %d", pageSize),
			Link: data.NextLink,
		})
	}
	if offset > 0 {
		prevVals := cloneValues(params)
		prevVals.Set("offset", strconv.Itoa(offset-pageSize))
		data.PrevLink = "/admin/users?" + prevVals.Encode()
		data.CustomIndexItems = append(data.CustomIndexItems, common.IndexItem{
			Name: fmt.Sprintf("Previous %d", pageSize),
			Link: data.PrevLink,
		})
	}

	handlers.TemplateHandler(w, r, "usersPage.gohtml", data)
}

func adminUserDisableConfirmPage(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	u, err := cd.Queries().GetUserById(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	data := struct {
		*common.CoreData
		Message      string
		ConfirmLabel string
		Back         string
	}{
		CoreData:     cd,
		Message:      fmt.Sprintf("Are you sure you want to disable user %s (ID %d)?", u.Username.String, u.Idusers),
		ConfirmLabel: "Confirm disable",
		Back:         "/admin/users",
	}
	handlers.TemplateHandler(w, r, "confirmPage.gohtml", data)
}

func adminUserDisablePage(w http.ResponseWriter, r *http.Request) {
	uid := mux.Vars(r)["id"]
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Back:     "/admin/users",
	}
	if uidi, err := strconv.Atoi(uid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else if _, err := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries().DB().ExecContext(r.Context(), "DELETE FROM users WHERE idusers = ?", uidi); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("delete user: %w", err).Error())
	}
	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}

func adminUserEditFormPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	uid, _ := strconv.Atoi(r.URL.Query().Get("uid"))
	urow, err := queries.GetUserById(r.Context(), int32(uid))
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	user := &db.User{Idusers: urow.Idusers, Username: urow.Username}
	data := struct {
		*common.CoreData
		User *db.User
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		User:     user,
	}
	handlers.TemplateHandler(w, r, "userEditPage.gohtml", data)
}

func adminUserEditSavePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	uid := r.PostFormValue("uid")
	username := r.PostFormValue("username")
	email := r.PostFormValue("email")
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Back:     "/admin/users",
	}
	if uidi, err := strconv.Atoi(uid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else if _, err := queries.DB().ExecContext(r.Context(), "UPDATE users SET username=?, email=? WHERE idusers=?", username, email, uidi); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("update user: %w", err).Error())
	}
	handlers.TemplateHandler(w, r, "runTaskPage.gohtml", data)
}

func adminUserResetPasswordPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()
	uid := r.PostFormValue("uid")
	data := struct {
		*common.CoreData
		Errors   []string
		Messages []string
		Back     string
		Password string
	}{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Back:     "/admin/users",
	}
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("rand.Read: %w", err).Error())
	}
	newPass := hex.EncodeToString(buf[:])
	if hash, alg, err := auth.HashPassword(newPass); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("hashPassword: %w", err).Error())
	} else if uidi, err := strconv.Atoi(uid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else if err := queries.InsertPassword(r.Context(), db.InsertPasswordParams{UsersIdusers: int32(uidi), Passwd: hash, PasswdAlgorithm: sql.NullString{String: alg, Valid: true}}); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("reset password: %w", err).Error())
	} else {
		data.Password = newPass
	}
	handlers.TemplateHandler(w, r, "userResetPasswordPage.gohtml", data)
}

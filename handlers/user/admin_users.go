package user

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/arran4/goa4web/core/consts"

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
		Rows     []*db.UserFilteredRow
		Search   string
		Role     string
		Status   string
		PageSize int
		Roles    []*db.Role
		Comments map[int32]*db.AdminUserComment
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	data := Data{
		Search:   r.URL.Query().Get("search"),
		Role:     r.URL.Query().Get("role"),
		Status:   r.URL.Query().Get("status"),
		PageSize: cd.PageSize(),
		Comments: map[int32]*db.AdminUserComment{},
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	queries := cd.Queries()
	cqueries, ok := queries.(interface {
		AdminSearchUsersFiltered(context.Context, db.AdminSearchUsersFilteredParams) ([]*db.UserFilteredRow, error)
		AdminListUsersFiltered(context.Context, db.AdminListUsersFilteredParams) ([]*db.UserFilteredRow, error)
	})
	if !ok {
		log.Printf("adminUsersPage: database not available")
		handlers.RenderErrorPage(w, r, fmt.Errorf("database not available"))
		return
	}
	if roles, err := cd.AllRoles(); err == nil {
		data.Roles = roles
	}

	pageSize := data.PageSize
	var rows []*db.UserFilteredRow
	var err error
	if data.Search != "" {
		rows, err = cqueries.AdminSearchUsersFiltered(r.Context(), db.AdminSearchUsersFilteredParams{
			Query:  data.Search,
			Role:   data.Role,
			Status: data.Status,
			Limit:  int32(pageSize + 1),
			Offset: int32(offset),
		})
	} else {
		rows, err = cqueries.AdminListUsersFiltered(r.Context(), db.AdminListUsersFilteredParams{
			Role:   data.Role,
			Status: data.Status,
			Limit:  int32(pageSize + 1),
			Offset: int32(offset),
		})
	}
	if err != nil {
		log.Printf("list users: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
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
		cd.NextLink = "/admin/users?" + nextVals.Encode()
	}
	if offset > 0 {
		prevVals := cloneValues(params)
		prevVals.Set("offset", strconv.Itoa(offset-pageSize))
		cd.PrevLink = "/admin/users?" + prevVals.Encode()
	}

	handlers.TemplateHandler(w, r, "admin/usersPage.gohtml", data)
}

func adminUserDisableConfirmPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	u := cd.CurrentProfileUser()
	if u == nil {
		log.Printf("adminUserDisableConfirmPage: user not found")
		handlers.RenderErrorPage(w, r, fmt.Errorf("user not found"))
		return
	}
	back := fmt.Sprintf("/admin/user/%d", u.Idusers)
	data := struct {
		Message      string
		ConfirmLabel string
		Back         string
	}{
		Message:      fmt.Sprintf("Are you sure you want to disable user %s (ID %d)?", u.Username.String, u.Idusers),
		ConfirmLabel: "Confirm disable",
		Back:         back,
	}
	handlers.TemplateHandler(w, r, "admin/confirmPage.gohtml", data)
}

func adminUserDisablePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	id := cd.CurrentProfileUser()
	back := "/admin/users"
	if id != nil {
		back = fmt.Sprintf("/admin/user/%d", id.Idusers)
	}
	data := struct {
		Errors   []string
		Messages []string
		Back     string
	}{
		Back: back,
	}
	if id == nil {
		data.Errors = append(data.Errors, "invalid user id")
	} else if err := cd.Queries().AdminDeleteUserByID(r.Context(), id.Idusers); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("delete user: %w", err).Error())
	}
	handlers.TemplateHandler(w, r, "admin/runTaskPage.gohtml", data)
}

func adminUserEditFormPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	urow := cd.CurrentProfileUser()
	if urow == nil {
		log.Printf("adminUserEditFormPage: user not found")
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
		return
	}
	data := struct {
		User *db.SystemGetUserByIDRow
	}{
		User: urow,
	}
	handlers.TemplateHandler(w, r, "admin/userEditPage.gohtml", data)
}

func adminUserEditSavePage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	queries := cd.Queries()
	uidStr := r.PostFormValue("uid")
	uid := cd.CurrentProfileUser()
	username := r.PostFormValue("username")
	email := r.PostFormValue("email")
	back := "/admin/users"
	if uid != nil {
		back = fmt.Sprintf("/admin/user/%d", uid.Idusers)
	}
	data := struct {
		Errors   []string
		Messages []string
		Back     string
	}{
		Back: back,
	}
	var targetID int32
	if uidStr != "" {
		if uidi, err := strconv.Atoi(uidStr); err != nil {
			data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
		} else {
			targetID = int32(uidi)
		}
	} else if uid != nil {
		targetID = uid.Idusers
	}
	if targetID != 0 {
		if err := queries.AdminUpdateUsernameByID(r.Context(), db.AdminUpdateUsernameByIDParams{Username: sql.NullString{String: username, Valid: username != ""}, Idusers: targetID}); err != nil {
			data.Errors = append(data.Errors, fmt.Errorf("update user: %w", err).Error())
		} else if err := queries.AdminUpdateUserEmail(r.Context(), db.AdminUpdateUserEmailParams{Email: email, UserID: targetID}); err != nil {
			data.Errors = append(data.Errors, fmt.Errorf("update user email: %w", err).Error())
		}
	}
	handlers.TemplateHandler(w, r, "admin/runTaskPage.gohtml", data)
}

func adminUserResetPasswordPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	uid := r.PostFormValue("uid")
	u := cd.CurrentProfileUser()
	back := "/admin/users"
	if u != nil {
		back = fmt.Sprintf("/admin/user/%d", u.Idusers)
	}
	data := struct {
		Errors   []string
		Messages []string
		Back     string
		Password string
	}{
		Back: back,
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
	handlers.TemplateHandler(w, r, "admin/userResetPasswordPage.gohtml", data)
}

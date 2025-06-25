package goa4web

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers/common"
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
		*CoreData
		Rows     []*User
		Search   string
		Role     string
		Status   string
		NextLink string
		PrevLink string
		PageSize int
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Search:   r.URL.Query().Get("search"),
		Role:     r.URL.Query().Get("role"),
		Status:   r.URL.Query().Get("status"),
		PageSize: common.GetPageSize(r),
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	pageSize := data.PageSize
	var rows []*User
	var err error
	if data.Search != "" {
		rows, err = queries.SearchUsersFiltered(r.Context(), SearchUsersFilteredParams{
			Query:  data.Search,
			Role:   data.Role,
			Status: data.Status,
			Limit:  int32(pageSize + 1),
			Offset: int32(offset),
		})
	} else {
		rows, err = queries.ListUsersFiltered(r.Context(), ListUsersFilteredParams{
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
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: fmt.Sprintf("Next %d", pageSize),
			Link: data.NextLink,
		})
	}
	if offset > 0 {
		prevVals := cloneValues(params)
		prevVals.Set("offset", strconv.Itoa(offset-pageSize))
		data.PrevLink = "/admin/users?" + prevVals.Encode()
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: fmt.Sprintf("Previous %d", pageSize),
			Link: data.PrevLink,
		})
	}

	err = templates.RenderTemplate(w, "usersPage.gohtml", data, common.NewFuncs(r))
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func adminUserDisablePage(w http.ResponseWriter, r *http.Request) {
	uid := r.PostFormValue("uid")
	data := struct {
		*CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Back:     "/admin/users",
	}
	if uidi, err := strconv.Atoi(uid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else if _, err := r.Context().Value(ContextValues("queries")).(*Queries).DB().ExecContext(r.Context(), "DELETE FROM users WHERE idusers = ?", uidi); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("delete user: %w", err).Error())
	}
	err := templates.RenderTemplate(w, "runTaskPage.gohtml", data, common.NewFuncs(r))
	if err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func adminUserEditFormPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	uid, _ := strconv.Atoi(r.URL.Query().Get("uid"))
	user, err := queries.GetUserById(r.Context(), int32(uid))
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	data := struct {
		*CoreData
		User *User
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		User:     user,
	}
	if err := templates.RenderTemplate(w, "userEditPage.gohtml", data, common.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func adminUserEditSavePage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	uid := r.PostFormValue("uid")
	username := r.PostFormValue("username")
	email := r.PostFormValue("email")
	data := struct {
		*CoreData
		Errors   []string
		Messages []string
		Back     string
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Back:     "/admin/users",
	}
	if uidi, err := strconv.Atoi(uid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else if _, err := queries.DB().ExecContext(r.Context(), "UPDATE users SET username=?, email=? WHERE idusers=?", username, email, uidi); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("update user: %w", err).Error())
	}
	if err := templates.RenderTemplate(w, "runTaskPage.gohtml", data, common.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func adminUserResetPasswordPage(w http.ResponseWriter, r *http.Request) {
	queries := r.Context().Value(ContextValues("queries")).(*Queries)
	uid := r.PostFormValue("uid")
	data := struct {
		*CoreData
		Errors   []string
		Messages []string
		Back     string
		Password string
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Back:     "/admin/users",
	}
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("rand.Read: %w", err).Error())
	}
	newPass := hex.EncodeToString(buf[:])
	if hash, alg, err := hashPassword(newPass); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("hashPassword: %w", err).Error())
	} else if uidi, err := strconv.Atoi(uid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else if _, err := queries.DB().ExecContext(r.Context(), "UPDATE users SET passwd=?, passwd_algorithm=? WHERE idusers=?", hash, alg, uidi); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("reset password: %w", err).Error())
	} else {
		data.Password = newPass
	}
	if err := templates.RenderTemplate(w, "userResetPasswordPage.gohtml", data, common.NewFuncs(r)); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

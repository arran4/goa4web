package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func adminUsersPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Rows     []*User
		Search   string
		NextLink string
		PrevLink string
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Search:   r.URL.Query().Get("search"),
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	const pageSize = 15
	var rows []*User
	var err error
	if data.Search != "" {
		rows, err = queries.SearchUsers(r.Context(), SearchUsersParams{
			Query:  data.Search,
			Limit:  pageSize + 1,
			Offset: int32(offset),
		})
	} else {
		rows, err = queries.ListUsers(r.Context(), ListUsersParams{
			Limit:  pageSize + 1,
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

	base := "/admin/users"
	if data.Search != "" {
		base += "?search=" + url.QueryEscape(data.Search)
	}
	if hasMore {
		if strings.Contains(base, "?") {
			data.NextLink = fmt.Sprintf("%s&offset=%d", base, offset+pageSize)
		} else {
			data.NextLink = fmt.Sprintf("%s?offset=%d", base, offset+pageSize)
		}
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Next 15",
			Link: data.NextLink,
		})
	}
	if offset > 0 {
		if strings.Contains(base, "?") {
			data.PrevLink = fmt.Sprintf("%s&offset=%d", base, offset-pageSize)
		} else {
			data.PrevLink = fmt.Sprintf("%s?offset=%d", base, offset-pageSize)
		}
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Previous 15",
			Link: data.PrevLink,
		})
	}

	err = renderTemplate(w, r, "adminUsersPage.gohtml", data)
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
		Errors []string
		Back   string
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Back:     "/admin/users",
	}
	if uidi, err := strconv.Atoi(uid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else if _, err := r.Context().Value(ContextValues("queries")).(*Queries).db.ExecContext(r.Context(), "DELETE FROM users WHERE idusers = ?", uidi); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("delete user: %w", err).Error())
	}
	err := renderTemplate(w, r, "adminRunTaskPage.gohtml", data)
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
	if err := renderTemplate(w, r, "adminUserEditPage.gohtml", data); err != nil {
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
		Errors []string
		Back   string
	}{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Back:     "/admin/users",
	}
	if uidi, err := strconv.Atoi(uid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else if _, err := queries.db.ExecContext(r.Context(), "UPDATE users SET username=?, email=? WHERE idusers=?", username, email, uidi); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("update user: %w", err).Error())
	}
	if err := renderTemplate(w, r, "adminRunTaskPage.gohtml", data); err != nil {
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
	if uidi, err := strconv.Atoi(uid); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("strconv.Atoi: %w", err).Error())
	} else if _, err := queries.db.ExecContext(r.Context(), "UPDATE users SET passwd=MD5(?) WHERE idusers=?", newPass, uidi); err != nil {
		data.Errors = append(data.Errors, fmt.Errorf("reset password: %w", err).Error())
	} else {
		data.Password = newPass
	}
	if err := renderTemplate(w, r, "adminUserResetPasswordPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

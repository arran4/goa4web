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
		Rows   []*User
		Search string
	}

	data := Data{
		CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		Search:   r.URL.Query().Get("search"),
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	rows, err := queries.AllUsers(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if data.Search != "" {
		q := strings.ToLower(data.Search)
		var filtered []*User
		for _, row := range rows {
			if strings.Contains(strings.ToLower(row.Username.String), q) ||
				strings.Contains(strings.ToLower(row.Email.String), q) {
				filtered = append(filtered, row)
			}
		}
		rows = filtered
	}

	const pageSize = 15
	if offset < 0 {
		offset = 0
	}
	if offset > len(rows) {
		offset = len(rows)
	}
	end := offset + pageSize
	if end > len(rows) {
		end = len(rows)
	}
	data.Rows = rows[offset:end]

	if data.Search != "" {
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Next 15",
			Link: fmt.Sprintf("/admin/users?search=%s&offset=%d", url.QueryEscape(data.Search), offset+pageSize),
		})
		if offset > 0 {
			data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
				Name: "Previous 15",
				Link: fmt.Sprintf("/admin/users?search=%s&offset=%d", url.QueryEscape(data.Search), offset-pageSize),
			})
		}
	} else {
		data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
			Name: "Next 15",
			Link: fmt.Sprintf("/admin/users?offset=%d", offset+pageSize),
		})
		if offset > 0 {
			data.CustomIndexItems = append(data.CustomIndexItems, IndexItem{
				Name: "Previous 15",
				Link: fmt.Sprintf("/admin/users?offset=%d", offset-pageSize),
			})
		}
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

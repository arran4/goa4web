package main

import (
	"database/sql"
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql" // Import the MySQL driver.
)

var (
	//go:embed "template.tmpl"
	adminUserPermissionsTemplate string
)

// Define your indexitem struct.
type indexitem struct {
	Name string // Name of URL displayed in <a href>
	Link string // URL for link.
}

// Your index items.
var indexItems = []indexitem{
	{Name: "News", Link: "index"},
	{Name: "FAQ", Link: "faq"},
	{Name: "Blogs", Link: "blogs"},
	{Name: "Forum", Link: "forum"},
	{Name: "Linker", Link: "linker"},
	{Name: "Bookmarks", Link: "bookmarks"},
	{Name: "ImageBBS", Link: "imagebbs"},
	{Name: "Search", Link: "search"},
	{Name: "Writings", Link: "writings"},
	{Name: "Information", Link: "information"},
	{Name: "Preferences", Link: "user"},
}

// AdminUserPermissionsData holds the data needed for rendering the template.
type AdminUserPermissionsData struct {
	Items  []indexitem
	Level  int
	UserID int // Replace this with the actual user ID from the session.
	Rows   []PermissionRow
}

// PermissionRow holds the data for each permission row.
type PermissionRow struct {
	ID     string
	User   string
	Email  string
	Level  string
	Where  string
	Delete template.HTML // Store the Delete button as raw HTML.
}

func adminUserPermissions(w http.ResponseWriter, r *http.Request) {
	// Get the session.
	session, err := store.Get(r, sessionName)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get level from the query parameters or the session.
	level, err := strconv.Atoi(r.FormValue("level"))
	if err != nil {
		level = -1
	}

	// Get UserID from the session (replace "your-user-id" with the actual key).
	userID, _ := session.Values["your-user-id"].(int)

	// Prepare the index items.
	data := AdminUserPermissionsData{
		Items:  indexItems,
		Level:  level,
		UserID: userID,
	}

	// Query the database (replace "your-db-connection-string" with the actual connection string).
	db, err := sql.Open("mysql", "a4web:a4web@tcp(localhost:3306)/a4web")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT p.idpermissions, p.level, u.username, u.email, p.section " +
		"FROM permissions p, users u " +
		"WHERE u.idusers=p.users_idusers " +
		"ORDER BY p.level")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var permID, permLevel, username, email, section string
		err := rows.Scan(&permID, &permLevel, &username, &email, &section)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		data.Rows = append(data.Rows, PermissionRow{
			ID:    permID,
			User:  username,
			Email: email,
			Level: permLevel,
			Where: section,
			Delete: template.HTML(fmt.Sprintf("<form method=\"post\">"+
				"<input type=\"hidden\" name=\"permid\" value=\"%s\">"+
				"<input type=\"submit\" name=\"task\" value=\"User Disallow\">"+
				"</form>", permID)),
		})
	}

	// Prepare the HTML template.
	tmpl := template.Must(template.New("adminUserPermissions").Parse(adminUserPermissionsTemplate))

	// Execute the template with the data and write the output to the response.
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

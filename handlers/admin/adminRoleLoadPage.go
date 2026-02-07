package admin

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	roletemplates "github.com/arran4/goa4web/internal/role_templates"
	"github.com/arran4/goa4web/internal/roles"
	"github.com/arran4/goa4web/internal/tasks"
)

const (
	// roleSQLUploadMaxBytes limits the size of uploaded role SQL files.
	roleSQLUploadMaxBytes = 2 * 1024 * 1024
)

type AdminRoleLoadPage struct {
	DBPool *sql.DB
}

func (p *AdminRoleLoadPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		RoleName       string
		ParsedRoleName string
		SQL            string
		Errors         []string
		Applied        bool
		ExistingRole   *db.Role
		GrantsAdded    []string
		GrantsRemoved  []string
		GrantGroups    []GrantGroup
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Load Role SQL"
	data := Data{}

	if r.Method == http.MethodPost {
		r.Body = http.MaxBytesReader(w, r.Body, roleSQLUploadMaxBytes)
		if err := r.ParseMultipartForm(roleSQLUploadMaxBytes); err != nil {
			data.Errors = append(data.Errors, fmt.Errorf("invalid upload: %w", err).Error())
			AdminRoleLoadPageTmpl.Handler(data).ServeHTTP(w, r)
			return
		}

		roleName := strings.TrimSpace(r.PostFormValue("role"))
		data.RoleName = roleName
		sqlText := strings.TrimSpace(r.PostFormValue("sql"))
		var sqlData []byte

		file, _, err := r.FormFile("role_sql")
		if err == nil {
			defer file.Close()
			if buf, readErr := io.ReadAll(file); readErr != nil {
				data.Errors = append(data.Errors, fmt.Errorf("read role file: %w", readErr).Error())
			} else {
				sqlData = buf
			}
		} else if err != http.ErrMissingFile {
			data.Errors = append(data.Errors, fmt.Errorf("role file: %w", err).Error())
		}

		if len(sqlData) == 0 && sqlText != "" {
			sqlData = []byte(sqlText)
		}
		data.SQL = string(sqlData)

		if roleName == "" {
			data.Errors = append(data.Errors, "role name is required")
		}
		if len(sqlData) == 0 {
			data.Errors = append(data.Errors, "role SQL is required")
		}

		var grants []*db.Grant
		if len(sqlData) > 0 {
			parsedName, err := roles.ParseRoleName(sqlData)
			if err != nil {
				data.Errors = append(data.Errors, fmt.Errorf("parse role name: %w", err).Error())
			} else {
				data.ParsedRoleName = parsedName
				if roleName != "" && roleName != parsedName {
					data.Errors = append(data.Errors, fmt.Sprintf("role name %q does not match SQL role %q", roleName, parsedName))
				}
			}

			parsedGrants, err := roles.ParseRoleGrants(sqlData)
			if err != nil {
				data.Errors = append(data.Errors, fmt.Errorf("parse grants: %w", err).Error())
			} else {
				grants = parsedGrants
			}
		}

		if len(grants) > 0 {
			groups, err := buildGrantGroupsFromGrants(r.Context(), cd, grants)
			if err != nil && err != sql.ErrNoRows {
				data.Errors = append(data.Errors, fmt.Errorf("build grant preview: %w", err).Error())
			} else {
				data.GrantGroups = groups
			}
		}

		if roleName != "" && len(grants) > 0 {
			if existing, err := cd.Queries().GetRoleByName(r.Context(), roleName); err == nil {
				data.ExistingRole = existing
				if currentGrants, err := cd.Queries().GetGrantsByRoleID(r.Context(), sql.NullInt32{Int32: existing.ID, Valid: true}); err == nil {
					added, removed := diffGrantKeys(currentGrants, grants)
					data.GrantsAdded = added
					data.GrantsRemoved = removed
				}
			}
		}

		if r.PostFormValue("apply") != "" && len(data.Errors) == 0 {
			if p.DBPool == nil {
				data.Errors = append(data.Errors, "database connection unavailable")
			} else if err := roles.ApplyRoleSQL(r.Context(), roleName, sqlData, p.DBPool); err != nil {
				data.Errors = append(data.Errors, err.Error())
			} else {
				data.Applied = true
			}
		}
	}

	AdminRoleLoadPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminRoleLoadPage) Breadcrumb() (string, string, common.HasBreadcrumb) {
	return "Role SQL Loader", "/admin/roles/load", &AdminRolesPage{}
}

func (p *AdminRoleLoadPage) PageTitle() string {
	return "Load Role SQL"
}

var _ common.Page = (*AdminRoleLoadPage)(nil)
var _ http.Handler = (*AdminRoleLoadPage)(nil)

func diffGrantKeys(current []*db.Grant, desired []*db.Grant) ([]string, []string) {
	currentMap := make(map[string]struct{}, len(current))
	for _, grant := range current {
		item := ""
		if grant.Item.Valid {
			item = grant.Item.String
		}
		itemID := int32(0)
		if grant.ItemID.Valid {
			itemID = grant.ItemID.Int32
		}
		currentMap[roletemplates.GrantKey(grant.Section, item, grant.Action, itemID)] = struct{}{}
	}

	desiredMap := make(map[string]struct{}, len(desired))
	for _, grant := range desired {
		item := ""
		if grant.Item.Valid {
			item = grant.Item.String
		}
		itemID := int32(0)
		if grant.ItemID.Valid {
			itemID = grant.ItemID.Int32
		}
		desiredMap[roletemplates.GrantKey(grant.Section, item, grant.Action, itemID)] = struct{}{}
	}

	added := make([]string, 0)
	removed := make([]string, 0)
	for key := range desiredMap {
		if _, ok := currentMap[key]; !ok {
			added = append(added, key)
		}
	}
	for key := range currentMap {
		if _, ok := desiredMap[key]; !ok {
			removed = append(removed, key)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	return added, removed
}

// AdminRoleLoadPageTmpl renders the role load admin page.
const AdminRoleLoadPageTmpl tasks.Template = "admin/roleLoadPage.gohtml"

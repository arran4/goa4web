package admin

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"sort"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	subscriptiontemplates "github.com/arran4/goa4web/internal/subscription_templates"
	"github.com/arran4/goa4web/internal/tasks"
)

const maxSubscriptionTemplateUploadBytes = 256 * 1024 // Maximum size for uploaded subscription template files.

type subscriptionTemplatePattern struct {
	Method  string
	Pattern string
}

type subscriptionTemplateInfo struct {
	Name     string
	Patterns []subscriptionTemplatePattern
}

type subscriptionTemplateUploadData struct {
	TemplateName string
	Filename     string
	Content      string
	Patterns     []subscriptionTemplatePattern
	Errors       map[string]string
}

func toSubscriptionTemplatePatterns(patterns []subscriptiontemplates.Pattern) []subscriptionTemplatePattern {
	formatted := make([]subscriptionTemplatePattern, 0, len(patterns))
	for _, entry := range patterns {
		formatted = append(formatted, subscriptionTemplatePattern{Method: entry.Method, Pattern: entry.Pattern})
	}
	return formatted
}

func validateSubscriptionTemplatePatterns(patterns []subscriptiontemplates.Pattern) error {
	if len(patterns) == 0 {
		return fmt.Errorf("has no patterns")
	}

	seen := make(map[string]struct{}, len(patterns))
	dupes := map[string]struct{}{}
	for _, entry := range patterns {
		key := entry.Method + "\x00" + entry.Pattern
		if _, ok := seen[key]; ok {
			dupes[entry.Method+" "+entry.Pattern] = struct{}{}
			continue
		}
		seen[key] = struct{}{}
	}
	if len(dupes) > 0 {
		entries := make([]string, 0, len(dupes))
		for entry := range dupes {
			entries = append(entries, entry)
		}
		sort.Strings(entries)
		return fmt.Errorf("duplicate patterns: %s", strings.Join(entries, ", "))
	}

	return nil
}

func readSubscriptionTemplateFile(file multipart.File) (string, error) {
	if file == nil {
		return "", fmt.Errorf("missing upload")
	}
	defer file.Close()

	limited := io.LimitReader(file, maxSubscriptionTemplateUploadBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return "", fmt.Errorf("read upload: %w", err)
	}
	if int64(len(data)) > maxSubscriptionTemplateUploadBytes {
		return "", fmt.Errorf("uploaded template exceeds %d bytes", maxSubscriptionTemplateUploadBytes)
	}
	return string(data), nil
}

// AdminSubscriptionTemplatesPage lists embedded subscription templates and their parsed patterns.
func AdminSubscriptionTemplatesPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Subscription Templates"
	if errMsg := r.URL.Query().Get("error"); errMsg != "" {
		cd.SetCurrentError(errMsg)
	}

	roles, err := cd.Queries().AdminListRoles(r.Context())
	if err != nil {
		handlers.RenderErrorPage(w, r, fmt.Errorf("list roles: %w", err))
		return
	}

	names, err := subscriptiontemplates.ListEmbeddedTemplates()
	if err != nil {
		handlers.RenderErrorPage(w, r, err)
		return
	}

	templates := make([]subscriptionTemplateInfo, 0, len(names))
	for _, name := range names {
		content, err := subscriptiontemplates.GetEmbeddedTemplate(name)
		if err != nil {
			handlers.RenderErrorPage(w, r, err)
			return
		}
		parsed := subscriptiontemplates.ParseTemplatePatterns(string(content))
		templates = append(templates, subscriptionTemplateInfo{Name: name, Patterns: toSubscriptionTemplatePatterns(parsed)})
	}

	upload := subscriptionTemplateUploadData{Errors: map[string]string{}}
	if r.Method == http.MethodPost {
		if err := r.ParseMultipartForm(maxSubscriptionTemplateUploadBytes); err != nil {
			upload.Errors["template_file"] = "Unable to parse the uploaded template file."
		} else {
			upload.TemplateName = strings.TrimSpace(r.PostFormValue("template_name"))
			file, header, err := r.FormFile("template_file")
			if errors.Is(err, http.ErrMissingFile) {
				upload.Errors["template_file"] = "Select a template file to upload."
			} else if err != nil {
				upload.Errors["template_file"] = "Unable to read the uploaded template file."
			} else {
				if header != nil {
					upload.Filename = header.Filename
					if upload.TemplateName == "" {
						upload.TemplateName = strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))
					}
				}
				content, err := readSubscriptionTemplateFile(file)
				if err != nil {
					upload.Errors["template_file"] = err.Error()
				} else {
					patterns := subscriptiontemplates.ParseTemplatePatterns(content)
					if err := validateSubscriptionTemplatePatterns(patterns); err != nil {
						upload.Errors["template_content"] = err.Error()
					} else {
						upload.Content = content
						upload.Patterns = toSubscriptionTemplatePatterns(patterns)
					}
				}
			}

			if upload.TemplateName == "" {
				upload.Errors["template_name"] = "Template name is required."
			}
		}
	}

	data := struct {
		*common.CoreData
		Templates []subscriptionTemplateInfo
		Roles     []*db.Role
		Upload    subscriptionTemplateUploadData
	}{
		CoreData:  cd,
		Templates: templates,
		Roles:     roles,
		Upload:    upload,
	}

	AdminSubscriptionTemplatesPageTmpl.Handle(w, r, data)
}

// AdminSubscriptionTemplatesPageTmpl renders the admin subscription templates page.
const AdminSubscriptionTemplatesPageTmpl tasks.Template = "admin/subscriptionTemplatesPage.gohtml"

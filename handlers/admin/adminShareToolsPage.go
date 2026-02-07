package admin

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/sign/signutil"
	"github.com/arran4/goa4web/internal/tasks"
)

type shareToolsResource struct {
	Value          string
	Label          string
	Hint           string
	RequiresThread bool
	BuildPath      func(primaryID int, threadID int) (string, error)
}

type shareToolsResourceOption struct {
	Value          string
	Label          string
	Hint           string
	RequiresThread bool
}

type shareToolsData struct {
	Resources        []shareToolsResourceOption
	SelectedResource string
	SelectedHint     string
	PrimaryID        string
	ThreadID         string
	Duration         string
	NoExpiry         bool
	SharedPath       string
	SignedURL        string
	Errors           []string
}

type AdminShareToolsPage struct{}

func (p *AdminShareToolsPage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Share Tools"

	resources, resourceMap := shareToolsResourceList()
	data := shareToolsData{
		Resources: resources,
		Duration:  "24h",
	}

	if r.Method != http.MethodPost {
		AdminShareToolsPageTmpl.Handler(data).ServeHTTP(w, r)
		return
	}

	if err := r.ParseForm(); err != nil {
		handlers.RenderErrorPage(w, r, handlers.ErrBadRequest)
		return
	}

	data.SelectedResource = strings.TrimSpace(r.PostFormValue("resource_type"))
	data.PrimaryID = strings.TrimSpace(r.PostFormValue("primary_id"))
	data.ThreadID = strings.TrimSpace(r.PostFormValue("thread_id"))
	data.Duration = strings.TrimSpace(r.PostFormValue("duration"))
	data.NoExpiry = r.PostFormValue("no_expiry") != ""
	if data.Duration == "" {
		data.Duration = "24h"
	}

	resource, ok := resourceMap[data.SelectedResource]
	if !ok {
		data.Errors = append(data.Errors, "Select a valid resource type.")
	} else {
		data.SelectedHint = resource.Hint
	}

	primaryID, err := parseShareToolsID(data.PrimaryID, "Primary ID")
	if err != nil {
		data.Errors = append(data.Errors, err.Error())
	}

	threadID := 0
	if data.ThreadID != "" || (ok && resource.RequiresThread) {
		parsedID, err := parseShareToolsID(data.ThreadID, "Thread ID")
		if err != nil {
			data.Errors = append(data.Errors, err.Error())
		}
		threadID = parsedID
	}

	if ok && resource.RequiresThread && threadID == 0 {
		data.Errors = append(data.Errors, "Thread ID is required for the selected resource type.")
	}

	if cd.ShareSignKey == "" {
		data.Errors = append(data.Errors, "Share signing key is not configured.")
	}

	if len(data.Errors) == 0 {
		basePath, err := resource.BuildPath(primaryID, threadID)
		if err != nil {
			data.Errors = append(data.Errors, err.Error())
		} else {
			sharedPath := signutil.InjectShared(basePath)
			signedPath, err := signutil.SignSharePath(sharedPath, cd.ShareSignKey, data.Duration, data.NoExpiry)
			if err != nil {
				data.Errors = append(data.Errors, fmt.Sprintf("Unable to sign URL: %v", err))
			} else {
				data.SharedPath = sharedPath
				data.SignedURL = cd.AbsoluteURL(signedPath)
			}
		}
	}

	AdminShareToolsPageTmpl.Handler(data).ServeHTTP(w, r)
}

func (p *AdminShareToolsPage) Breadcrumb() (string, string, common.HasBreadcrumb) {
	return "Share Tools", "/admin/share/tools", &AdminPage{}
}

func (p *AdminShareToolsPage) PageTitle() string {
	return "Share Tools"
}

var _ common.Page = (*AdminShareToolsPage)(nil)
var _ http.Handler = (*AdminShareToolsPage)(nil)

func shareToolsResourceList() ([]shareToolsResourceOption, map[string]shareToolsResource) {
	resources := []shareToolsResource{
		{
			Value:     "blogs",
			Label:     "Blog post",
			Hint:      "/blogs/blog/{blogID}",
			BuildPath: func(primaryID int, _ int) (string, error) { return fmt.Sprintf("/blogs/blog/%d", primaryID), nil },
		},
		{
			Value:     "news",
			Label:     "News post",
			Hint:      "/news/news/{newsID}",
			BuildPath: func(primaryID int, _ int) (string, error) { return fmt.Sprintf("/news/news/%d", primaryID), nil },
		},
		{
			Value:     "writings",
			Label:     "Writing article",
			Hint:      "/writings/article/{writingID}",
			BuildPath: func(primaryID int, _ int) (string, error) { return fmt.Sprintf("/writings/article/%d", primaryID), nil },
		},
		{
			Value:     "forum-topic",
			Label:     "Forum topic",
			Hint:      "/forum/topic/{topicID}",
			BuildPath: func(primaryID int, _ int) (string, error) { return fmt.Sprintf("/forum/topic/%d", primaryID), nil },
		},
		{
			Value:          "forum-thread",
			Label:          "Forum thread",
			Hint:           "/forum/topic/{topicID}/thread/{threadID}",
			RequiresThread: true,
			BuildPath: func(primaryID int, threadID int) (string, error) {
				return fmt.Sprintf("/forum/topic/%d/thread/%d", primaryID, threadID), nil
			},
		},
		{
			Value:     "private-topic",
			Label:     "Private forum topic",
			Hint:      "/private/topic/{topicID}",
			BuildPath: func(primaryID int, _ int) (string, error) { return fmt.Sprintf("/private/topic/%d", primaryID), nil },
		},
		{
			Value:          "private-thread",
			Label:          "Private forum thread",
			Hint:           "/private/topic/{topicID}/thread/{threadID}",
			RequiresThread: true,
			BuildPath: func(primaryID int, threadID int) (string, error) {
				return fmt.Sprintf("/private/topic/%d/thread/%d", primaryID, threadID), nil
			},
		},
	}

	resourceMap := make(map[string]shareToolsResource, len(resources))
	options := make([]shareToolsResourceOption, 0, len(resources))
	for _, resource := range resources {
		resourceMap[resource.Value] = resource
		options = append(options, shareToolsResourceOption{
			Value:          resource.Value,
			Label:          resource.Label,
			Hint:           resource.Hint,
			RequiresThread: resource.RequiresThread,
		})
	}

	return options, resourceMap
}

func parseShareToolsID(value string, label string) (int, error) {
	if value == "" {
		return 0, fmt.Errorf("%s is required.", label)
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive number.", label)
	}
	return parsed, nil
}

// AdminShareToolsPageTmpl is the template for the admin share tools page.
const AdminShareToolsPageTmpl tasks.Template = "admin/shareToolsPage.gohtml"

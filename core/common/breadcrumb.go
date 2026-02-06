package common

import (
	"fmt"
	"log"
	"strings"

	"github.com/arran4/goa4web/internal/tasks"
)

// Breadcrumb represents a single navigation step.
type Breadcrumb struct {
	Title string
	Link  string
}

// Breadcrumbs builds the breadcrumb trail for the current section based on
// selection information stored on CoreData. It returns nil if no breadcrumbs
// are applicable.
func (cd *CoreData) Breadcrumbs() []Breadcrumb {
	if cd == nil || cd.queries == nil {
		return nil
	}

	if hb := cd.currentPage; hb != nil {
		return buildBreadcrumbs(hb)
	}
	if cd.event != nil && cd.event.Task != nil {
		if hb, ok := cd.event.Task.(tasks.HasBreadcrumb); ok {
			return buildBreadcrumbs(hb)
		}
	}

	var (
		crumbs []Breadcrumb
		err    error
	)
	switch cd.currentSection {
	case "forum", "privateforum":
		crumbs, err = cd.forumBreadcrumbs()
	case "writings":
		crumbs, err = cd.writingBreadcrumbs()
	case "linker":
		crumbs, err = cd.linkerBreadcrumbs()
	case "imagebbs":
		crumbs, err = cd.imageboardBreadcrumbs()
	case "admin":
		crumbs, err = cd.adminBreadcrumbs()
	default:
		return nil
	}
	if err != nil {
		log.Printf("breadcrumbs: %v", err)
	}
	if cd.PageTitle != "" && len(crumbs) > 0 {
		crumbs = crumbs[:len(crumbs)-1]
	}
	return crumbs
}

func buildBreadcrumbs(hb tasks.HasBreadcrumb) []Breadcrumb {
	var crumbs []Breadcrumb
	for hb != nil {
		title, link, parent := hb.Breadcrumb()
		if title != "" {
			// Prepend because we traverse from child to parent
			crumbs = append([]Breadcrumb{{Title: title, Link: link}}, crumbs...)
		}
		hb = parent
	}
	return crumbs
}

func (cd *CoreData) forumBreadcrumbs() ([]Breadcrumb, error) {
	base := cd.ForumBasePath
	if base == "" {
		base = "/forum"
	}
	crumbTitle := "Forum"
	if cd != nil && cd.currentSection == "privateforum" {
		crumbTitle = "Private"
	}
	crumbs := []Breadcrumb{{Title: crumbTitle, Link: base}}
	catID := cd.currentCategoryID
	topicID := cd.currentTopicID
	threadID := cd.currentThreadID

	if threadID != 0 && topicID == 0 {
		if th, err := cd.SelectedThread(); err == nil && th != nil {
			topicID = th.ForumtopicIdforumtopic
		}
	}
	if topicID != 0 && catID == 0 {
		if tp, err := cd.ForumTopicByID(topicID); err == nil && tp != nil {
			catID = tp.ForumcategoryIdforumcategory
		}
	}
	if catID != 0 {
		rows, err := cd.queries.ListForumcategoryPath(cd.ctx, catID)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			title := fmt.Sprintf("Category %d", row.Idforumcategory)
			if row.Title.Valid {
				title = row.Title.String
			}
			link := fmt.Sprintf("%s/category/%d", base, row.Idforumcategory)
			if row.Idforumcategory == catID && topicID == 0 && threadID == 0 {
				crumbs = append(crumbs, Breadcrumb{Title: title})
			} else {
				crumbs = append(crumbs, Breadcrumb{Title: title, Link: link})
			}
		}
	}
	if topicID != 0 {
		if topic, err := cd.ForumTopicByID(topicID); err == nil && topic != nil {
			title := fmt.Sprintf("Topic %d", topicID)
			if topic.Title.Valid {
				title = topic.Title.String
			}
			if topic.Handler == "private" {
				title = cd.GetPrivateTopicDisplayTitle(topic.Idforumtopic, title)
			}
			link := fmt.Sprintf("%s/topic/%d", base, topicID)
			if threadID == 0 {
				crumbs = append(crumbs, Breadcrumb{Title: title})
			} else {
				crumbs = append(crumbs, Breadcrumb{Title: title, Link: link})
			}
		}
	}
	if threadID != 0 {
		crumbs = append(crumbs, Breadcrumb{Title: fmt.Sprintf("Thread %d", threadID)})
	}
	return crumbs, nil
}

func (cd *CoreData) writingBreadcrumbs() ([]Breadcrumb, error) {
	crumbs := []Breadcrumb{{Title: "Writings", Link: "/writings"}}
	catID := cd.currentCategoryID
	var writingTitle string
	if cd.currentWritingID != 0 {
		if w, err := cd.CurrentWriting(); err == nil && w != nil {
			catID = w.WritingCategoryID
			writingTitle = fmt.Sprintf("Writing %d", w.Idwriting)
			if w.Title.Valid {
				writingTitle = w.Title.String
			}
		}
	}
	if catID != 0 {
		rows, err := cd.queries.ListWritingcategoryPath(cd.ctx, catID)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			title := fmt.Sprintf("Category %d", row.Idwritingcategory)
			if row.Title.Valid {
				title = row.Title.String
			}
			link := fmt.Sprintf("/writings/category/%d", row.Idwritingcategory)
			if row.Idwritingcategory == catID && writingTitle == "" {
				crumbs = append(crumbs, Breadcrumb{Title: title})
			} else {
				crumbs = append(crumbs, Breadcrumb{Title: title, Link: link})
			}
		}
	}
	if writingTitle != "" {
		crumbs = append(crumbs, Breadcrumb{Title: writingTitle})
	}
	return crumbs, nil
}

func (cd *CoreData) linkerBreadcrumbs() ([]Breadcrumb, error) {
	crumbs := []Breadcrumb{{Title: "Linker", Link: "/linker"}}
	catID := cd.currentCategoryID
	if catID != 0 {
		rows, err := cd.queries.ListLinkerCategoryPath(cd.ctx, catID)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			title := fmt.Sprintf("Category %d", row.ID)
			if row.Title.Valid {
				title = row.Title.String
			}
			link := fmt.Sprintf("/linker/category/%d", row.ID)
			if row.ID == catID {
				crumbs = append(crumbs, Breadcrumb{Title: title})
			} else {
				crumbs = append(crumbs, Breadcrumb{Title: title, Link: link})
			}
		}
	}
	return crumbs, nil
}

func (cd *CoreData) imageboardBreadcrumbs() ([]Breadcrumb, error) {
	crumbs := []Breadcrumb{{Title: "ImageBBS", Link: "/imagebbs"}}
	boardID := cd.currentBoardID
	threadID := cd.currentThreadID
	if boardID != 0 {
		rows, err := cd.queries.ListImageboardPath(cd.ctx, boardID)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			title := fmt.Sprintf("Board %d", row.Idimageboard)
			if row.Title.Valid {
				title = row.Title.String
			}
			link := fmt.Sprintf("/imagebbs/board/%d", row.Idimageboard)
			if row.Idimageboard == boardID && threadID == 0 {
				crumbs = append(crumbs, Breadcrumb{Title: title})
			} else {
				crumbs = append(crumbs, Breadcrumb{Title: title, Link: link})
			}
		}
	}
	if threadID != 0 {
		crumbs = append(crumbs, Breadcrumb{Title: fmt.Sprintf("Thread %d", threadID)})
	}
	return crumbs, nil
}

func (cd *CoreData) adminBreadcrumbs() ([]Breadcrumb, error) {
	crumbs := []Breadcrumb{{Title: "Admin", Link: "/admin"}}
	switch {
	case cd.currentProfileUserID != 0:
		crumbs = append(crumbs, Breadcrumb{Title: "Users", Link: "/admin/user"})
		if u := cd.CurrentProfileUser(); u != nil {
			title := fmt.Sprintf("User %d", u.Idusers)
			if u.Username.Valid {
				title = fmt.Sprintf("User %s", u.Username.String)
			}
			crumbs = append(crumbs, Breadcrumb{Title: title})
		} else {
			crumbs = append(crumbs, Breadcrumb{Title: fmt.Sprintf("User %d", cd.currentProfileUserID)})
		}
	case cd.currentRoleID != 0:
		crumbs = append(crumbs, Breadcrumb{Title: "Roles", Link: "/admin/roles"})
		if r, err := cd.RoleByID(cd.currentRoleID); err == nil && r != nil {
			title := fmt.Sprintf("Role %d", cd.currentRoleID)
			if r.Name != "" {
				title = fmt.Sprintf("Role %s", r.Name)
			}
			crumbs = append(crumbs, Breadcrumb{Title: title})
		} else {
			crumbs = append(crumbs, Breadcrumb{Title: fmt.Sprintf("Role %d", cd.currentRoleID)})
		}
	case cd.currentCategoryID != 0 && strings.HasPrefix(cd.PageTitle, "Linker Category"):
		crumbs = append(crumbs, Breadcrumb{Title: "Linker Categories", Link: "/admin/linker/categories"})
		crumbs = append(crumbs, Breadcrumb{Title: cd.PageTitle})
	case cd.currentCategoryID != 0 && strings.HasPrefix(cd.PageTitle, "Edit Category"):
		crumbs = append(crumbs, Breadcrumb{Title: "Writing Categories", Link: "/admin/writings/categories"})
		crumbs = append(crumbs, Breadcrumb{Title: fmt.Sprintf("Writing Category %d", cd.currentCategoryID), Link: fmt.Sprintf("/admin/writings/categories/category/%d", cd.currentCategoryID)})
		crumbs = append(crumbs, Breadcrumb{Title: cd.PageTitle})
	case cd.currentCategoryID != 0 && strings.Contains(cd.PageTitle, "Category Grants"):
		crumbs = append(crumbs, Breadcrumb{Title: "Writing Categories", Link: "/admin/writings/categories"})
		crumbs = append(crumbs, Breadcrumb{Title: fmt.Sprintf("Writing Category %d", cd.currentCategoryID), Link: fmt.Sprintf("/admin/writings/categories/category/%d", cd.currentCategoryID)})
		crumbs = append(crumbs, Breadcrumb{Title: cd.PageTitle})
	case cd.currentCategoryID != 0 && strings.HasPrefix(cd.PageTitle, "Writing Category"):
		crumbs = append(crumbs, Breadcrumb{Title: "Writing Categories", Link: "/admin/writings/categories"})
		crumbs = append(crumbs, Breadcrumb{Title: cd.PageTitle})
	case strings.Contains(cd.PageTitle, "FAQ Category") && !strings.Contains(cd.PageTitle, "FAQ Categories"):
		crumbs = append(crumbs, Breadcrumb{Title: "FAQ Categories", Link: "/admin/faq/categories"})
		crumbs = append(crumbs, Breadcrumb{Title: cd.PageTitle})
	case cd.currentLinkID != 0 && strings.HasPrefix(cd.PageTitle, "Link"):
		crumbs = append(crumbs, Breadcrumb{Title: "Linker Links", Link: "/admin/linker/links"})
		crumbs = append(crumbs, Breadcrumb{Title: cd.PageTitle})
	case cd.currentBoardID != 0 && strings.Contains(cd.PageTitle, "Image Board"):
		crumbs = append(crumbs, Breadcrumb{Title: "Image Boards", Link: "/admin/imagebbs/boards"})
		title := fmt.Sprintf("Board %d", cd.currentBoardID)
		if boards, err := cd.ImageBoards(); err == nil {
			for _, b := range boards {
				if b.Idimageboard == cd.currentBoardID {
					if b.Title.Valid {
						title = b.Title.String
					}
					break
				}
			}
		}
		crumbs = append(crumbs, Breadcrumb{Title: title})
	case cd.currentRequestID != 0:
		crumbs = append(crumbs, Breadcrumb{Title: "Requests", Link: "/admin/requests"})
		crumbs = append(crumbs, Breadcrumb{Title: fmt.Sprintf("Request %d", cd.currentRequestID)})
	case cd.currentTopicID != 0 && (strings.Contains(cd.PageTitle, "Forum Topic") || strings.Contains(cd.PageTitle, "Edit Forum Topic")):
		crumbs = append(crumbs, Breadcrumb{Title: "Forum Topics", Link: "/admin/forum/topics"})
		if t, err := cd.CurrentTopic(); err == nil && t != nil {
			title := fmt.Sprintf("Topic %d", cd.currentTopicID)
			if t.Title.Valid {
				title = t.Title.String
			}
			crumbs = append(crumbs, Breadcrumb{Title: title, Link: fmt.Sprintf("/admin/forum/topics/topic/%d", cd.currentTopicID)})
		}
		if cd.PageTitle != "" {
			crumbs = append(crumbs, Breadcrumb{Title: cd.PageTitle})
		}
	case strings.HasPrefix(cd.PageTitle, "Email"):
		crumbs = append(crumbs, Breadcrumb{Title: "Email", Link: "/admin/email/queue"})
		crumbs = append(crumbs, Breadcrumb{Title: cd.PageTitle})
	case strings.HasPrefix(cd.PageTitle, "Comment"):
		if cd.PageTitle != "Comments" {
			crumbs = append(crumbs, Breadcrumb{Title: "Comments", Link: "/admin/comments"})
		}
		crumbs = append(crumbs, Breadcrumb{Title: cd.PageTitle})
	case strings.Contains(cd.PageTitle, "Announcements"):
		if cd.PageTitle != "Admin Announcements" && cd.PageTitle != "Announcements" {
			crumbs = append(crumbs, Breadcrumb{Title: "Announcements", Link: "/admin/announcements"})
		}
		crumbs = append(crumbs, Breadcrumb{Title: cd.PageTitle})
	case strings.HasPrefix(cd.PageTitle, "Database"):
		crumbs = append(crumbs, Breadcrumb{Title: "Database", Link: "/admin/db/status"})
		crumbs = append(crumbs, Breadcrumb{Title: cd.PageTitle})
	case strings.HasPrefix(cd.PageTitle, "Site Settings"):
		crumbs = append(crumbs, Breadcrumb{Title: cd.PageTitle})
	case strings.HasPrefix(cd.PageTitle, "Server Stats"):
		crumbs = append(crumbs, Breadcrumb{Title: cd.PageTitle})
	case strings.HasPrefix(cd.PageTitle, "IP Bans"):
		crumbs = append(crumbs, Breadcrumb{Title: cd.PageTitle})
	default:
		if cd.PageTitle != "" {
			crumbs = append(crumbs, Breadcrumb{Title: cd.PageTitle})
		} else {
			crumbs = append(crumbs, Breadcrumb{Title: "Admin"})
		}
	}
	return crumbs, nil
}

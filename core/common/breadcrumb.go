package common

import (
	"fmt"
	"log"
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
	var (
		crumbs []Breadcrumb
		err    error
	)
	switch cd.currentSection {
	case "forum":
		crumbs, err = cd.forumBreadcrumbs()
	case "writings":
		crumbs, err = cd.writingBreadcrumbs()
	case "linker":
		crumbs, err = cd.linkerBreadcrumbs()
	case "imagebbs":
		crumbs, err = cd.imageboardBreadcrumbs()
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

func (cd *CoreData) forumBreadcrumbs() ([]Breadcrumb, error) {
	crumbs := []Breadcrumb{{Title: "Forum", Link: "/forum"}}
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
			link := fmt.Sprintf("/forum/category/%d", row.Idforumcategory)
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
			link := fmt.Sprintf("/forum/topic/%d", topicID)
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
			title := fmt.Sprintf("Category %d", row.Idlinkercategory)
			if row.Title.Valid {
				title = row.Title.String
			}
			link := fmt.Sprintf("/linker/category/%d", row.Idlinkercategory)
			if row.Idlinkercategory == catID {
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

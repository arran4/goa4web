package common

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/arran4/goa4web/internal/algorithms"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/lazy"
)

const (
	WritingTopicName        = "A WRITING TOPIC"
	WritingTopicDescription = "THIS IS A HIDDEN FORUM FOR A WRITING"
)

// Article returns the currently requested writing.
func (cd *CoreData) Article(ops ...lazy.Option[*db.GetWritingForListerByIDRow]) (*db.GetWritingForListerByIDRow, error) {
	return cd.CurrentWriting(ops...)
}

// ArticleComments returns comments for the current article.
func (cd *CoreData) ArticleComments() ([]*db.GetCommentsByThreadIdForUserRow, error) {
	w, err := cd.Article()
	if err != nil || w == nil {
		return nil, err
	}
	return cd.SectionThreadComments("writing", "article", w.ForumthreadID)
}

// WritingCategories returns all writing categories cached once.
func (cd *CoreData) WritingCategories() ([]*db.WritingCategory, error) {
	return cd.writingCategories.Load(func() ([]*db.WritingCategory, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.SystemListWritingCategories(cd.ctx, db.SystemListWritingCategoriesParams{Limit: math.MaxInt32, Offset: 0})
	})
}

// EditableArticle returns the current article if owned by the user.
func (cd *CoreData) EditableArticle() (*db.GetWritingForListerByIDRow, error) {
	w, err := cd.Article()
	if err != nil || w == nil {
		return w, err
	}
	if w.UsersIdusers != cd.UserID {
		return nil, sql.ErrNoRows
	}
	return w, nil
}

// ArticleComment returns the requested comment for the article.
func (cd *CoreData) ArticleComment(r *http.Request, ops ...lazy.Option[*db.GetCommentByIdForUserRow]) (*db.GetCommentByIdForUserRow, error) {
	return cd.CurrentComment(r, ops...)
}

// UpdateArticleComment updates a comment on a writing.
func (cd *CoreData) UpdateArticleComment(commentID, languageID int32, text string) error {
	uid := cd.UserID
	paths, err := cd.imagePathsFromText(text)
	if err != nil {
		return fmt.Errorf("parse images: %w", err)
	}
	comment, err := cd.CommentByID(commentID)
	if err != nil || comment == nil {
		return fmt.Errorf("load comment: %w", err)
	}
	if err := cd.validateImagePathsForThread(uid, comment.ForumthreadID, paths); err != nil {
		return fmt.Errorf("validate images: %w", err)
	}
	if err := cd.queries.UpdateCommentForEditor(cd.ctx, db.UpdateCommentForEditorParams{
		LanguageID:  sql.NullInt32{Int32: languageID, Valid: languageID != 0},
		Text:        sql.NullString{String: text, Valid: true},
		CommentID:   commentID,
		CommenterID: uid,
		EditorID:    sql.NullInt32{Int32: uid, Valid: uid != 0},
	}); err != nil {
		return err
	}
	if err := cd.recordThreadImages(comment.ForumthreadID, paths); err != nil {
		log.Printf("record thread images: %v", err)
	}
	return nil
}

// WriterByUsername fetches a user by username.
func (cd *CoreData) WriterByUsername(username string) (*db.SystemGetUserByUsernameRow, error) {
	if cd.queries == nil {
		return nil, nil
	}
	return cd.queries.SystemGetUserByUsername(cd.ctx, sql.NullString{String: username, Valid: true})
}

// WriterWritings returns public writings for the specified author respecting cd's permissions.
func (cd *CoreData) WriterWritings(userID int32, r *http.Request) ([]*db.ListPublicWritingsByUserForListerRow, error) {
	if cd.writerWritings == nil {
		cd.writerWritings = map[int32]*lazy.Value[[]*db.ListPublicWritingsByUserForListerRow]{}
	}
	lv, ok := cd.writerWritings[userID]
	if !ok {
		lv = &lazy.Value[[]*db.ListPublicWritingsByUserForListerRow]{}
		cd.writerWritings[userID] = lv
	}
	return lv.Load(func() ([]*db.ListPublicWritingsByUserForListerRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		rows, err := cd.queries.ListPublicWritingsByUserForLister(cd.ctx, db.ListPublicWritingsByUserForListerParams{
			ListerID: cd.UserID,
			AuthorID: userID,
			UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
			Limit:    int32(cd.PageSize()),
			Offset:   int32(offset),
		})
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		var list []*db.ListPublicWritingsByUserForListerRow
		for _, row := range rows {
			if !cd.HasGrant("writing", "article", "see", row.Idwriting) {
				continue
			}
			list = append(list, row)
		}
		return list, nil
	})
}

// UpdateWritingReply updates a comment reply and returns thread metadata.
func (cd *CoreData) UpdateWritingReply(commentID, languageID int32, text string) (*db.GetThreadLastPosterAndPermsRow, error) {
	cmt, err := cd.CommentByID(commentID)
	if err != nil || cmt == nil {
		return nil, err
	}
	uid := cd.UserID
	paths, err := cd.imagePathsFromText(text)
	if err != nil {
		return nil, fmt.Errorf("parse images: %w", err)
	}
	if err := cd.validateImagePathsForThread(uid, cmt.ForumthreadID, paths); err != nil {
		return nil, fmt.Errorf("validate images: %w", err)
	}
	thread, err := cd.queries.GetThreadLastPosterAndPerms(cd.ctx, db.GetThreadLastPosterAndPermsParams{
		ViewerID:      uid,
		ThreadID:      cmt.ForumthreadID,
		ViewerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if err := cd.queries.UpdateCommentForEditor(cd.ctx, db.UpdateCommentForEditorParams{
		LanguageID:  sql.NullInt32{Int32: languageID, Valid: languageID != 0},
		Text:        sql.NullString{String: text, Valid: true},
		CommentID:   cmt.Idcomments,
		CommenterID: uid,
		EditorID:    sql.NullInt32{Int32: uid, Valid: uid != 0},
	}); err != nil {
		return nil, err
	}
	if err := cd.recordThreadImages(cmt.ForumthreadID, paths); err != nil {
		log.Printf("record thread images: %v", err)
	}
	return thread, nil
}

// CreateWritingReply creates a comment reply and ensures the thread exists.
func (cd *CoreData) CreateWritingReply(w *db.GetWritingForListerByIDRow, languageID int32, text string) (int64, int32, int32, error) {
	if cd.queries == nil || w == nil {
		return 0, 0, 0, fmt.Errorf("invalid writing")
	}
	pthid := w.ForumthreadID
	pt, err := cd.queries.SystemGetForumTopicByTitle(cd.ctx, sql.NullString{String: WritingTopicName, Valid: true})
	var ptid int32
	if errors.Is(err, sql.ErrNoRows) {
		ptidi, err := cd.queries.CreateForumTopicForPoster(cd.ctx, db.CreateForumTopicForPosterParams{
			ForumcategoryID: 0,
			ForumLang:       w.LanguageID,
			Title:           sql.NullString{String: WritingTopicName, Valid: true},
			Description:     sql.NullString{String: WritingTopicDescription, Valid: true},
			Handler:         "writing",
			Section:         "forum",
			GrantCategoryID: sql.NullInt32{},
			GranteeID:       sql.NullInt32{},
			PosterID:        0,
		})
		if err != nil {
			return 0, 0, 0, err
		}
		ptid = int32(ptidi)
	} else if err != nil {
		return 0, 0, 0, err
	} else {
		ptid = pt.Idforumtopic
	}
	if pthid == 0 {
		pthidi, err := cd.queries.SystemCreateThread(cd.ctx, ptid)
		if err != nil {
			return 0, 0, 0, err
		}
		pthid = int32(pthidi)
		if err := cd.queries.SystemAssignWritingThreadID(cd.ctx, db.SystemAssignWritingThreadIDParams{ForumthreadID: pthid, Idwriting: w.Idwriting}); err != nil {
			return 0, 0, 0, err
		}
	}
	cid, err := cd.CreateWritingCommentForCommenter(cd.UserID, pthid, w.Idwriting, languageID, text)
	if err != nil {
		return 0, 0, 0, err
	}
	return cid, pthid, ptid, nil
}

// UpdateWriting updates an existing article.
func (cd *CoreData) UpdateWriting(w *db.GetWritingForListerByIDRow, title, abstract, body string, private bool, languageID int32) error {
	if cd.queries == nil || w == nil {
		return fmt.Errorf("invalid writing")
	}
	if err := cd.validateCodeImagesForUser(cd.UserID, title); err != nil {
		return fmt.Errorf("validate title images: %w", err)
	}
	if err := cd.validateCodeImagesForUser(cd.UserID, abstract); err != nil {
		return fmt.Errorf("validate abstract images: %w", err)
	}
	if err := cd.validateCodeImagesForUser(cd.UserID, body); err != nil {
		return fmt.Errorf("validate body images: %w", err)
	}
	return cd.queries.UpdateWritingForWriter(cd.ctx, db.UpdateWritingForWriterParams{
		Title:      sql.NullString{Valid: true, String: title},
		Abstract:   sql.NullString{Valid: true, String: abstract},
		Content:    sql.NullString{Valid: true, String: body},
		Private:    sql.NullBool{Valid: true, Bool: private},
		LanguageID: sql.NullInt32{Int32: languageID, Valid: languageID != 0},
		WritingID:  w.Idwriting,
		WriterID:   cd.UserID,
		GranteeID:  sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
}

// CreateWriting creates a new article in the given category.
func (cd *CoreData) CreateWriting(categoryID, languageID int32, title, abstract, body string, private bool) (int64, error) {
	if cd.queries == nil {
		return 0, fmt.Errorf("no queries")
	}
	if !cd.HasGrant("writing", "category", "post", categoryID) {
		return 0, sql.ErrNoRows
	}
	if err := cd.validateCodeImagesForUser(cd.UserID, title); err != nil {
		return 0, fmt.Errorf("validate title images: %w", err)
	}
	if err := cd.validateCodeImagesForUser(cd.UserID, abstract); err != nil {
		return 0, fmt.Errorf("validate abstract images: %w", err)
	}
	if err := cd.validateCodeImagesForUser(cd.UserID, body); err != nil {
		return 0, fmt.Errorf("validate body images: %w", err)
	}
	return cd.queries.CreateWritingForWriter(cd.ctx, db.CreateWritingForWriterParams{
		WriterID:          cd.UserID,
		WritingCategoryID: categoryID,
		Title:             sql.NullString{Valid: true, String: title},
		Abstract:          sql.NullString{Valid: true, String: abstract},
		Writing:           sql.NullString{Valid: true, String: body},
		Private:           sql.NullBool{Valid: true, Bool: private},
		LanguageID:        sql.NullInt32{Int32: languageID, Valid: languageID != 0},
		Published:         sql.NullTime{Time: time.Now().UTC(), Valid: true},
		Timezone:          sql.NullString{String: cd.Location().String(), Valid: true},
		GrantCategoryID:   sql.NullInt32{Int32: categoryID, Valid: true},
		GranteeID:         sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
}

// GrantWritingCategory adds a grant for a writing category.
func (cd *CoreData) GrantWritingCategory(categoryID, uid, rid int32, actions []string) error {
	if cd.queries == nil {
		return fmt.Errorf("no queries")
	}
	var user sql.NullInt32
	if uid != 0 {
		user = sql.NullInt32{Int32: uid, Valid: true}
	}
	var role sql.NullInt32
	if rid != 0 {
		role = sql.NullInt32{Int32: rid, Valid: true}
	}
	if len(actions) == 0 {
		actions = []string{"see"}
	}
	for _, action := range actions {
		if action == "" {
			action = "see"
		}
		if _, err := cd.queries.AdminCreateGrant(cd.ctx, db.AdminCreateGrantParams{
			UserID:   user,
			RoleID:   role,
			Section:  "writing",
			Item:     sql.NullString{String: "category", Valid: true},
			RuleType: "allow",
			ItemID:   sql.NullInt32{Int32: categoryID, Valid: true},
			Action:   action,
		}); err != nil {
			return err
		}
	}
	return nil
}

// RevokeWritingCategory removes a grant.
func (cd *CoreData) RevokeWritingCategory(grantID int32) error {
	if cd.queries == nil {
		return fmt.Errorf("no queries")
	}
	return cd.queries.AdminDeleteGrant(cd.ctx, grantID)
}

// CreateWritingCategory inserts a new writing category.
func (cd *CoreData) CreateWritingCategory(parentID int32, name, desc string) error {
	if cd.queries == nil {
		return fmt.Errorf("no queries")
	}
	cats, err := cd.queries.SystemListWritingCategories(cd.ctx, db.SystemListWritingCategoriesParams{Limit: math.MaxInt32, Offset: 0})
	if err != nil {
		return err
	}
	parents := make(map[int32]int32, len(cats))
	for _, c := range cats {
		var pid int32
		if c.WritingCategoryID.Valid {
			pid = c.WritingCategoryID.Int32
		}
		parents[c.Idwritingcategory] = pid
	}
	if path, loop := algorithms.WouldCreateLoop(parents, 0, parentID); loop && len(path) > 0 {
		return UserError{ErrorMessage: "invalid parent category: loop detected"}
	}
	return cd.queries.AdminInsertWritingCategory(cd.ctx, db.AdminInsertWritingCategoryParams{
		WritingCategoryID: sql.NullInt32{Int32: parentID, Valid: parentID != 0},
		Title: sql.NullString{
			Valid:  true,
			String: name,
		},
		Description: sql.NullString{
			Valid:  true,
			String: desc,
		},
	})
}

// ChangeWritingCategory updates an existing writing category.
func (cd *CoreData) ChangeWritingCategory(id, parentID int32, name, desc string) error {
	if cd.queries == nil {
		return fmt.Errorf("no queries")
	}
	cats, err := cd.queries.SystemListWritingCategories(cd.ctx, db.SystemListWritingCategoriesParams{Limit: math.MaxInt32, Offset: 0})
	if err != nil {
		return err
	}
	parents := make(map[int32]int32, len(cats))
	for _, c := range cats {
		var pid int32
		if c.WritingCategoryID.Valid {
			pid = c.WritingCategoryID.Int32
		}
		parents[c.Idwritingcategory] = pid
	}
	if path, loop := algorithms.WouldCreateLoop(parents, id, parentID); loop {
		return UserError{ErrorMessage: fmt.Sprintf("invalid parent category: loop %v", path)}
	}
	return cd.queries.AdminUpdateWritingCategory(cd.ctx, db.AdminUpdateWritingCategoryParams{
		Title:             sql.NullString{Valid: true, String: name},
		Description:       sql.NullString{Valid: true, String: desc},
		Idwritingcategory: id,
		WritingCategoryID: sql.NullInt32{Int32: parentID, Valid: parentID != 0},
	})
}

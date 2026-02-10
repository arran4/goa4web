package common

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/internal/db"
)

// linkerTopicName matches the hidden linker forum name.
const linkerTopicName = "A LINKER TOPIC"

// writingTopicName matches the hidden writing forum name.
const writingTopicName = "A WRITING TOPIC"

// bloggerTopicName matches the hidden blogger forum name.
const bloggerTopicName = "A BLOGGER TOPIC"

// SearchLinker populates linker and comment search results on cd.
func (cd *CoreData) SearchLinker(r *http.Request) error {
	uid := cd.UserID

	ftbn, err := cd.queries.SystemGetForumTopicByTitle(cd.ctx, sql.NullString{Valid: true, String: linkerTopicName})
	if err != nil {
		log.Printf("findForumTopicByTitle Error: %s", err)
		return ErrInternalServerError
	}

	comments, emptyWords, noResults, err := cd.forumCommentSearchInRestrictedTopic(r, []int32{ftbn.Idforumtopic}, uid)
	if err != nil {
		return err
	}
	cd.cache.searchComments = comments
	cd.cache.searchCommentsNoResults = emptyWords
	cd.cache.searchCommentsEmptyWords = noResults

	links, emptyWords2, noResults2, err := cd.linkerSearch(r, uid)
	if err != nil {
		return err
	}
	cd.cache.searchLinkerItems = links
	cd.cache.searchLinkerNoResults = emptyWords2
	cd.cache.searchLinkerEmptyWords = noResults2
	return nil
}

// SearchWritings populates writing and comment search results on cd.
func (cd *CoreData) SearchWritings(r *http.Request) error {
	uid := cd.UserID

	ftbn, err := cd.queries.SystemGetForumTopicByTitle(cd.ctx, sql.NullString{Valid: true, String: writingTopicName})
	if err != nil {
		log.Printf("findForumTopicByTitle Error: %s", err)
		return ErrInternalServerError
	}

	comments, emptyWords, noResults, err := cd.forumCommentSearchInRestrictedTopic(r, []int32{ftbn.Idforumtopic}, uid)
	if err != nil {
		return err
	}
	cd.cache.searchComments = comments
	cd.cache.searchCommentsNoResults = emptyWords
	cd.cache.searchCommentsEmptyWords = noResults

	writings, emptyWords2, noResults2, err := cd.writingSearch(r, uid)
	if err != nil {
		return err
	}
	cd.cache.searchWritings = writings
	cd.cache.searchWritingsNoResults = emptyWords2
	cd.cache.searchWritingsEmptyWords = noResults2
	return nil
}

// SearchBlogs populates blog and comment search results on cd.
func (cd *CoreData) SearchBlogs(r *http.Request) error {
	uid := cd.UserID

	ftbn, err := cd.queries.SystemGetForumTopicByTitle(cd.ctx, sql.NullString{Valid: true, String: bloggerTopicName})
	if err != nil {
		log.Printf("findForumTopicByTitle Error: %s", err)
		return ErrInternalServerError
	}

	comments, emptyWords, noResults, err := cd.forumCommentSearchInRestrictedTopic(r, []int32{ftbn.Idforumtopic}, uid)
	if err != nil {
		return err
	}
	cd.cache.searchComments = comments
	cd.cache.searchCommentsNoResults = emptyWords
	cd.cache.searchCommentsEmptyWords = noResults

	blogs, emptyWords2, noResults2, err := cd.blogSearch(r, uid)
	if err != nil {
		return err
	}
	cd.cache.searchBlogs = blogs
	cd.cache.searchBlogsNoResults = emptyWords2
	cd.cache.searchBlogsEmptyWords = noResults2
	return nil
}

// SearchForum populates forum comment search results on cd.
func (cd *CoreData) SearchForum(r *http.Request) error {
	uid := cd.UserID
	comments, emptyWords, noResults, err := cd.forumCommentSearchNotInRestrictedTopic(r, uid)
	if err != nil {
		return err
	}
	cd.cache.searchComments = comments
	cd.cache.searchCommentsNoResults = emptyWords
	cd.cache.searchCommentsEmptyWords = noResults
	return nil
}

// SearchComments returns forum comment search results.
func (cd *CoreData) SearchComments() []*db.GetCommentsByIdsForUserWithThreadInfoRow {
	return cd.cache.searchComments
}

// SearchCommentsNoResults reports whether the comment search had no matches.
func (cd *CoreData) SearchCommentsNoResults() bool {
	return cd.cache.searchCommentsNoResults
}

// SearchCommentsEmptyWords reports whether the comment search lacked words.
func (cd *CoreData) SearchCommentsEmptyWords() bool {
	return cd.cache.searchCommentsEmptyWords
}

// SearchLinkerItems returns linker search results.
func (cd *CoreData) SearchLinkerItems() []*db.GetLinkerItemsByIdsWithPosterUsernameAndCategoryTitleDescendingRow {
	return cd.cache.searchLinkerItems
}

// SearchLinkerNoResults reports whether the linker search found nothing.
func (cd *CoreData) SearchLinkerNoResults() bool {
	return cd.cache.searchLinkerNoResults
}

// SearchLinkerEmptyWords reports whether the linker search lacked words.
func (cd *CoreData) SearchLinkerEmptyWords() bool {
	return cd.cache.searchLinkerEmptyWords
}

// SearchWritingsResults returns writing search results.
func (cd *CoreData) SearchWritingsResults() []*db.ListWritingsByIDsForListerRow {
	return cd.cache.searchWritings
}

// SearchWritingsNoResults reports whether the writing search found nothing.
func (cd *CoreData) SearchWritingsNoResults() bool {
	return cd.cache.searchWritingsNoResults
}

// SearchWritingsEmptyWords reports whether the writing search lacked words.
func (cd *CoreData) SearchWritingsEmptyWords() bool {
	return cd.cache.searchWritingsEmptyWords
}

// SearchBlogsResults returns blog search results.
func (cd *CoreData) SearchBlogsResults() []*db.Blog {
	return cd.cache.searchBlogs
}

// SearchBlogsNoResults reports whether the blog search found nothing.
func (cd *CoreData) SearchBlogsNoResults() bool {
	return cd.cache.searchBlogsNoResults
}

// SearchBlogsEmptyWords reports whether the blog search lacked words.
func (cd *CoreData) SearchBlogsEmptyWords() bool {
	return cd.cache.searchBlogsEmptyWords
}

func (cd *CoreData) linkerSearch(r *http.Request, uid int32) ([]*db.GetLinkerItemsByIdsWithPosterUsernameAndCategoryTitleDescendingRow, bool, bool, error) {
	searchWords := cd.searchWordsFromRequest(r)
	var linkerIDs []int32

	if len(searchWords) == 0 {
		return nil, true, false, nil
	}

	for i, word := range searchWords {
		if i == 0 {
			ids, err := cd.queries.LinkerSearchFirst(cd.ctx, db.LinkerSearchFirstParams{
				ListerID: uid,
				Word:     sql.NullString{String: word, Valid: true},
				UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
			})
			if err != nil {
				if !errors.Is(err, sql.ErrNoRows) {
					log.Printf("LinkersSearchFirst Error: %s", err)
					return nil, false, false, ErrInternalServerError
				}
			}
			linkerIDs = ids
		} else {
			ids, err := cd.queries.LinkerSearchNext(cd.ctx, db.LinkerSearchNextParams{
				ListerID: uid,
				Word:     sql.NullString{String: word, Valid: true},
				Ids:      linkerIDs,
				UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
			})
			if err != nil {
				if !errors.Is(err, sql.ErrNoRows) {
					log.Printf("LinkersSearchNext Error: %s", err)
					return nil, false, false, ErrInternalServerError
				}
			}
			linkerIDs = ids
		}
		if len(linkerIDs) == 0 {
			return nil, false, true, nil
		}
	}

	links, err := cd.queries.GetLinkerItemsByIdsWithPosterUsernameAndCategoryTitleDescending(cd.ctx, linkerIDs)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("getLinkers Error: %s", err)
			return nil, false, false, ErrInternalServerError
		}
	}

	return links, false, false, nil
}

func (cd *CoreData) writingSearch(r *http.Request, uid int32) ([]*db.ListWritingsByIDsForListerRow, bool, bool, error) {
	searchWords := cd.searchWordsFromRequest(r)
	var writingsIDs []int32

	if len(searchWords) == 0 {
		return nil, true, false, nil
	}

	for i, word := range searchWords {
		if i == 0 {
			ids, err := cd.queries.ListWritingSearchFirstForLister(cd.ctx, db.ListWritingSearchFirstForListerParams{
				ListerID: uid,
				Word:     sql.NullString{String: word, Valid: true},
				UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
			})
			if err != nil {
				if !errors.Is(err, sql.ErrNoRows) {
					log.Printf("writingSearchFirst Error: %s", err)
					return nil, false, false, ErrInternalServerError
				}
			}
			writingsIDs = ids
		} else {
			ids, err := cd.queries.ListWritingSearchNextForLister(cd.ctx, db.ListWritingSearchNextForListerParams{
				ListerID: uid,
				Word:     sql.NullString{String: word, Valid: true},
				Ids:      writingsIDs,
				UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
			})
			if err != nil {
				if !errors.Is(err, sql.ErrNoRows) {
					log.Printf("writingSearchNext Error: %s", err)
					return nil, false, false, ErrInternalServerError
				}
			}
			writingsIDs = ids
		}
		if len(writingsIDs) == 0 {
			return nil, false, true, nil
		}
	}

	limit := int32(cd.Config.PageSizeDefault)
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	writings, err := cd.queries.ListWritingsByIDsForLister(cd.ctx, db.ListWritingsByIDsForListerParams{
		ListerID:      uid,
		ListerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
		WritingIds:    writingsIDs,
		Limit:         limit,
		Offset:        int32(offset),
	})
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("getWritings Error: %s", err)
			return nil, false, false, ErrInternalServerError
		}
	}

	return writings, false, false, nil
}

func (cd *CoreData) blogSearch(r *http.Request, uid int32) ([]*db.Blog, bool, bool, error) {
	searchWords := cd.searchWordsFromRequest(r)
	var blogIDs []int32

	if len(searchWords) == 0 {
		return nil, true, false, nil
	}

	for i, word := range searchWords {
		if i == 0 {
			ids, err := cd.queries.ListBlogIDsBySearchWordFirstForLister(cd.ctx, db.ListBlogIDsBySearchWordFirstForListerParams{
				ListerID: uid,
				Word:     sql.NullString{String: word, Valid: true},
				UserID:   sql.NullInt32{Int32: uid, Valid: true},
			})
			if err != nil {
				log.Printf("ListBlogIDsBySearchWordFirstForLister Error: %s", err)
				return nil, false, false, ErrInternalServerError
			}
			blogIDs = ids
		} else {
			ids, err := cd.queries.ListBlogIDsBySearchWordNextForLister(cd.ctx, db.ListBlogIDsBySearchWordNextForListerParams{
				ListerID: uid,
				Word:     sql.NullString{String: word, Valid: true},
				Ids:      blogIDs,
				UserID:   sql.NullInt32{Int32: uid, Valid: true},
			})
			if err != nil {
				log.Printf("ListBlogIDsBySearchWordNextForLister Error: %s", err)
				return nil, false, false, ErrInternalServerError
			}
			blogIDs = ids
		}
		if len(blogIDs) == 0 {
			return nil, false, true, nil
		}
	}

	limit := int32(cd.Config.PageSizeDefault)
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	rows, err := cd.queries.ListBlogEntriesByIDsForLister(cd.ctx, db.ListBlogEntriesByIDsForListerParams{
		ListerID: uid,
		UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
		Blogids:  blogIDs,
		Limit:    limit,
		Offset:   int32(offset),
	})
	if err != nil {
		log.Printf("getBlogEntriesByIdsDescending Error: %s", err)
		return nil, false, false, ErrInternalServerError
	}
	blogs := make([]*db.Blog, 0, len(rows))
	for _, r := range rows {
		blogs = append(blogs, &db.Blog{
			Idblogs:       r.Idblogs,
			ForumthreadID: r.ForumthreadID,
			UsersIdusers:  r.UsersIdusers,
			LanguageID:    r.LanguageID,
			Blog:          r.Blog,
			Written:       r.Written,
		})
	}

	return blogs, false, false, nil
}

func (cd *CoreData) forumCommentSearchNotInRestrictedTopic(r *http.Request, uid int32) ([]*db.GetCommentsByIdsForUserWithThreadInfoRow, bool, bool, error) {
	searchWords := cd.searchWordsFromRequest(r)
	var commentIDs []int32

	if len(searchWords) == 0 {
		return nil, true, false, nil
	}

	for i, word := range searchWords {
		if i == 0 {
			ids, err := cd.queries.ListCommentIDsBySearchWordFirstForListerNotInRestrictedTopic(cd.ctx, db.ListCommentIDsBySearchWordFirstForListerNotInRestrictedTopicParams{
				ListerID: uid,
				Word:     sql.NullString{String: word, Valid: true},
				UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
			})
			if err != nil {
				if !errors.Is(err, sql.ErrNoRows) {
					log.Printf("ListCommentIDsBySearchWordFirstForListerNotInRestrictedTopic Error: %s", err)
					return nil, false, false, ErrInternalServerError
				}
			}
			commentIDs = ids
		} else {
			ids, err := cd.queries.ListCommentIDsBySearchWordNextForListerNotInRestrictedTopic(cd.ctx, db.ListCommentIDsBySearchWordNextForListerNotInRestrictedTopicParams{
				ListerID: uid,
				Word:     sql.NullString{String: word, Valid: true},
				Ids:      commentIDs,
				UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
			})
			if err != nil {
				if !errors.Is(err, sql.ErrNoRows) {
					log.Printf("ListCommentIDsBySearchWordNextForListerNotInRestrictedTopic Error: %s", err)
					return nil, false, false, ErrInternalServerError
				}
			}
			commentIDs = ids
		}
		if len(commentIDs) == 0 {
			return nil, false, true, nil
		}
	}

	comments, err := cd.queries.GetCommentsByIdsForUserWithThreadInfo(cd.ctx, db.GetCommentsByIdsForUserWithThreadInfoParams{
		ViewerID: uid,
		Ids:      commentIDs,
		UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("getCommentsByIds Error: %s", err)
			return nil, false, false, ErrInternalServerError
		}
	}

	return comments, false, false, nil
}

func (cd *CoreData) forumCommentSearchInRestrictedTopic(r *http.Request, forumTopicIDs []int32, uid int32) ([]*db.GetCommentsByIdsForUserWithThreadInfoRow, bool, bool, error) {
	searchWords := cd.searchWordsFromRequest(r)
	var commentIDs []int32

	if len(searchWords) == 0 {
		return nil, true, false, nil
	}

	for i, word := range searchWords {
		if i == 0 {
			ids, err := cd.queries.ListCommentIDsBySearchWordFirstForListerInRestrictedTopic(cd.ctx, db.ListCommentIDsBySearchWordFirstForListerInRestrictedTopicParams{
				ListerID: uid,
				Word:     sql.NullString{String: word, Valid: true},
				Ftids:    forumTopicIDs,
				UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
			})
			if err != nil {
				if !errors.Is(err, sql.ErrNoRows) {
					log.Printf("ListCommentIDsBySearchWordFirstForListerInRestrictedTopic Error: %s", err)
					return nil, false, false, ErrInternalServerError
				}
			}
			commentIDs = ids
		} else {
			ids, err := cd.queries.ListCommentIDsBySearchWordNextForListerInRestrictedTopic(cd.ctx, db.ListCommentIDsBySearchWordNextForListerInRestrictedTopicParams{
				ListerID: uid,
				Word:     sql.NullString{String: word, Valid: true},
				Ids:      commentIDs,
				Ftids:    forumTopicIDs,
				UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
			})
			if err != nil {
				if !errors.Is(err, sql.ErrNoRows) {
					log.Printf("ListCommentIDsBySearchWordNextForListerInRestrictedTopic Error: %s", err)
					return nil, false, false, ErrInternalServerError
				}
			}
			commentIDs = ids
		}
		if len(commentIDs) == 0 {
			return nil, false, true, nil
		}
	}

	comments, err := cd.queries.GetCommentsByIdsForUserWithThreadInfo(cd.ctx, db.GetCommentsByIdsForUserWithThreadInfoParams{
		ViewerID: uid,
		Ids:      commentIDs,
		UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
	})
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("getCommentsByIds Error: %s", err)
			return nil, false, false, ErrInternalServerError
		}
	}

	return comments, false, false, nil
}

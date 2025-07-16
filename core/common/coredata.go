package common

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
)

// ContextValues represents context key names used across the application.
type ContextValues string

// IndexItem represents a navigation item linking to site sections.
type IndexItem struct {
	Name string
	Link string
}

const (
	defaultPageSize = 15
	pageSizeMin     = 5
	pageSizeMax     = 50
)

// NewsPost describes a news entry with access metadata.
type NewsPost struct {
	*db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow
	ShowReply    bool
	ShowEdit     bool
	Editing      bool
	Announcement *db.SiteAnnouncement
	IsAdmin      bool
}

type CoreData struct {
	IndexItems       []IndexItem
	CustomIndexItems []IndexItem
	UserID           int32
	Title            string
	AutoRefresh      bool
	FeedsEnabled     bool
	RSSFeedUrl       string
	AtomFeedUrl      string
	// AdminMode indicates whether admin-only UI elements should be displayed.
	AdminMode         bool
	NotificationCount int32
	a4codeMapper      func(tag, val string) string

	session *sessions.Session

	ctx     context.Context
	queries *db.Queries

	user              lazyValue[*db.User]
	perms             lazyValue[[]*db.GetPermissionsByUserIDRow]
	pref              lazyValue[*db.Preference]
	langs             lazyValue[[]*db.Language]
	roles             lazyValue[[]string]
	allRoles          lazyValue[[]*db.Role]
	announcement      lazyValue[*db.GetActiveAnnouncementWithNewsRow]
	forumCategories   lazyValue[[]*db.Forumcategory]
	latestNews        lazyValue[[]*NewsPost]
	latestWritings    lazyValue[[]*db.Writing]
	writingCategories lazyValue[[]*db.WritingCategory]
	publicWritings    map[string]*lazyValue[[]*db.GetPublicWritingsInCategoryForUserRow]
	bloggers          lazyValue[[]*db.BloggerCountRow]
	writers           lazyValue[[]*db.WriterCountRow]
	imageBoards       map[int32]*lazyValue[[]*db.Imageboard]
	imageBoardPosts   map[int32]*lazyValue[[]*db.GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCountForUserRow]
	forumThreads      map[int32]*lazyValue[[]*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow]
	bookmarks         lazyValue[*db.GetBookmarksForUserRow]
	newsAnnouncements map[int32]*lazyValue[*db.SiteAnnouncement]
	annMu             sync.Mutex
	forumTopics       map[int32]*lazyValue[*db.GetForumTopicByIdForUserRow]
	notifCount        lazyValue[int32]
	unreadCount       lazyValue[int64]
	writerWritings    map[int32]*lazyValue[[]*db.GetPublicWritingsByUserForViewerRow]
	linkerCategories  lazyValue[[]*db.GetLinkerCategoryLinkCountsRow]

	event *eventbus.Event
}

// SetRole preloads the current role value.
func (cd *CoreData) SetRoles(r []string) { cd.roles.set(r) }

// CoreOption configures a new CoreData instance.
type CoreOption func(*CoreData)

// WithImageURLMapper sets the a4code image mapper option.
func WithImageURLMapper(fn func(tag, val string) string) CoreOption {
	return func(cd *CoreData) { cd.a4codeMapper = fn }
}

// WithSession stores the gorilla session on the CoreData object.
func WithSession(s *sessions.Session) CoreOption {
	return func(cd *CoreData) { cd.session = s }
}

// WithEvent links an event to the CoreData object.
func WithEvent(evt *eventbus.Event) CoreOption { return func(cd *CoreData) { cd.event = evt } }

// NewCoreData creates a CoreData with context and queries applied.
func NewCoreData(ctx context.Context, q *db.Queries, opts ...CoreOption) *CoreData {
	cd := &CoreData{ctx: ctx, queries: q, newsAnnouncements: map[int32]*lazyValue[*db.SiteAnnouncement]{}}
	for _, o := range opts {
		o(cd)
	}
	return cd
}

// ImageURLMapper maps image references like "image:" or "cache:" to full URLs.
func (cd *CoreData) ImageURLMapper(tag, val string) string {
	if cd.a4codeMapper != nil {
		return cd.a4codeMapper(tag, val)
	}
	return val
}

// HasRole reports whether the current user explicitly has the named role.
func (cd *CoreData) HasRole(role string) bool {
	for _, r := range cd.Roles() {
		if r == role {
			return true
		}
	}
	if cd.queries != nil {
		for _, r := range cd.Roles() {
			if _, err := cd.queries.CheckRoleGrant(cd.ctx, db.CheckRoleGrantParams{Name: r, Action: role}); err == nil {
				return true
			}
		}
	} else {
		for _, r := range cd.Roles() {
			switch r {
			case "administrator":
				if role == "moderator" || role == "content writer" || role == "user" {
					return true
				}
			case "moderator":
				if role == "user" {
					return true
				}
			case "content writer":
				if role == "user" {
					return true
				}
			}
		}
	}
	return false
}

// ContainsItem returns true if items includes an entry with the given name.
func ContainsItem(items []IndexItem, name string) bool {
	for _, it := range items {
		if it.Name == name {
			return true
		}
	}
	return false
}

func pageSize(r *http.Request) int {
	size := defaultPageSize
	if pref, _ := r.Context().Value(ContextValues("preference")).(*db.Preference); pref != nil && pref.PageSize != 0 {
		size = int(pref.PageSize)
	}
	if size < pageSizeMin {
		size = pageSizeMin
	}
	if size > pageSizeMax {
		size = pageSizeMax
	}
	return size
}

// Role returns the user role loaded lazily.
func (cd *CoreData) Roles() []string {
	roles, _ := cd.roles.load(func() ([]string, error) {
		rs := []string{"anonymous"}
		if cd.UserID == 0 || cd.queries == nil {
			return rs, nil
		}
		rs = append(rs, "user")
		perms, err := cd.queries.GetPermissionsByUserID(cd.ctx, cd.UserID)
		if err != nil {
			return rs, nil
		}
		for _, p := range perms {
			if p.Name != "" {
				rs = append(rs, p.Name)
			}
		}
		return rs, nil
	})
	return roles
}

// Role returns the first loaded role or "anonymous" when none.
func (cd *CoreData) Role() string {
	roles := cd.Roles()
	if len(roles) == 0 {
		return "anonymous"
	}
	return roles[0]
}

// SetSession stores s on cd for later retrieval.
func (cd *CoreData) SetSession(s *sessions.Session) { cd.session = s }

// Session returns the request session if available.
func (cd *CoreData) Session() *sessions.Session { return cd.session }

// SetEvent stores evt on cd for handler access.
func (cd *CoreData) SetEvent(evt *eventbus.Event) { cd.event = evt }

// Event returns the event associated with the request, if any.
func (cd *CoreData) Event() *eventbus.Event { return cd.event }

// CurrentUser returns the logged in user's record loaded on demand.
func (cd *CoreData) CurrentUser() (*db.User, error) {
	return cd.user.load(func() (*db.User, error) {
		if cd.UserID == 0 || cd.queries == nil {
			return nil, nil
		}
		row, err := cd.queries.GetUserById(cd.ctx, cd.UserID)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return nil, err
			}
			return nil, nil
		}
		return &db.User{Idusers: row.Idusers, Username: row.Username}, nil
	})
}

// Permissions returns the user's permissions loaded on demand.
func (cd *CoreData) Permissions() ([]*db.GetPermissionsByUserIDRow, error) {
	return cd.perms.load(func() ([]*db.GetPermissionsByUserIDRow, error) {
		if cd.UserID == 0 || cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetPermissionsByUserID(cd.ctx, cd.UserID)
	})
}

// Preference returns the user's preferences loaded on demand.
func (cd *CoreData) Preference() (*db.Preference, error) {
	return cd.pref.load(func() (*db.Preference, error) {
		if cd.UserID == 0 || cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetPreferenceByUserID(cd.ctx, cd.UserID)
	})
}

// Languages returns the list of available languages loaded on demand.
func (cd *CoreData) Languages() ([]*db.Language, error) {
	return cd.langs.load(func() ([]*db.Language, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.FetchLanguages(cd.ctx)
	})
}

// AllRoles returns every defined role loaded once from the database.
func (cd *CoreData) AllRoles() ([]*db.Role, error) {
	return cd.allRoles.load(func() ([]*db.Role, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.ListRoles(cd.ctx)
	})
}

// Announcement returns the active announcement row loaded lazily.
func (cd *CoreData) Announcement() *db.GetActiveAnnouncementWithNewsRow {
	ann, _ := cd.announcement.load(func() (*db.GetActiveAnnouncementWithNewsRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		row, err := cd.queries.GetActiveAnnouncementWithNews(cd.ctx)
		if err != nil {
			return nil, err
		}
		return row, nil
	})
	return ann
}

// AnnouncementForNews fetches the latest announcement for the given news post
// only once.
func (cd *CoreData) AnnouncementForNews(id int32) (*db.SiteAnnouncement, error) {
	if cd.newsAnnouncements == nil {
		cd.newsAnnouncements = map[int32]*lazyValue[*db.SiteAnnouncement]{}
	}
	lv, ok := cd.newsAnnouncements[id]
	if !ok {
		lv = &lazyValue[*db.SiteAnnouncement]{}
		cd.newsAnnouncements[id] = lv
	}
	return lv.load(func() (*db.SiteAnnouncement, error) {
		if cd.queries == nil {
			return nil, nil
		}
		ann, err := cd.queries.GetLatestAnnouncementByNewsID(cd.ctx, id)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return ann, err
	})
}

// NewsAnnouncement returns the latest announcement for the given news post. The
// result is cached so repeated lookups for the same id hit the database only
// once.
func (cd *CoreData) NewsAnnouncement(id int32) (*db.SiteAnnouncement, error) {
	cd.annMu.Lock()
	lv, ok := cd.newsAnnouncements[id]
	if !ok {
		lv = &lazyValue[*db.SiteAnnouncement]{}
		cd.newsAnnouncements[id] = lv
	}
	cd.annMu.Unlock()

	return lv.load(func() (*db.SiteAnnouncement, error) {
		if cd.queries == nil {
			return nil, nil
		}
		ann, err := cd.queries.GetLatestAnnouncementByNewsID(cd.ctx, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}
			return nil, err
		}
		return ann, nil
	})
}

// ForumCategories loads all forum categories once.
func (cd *CoreData) ForumCategories() ([]*db.Forumcategory, error) {
	return cd.forumCategories.load(func() ([]*db.Forumcategory, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetAllForumCategories(cd.ctx)
	})
}

// ForumThreads loads the threads for a forum topic once per topic.
func (cd *CoreData) ForumThreads(topicID int32) ([]*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow, error) {
	if cd.forumThreads == nil {
		cd.forumThreads = make(map[int32]*lazyValue[[]*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow])
	}
	lv, ok := cd.forumThreads[topicID]
	if !ok {
		lv = &lazyValue[[]*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow]{}
		cd.forumThreads[topicID] = lv
	}
	return lv.load(func() ([]*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostText(cd.ctx, db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextParams{
			ViewerID:      cd.UserID,
			TopicID:       topicID,
			ViewerMatchID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
	})
}

// LatestNews returns recent news posts with permission data.
func (cd *CoreData) LatestNews(r *http.Request) ([]*NewsPost, error) {
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	replyID, _ := strconv.Atoi(r.URL.Query().Get("reply"))
	return cd.latestNews.load(func() ([]*NewsPost, error) {
		return cd.fetchLatestNews(int32(offset), 15, replyID)
	})
}

// LatestNewsList returns recent news posts without needing an HTTP request.
func (cd *CoreData) LatestNewsList(offset, limit int32) ([]*NewsPost, error) {
	return cd.fetchLatestNews(offset, limit, 0)
}

// fetchLatestNews loads news posts from the database with permission data.
func (cd *CoreData) fetchLatestNews(offset, limit int32, replyID int) ([]*NewsPost, error) {
	if cd.queries == nil {
		return nil, nil
	}
	rows, err := cd.queries.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescending(cd.ctx, db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingParams{
		ViewerID: cd.UserID,
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	var posts []*NewsPost
	for _, row := range rows {
		if !cd.HasGrant("news", "post", "see", row.Idsitenews) {
			continue
		}
		ann, err := cd.queries.GetLatestAnnouncementByNewsID(cd.ctx, row.Idsitenews)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		if !cd.HasGrant("news", "post", "see", row.Idsitenews) {
			continue
		}
		posts = append(posts, &NewsPost{
			GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow: row,
			ShowReply:    cd.UserID != 0,
			ShowEdit:     cd.HasGrant("news", "post", "edit", row.Idsitenews) && (cd.AdminMode || cd.UserID != 0),
			Editing:      replyID == int(row.Idsitenews),
			Announcement: ann,
			IsAdmin:      cd.HasRole("administrator") && cd.AdminMode,
		})
	}
	return posts, nil
}

// LatestWritings returns recent public writings with permission data.
func (cd *CoreData) LatestWritings(r *http.Request) ([]*db.Writing, error) {
	return cd.latestWritings.load(func() ([]*db.Writing, error) {
		if cd.queries == nil {
			return nil, nil
		}
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		rows, err := cd.queries.GetPublicWritings(cd.ctx, db.GetPublicWritingsParams{
			Limit:  15,
			Offset: int32(offset),
		})
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		var writings []*db.Writing
		for _, row := range rows {
			if !cd.HasGrant("writing", "article", "see", row.Idwriting) {
				continue
			}
			writings = append(writings, row)
		}
		return writings, nil
	})
}

// WritingCategories returns the visible writing categories for userID.
func (cd *CoreData) WritingCategories(userID int32) ([]*db.WritingCategory, error) {
	return cd.writingCategories.load(func() ([]*db.WritingCategory, error) {
		if cd.queries == nil {
			return nil, nil
		}
		rows, err := cd.queries.FetchCategoriesForUser(cd.ctx, db.FetchCategoriesForUserParams{
			ViewerID: cd.UserID,
			UserID:   sql.NullInt32{Int32: userID, Valid: userID != 0},
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}
			return nil, err
		}
		var cats []*db.WritingCategory
		for _, row := range rows {
			if cd.HasGrant("writing", "category", "see", row.Idwritingcategory) {
				cats = append(cats, row)
			}
		}
		return cats, nil
	})
}

// PublicWritings returns public writings in a category, cached per category and offset.
func (cd *CoreData) PublicWritings(categoryID int32, r *http.Request) ([]*db.GetPublicWritingsInCategoryForUserRow, error) {
	if cd.publicWritings == nil {
		cd.publicWritings = map[string]*lazyValue[[]*db.GetPublicWritingsInCategoryForUserRow]{}
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	key := fmt.Sprintf("%d:%d", categoryID, offset)
	lv, ok := cd.publicWritings[key]
	if !ok {
		lv = &lazyValue[[]*db.GetPublicWritingsInCategoryForUserRow]{}
		cd.publicWritings[key] = lv
	}
	return lv.load(func() ([]*db.GetPublicWritingsInCategoryForUserRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		rows, err := cd.queries.GetPublicWritingsInCategoryForUser(cd.ctx, db.GetPublicWritingsInCategoryForUserParams{
			ViewerID:          cd.UserID,
			WritingCategoryID: categoryID,
			UserID:            sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
			Limit:             15,
			Offset:            int32(offset),
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}
			return nil, err
		}
		var res []*db.GetPublicWritingsInCategoryForUserRow
		for _, row := range rows {
			if cd.HasGrant("writing", "article", "see", row.Idwriting) {
				res = append(res, row)
			}
		}
		return res, nil
	})
}

// Bloggers returns bloggers ordered by username with post counts.
func (cd *CoreData) Bloggers(r *http.Request) ([]*db.BloggerCountRow, error) {
	return cd.bloggers.load(func() ([]*db.BloggerCountRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		ps := pageSize(r)
		search := r.URL.Query().Get("search")
		if search != "" {
			return cd.queries.SearchBloggers(cd.ctx, db.SearchBloggersParams{
				ViewerID: cd.UserID,
				Query:    search,
				Limit:    int32(ps + 1),
				Offset:   int32(offset),
			})
		}
		return cd.queries.ListBloggers(cd.ctx, db.ListBloggersParams{
			ViewerID: cd.UserID,
			Limit:    int32(ps + 1),
			Offset:   int32(offset),
		})
	})
}

// Writers returns writers ordered by username with article counts.
func (cd *CoreData) Writers(r *http.Request) ([]*db.WriterCountRow, error) {
	return cd.writers.load(func() ([]*db.WriterCountRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		ps := pageSize(r)
		search := r.URL.Query().Get("search")
		if search != "" {
			return cd.queries.SearchWriters(cd.ctx, db.SearchWritersParams{
				ViewerID: cd.UserID,
				Query:    search,
				Limit:    int32(ps + 1),
				Offset:   int32(offset),
			})
		}
		return cd.queries.ListWriters(cd.ctx, db.ListWritersParams{
			ViewerID: cd.UserID,
			Limit:    int32(ps + 1),
			Offset:   int32(offset),
		})
	})
}

// ForumTopicByID loads a forum topic once per ID using caching.
func (cd *CoreData) ForumTopicByID(id int32) (*db.GetForumTopicByIdForUserRow, error) {
	if cd.queries == nil {
		return nil, nil
	}
	if cd.forumTopics == nil {
		cd.forumTopics = make(map[int32]*lazyValue[*db.GetForumTopicByIdForUserRow])
	}
	lv, ok := cd.forumTopics[id]
	if !ok {
		lv = &lazyValue[*db.GetForumTopicByIdForUserRow]{}
		cd.forumTopics[id] = lv
	}
	return lv.load(func() (*db.GetForumTopicByIdForUserRow, error) {
		return cd.queries.GetForumTopicByIdForUser(cd.ctx, db.GetForumTopicByIdForUserParams{
			ViewerID:      cd.UserID,
			Idforumtopic:  id,
			ViewerMatchID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
	})
}

// WriterWritings returns public writings for the specified author respecting cd's permissions.
func (cd *CoreData) WriterWritings(userID int32, r *http.Request) ([]*db.GetPublicWritingsByUserForViewerRow, error) {
	if cd.writerWritings == nil {
		cd.writerWritings = map[int32]*lazyValue[[]*db.GetPublicWritingsByUserForViewerRow]{}
	}
	lv, ok := cd.writerWritings[userID]
	if !ok {
		lv = &lazyValue[[]*db.GetPublicWritingsByUserForViewerRow]{}
		cd.writerWritings[userID] = lv
	}
	return lv.load(func() ([]*db.GetPublicWritingsByUserForViewerRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		rows, err := cd.queries.GetPublicWritingsByUserForViewer(cd.ctx, db.GetPublicWritingsByUserForViewerParams{
			ViewerID: cd.UserID,
			AuthorID: userID,
			UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
			Limit:    15,
			Offset:   int32(offset),
		})
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		var list []*db.GetPublicWritingsByUserForViewerRow
		for _, row := range rows {
			if !cd.HasGrant("writing", "article", "see", row.Idwriting) {
				continue
			}
			list = append(list, row)
		}
		return list, nil
	})
}

// Bookmarks returns the user's bookmark list loaded lazily.
func (cd *CoreData) Bookmarks() (*db.GetBookmarksForUserRow, error) {
	return cd.bookmarks.load(func() (*db.GetBookmarksForUserRow, error) {
		if cd.UserID == 0 || cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetBookmarksForUser(cd.ctx, cd.UserID)
	})
}

// CanEditAny reports whether cd is in admin mode with administrator role.
func (cd *CoreData) CanEditAny() bool {
	return cd.HasRole("administrator") && cd.AdminMode
}

// ImageBoards retrieves sub-boards under parentID lazily.
func (cd *CoreData) ImageBoards(parentID int32) ([]*db.Imageboard, error) {
	if cd.queries == nil {
		return nil, nil
	}
	if cd.imageBoards == nil {
		cd.imageBoards = make(map[int32]*lazyValue[[]*db.Imageboard])
	}
	lv, ok := cd.imageBoards[parentID]
	if !ok {
		lv = &lazyValue[[]*db.Imageboard]{}
		cd.imageBoards[parentID] = lv
	}
	return lv.load(func() ([]*db.Imageboard, error) {
		return cd.queries.GetAllBoardsByParentBoardIdForUser(cd.ctx, db.GetAllBoardsByParentBoardIdForUserParams{
			ViewerID:     cd.UserID,
			ParentID:     parentID,
			ViewerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
	})
}

// ImageBoardPosts retrieves approved posts for the board lazily.
func (cd *CoreData) ImageBoardPosts(boardID int32) ([]*db.GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCountForUserRow, error) {
	if cd.queries == nil {
		return nil, nil
	}
	if cd.imageBoardPosts == nil {
		cd.imageBoardPosts = make(map[int32]*lazyValue[[]*db.GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCountForUserRow])
	}
	lv, ok := cd.imageBoardPosts[boardID]
	if !ok {
		lv = &lazyValue[[]*db.GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCountForUserRow]{}
		cd.imageBoardPosts[boardID] = lv
	}
	return lv.load(func() ([]*db.GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCountForUserRow, error) {
		return cd.queries.GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCountForUser(cd.ctx, db.GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCountForUserParams{
			ViewerID:     cd.UserID,
			BoardID:      boardID,
			ViewerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
	})
}

// UnreadNotificationCount returns the number of unread notifications for the
// current user. The value is fetched lazily on the first call and cached for
// subsequent calls.
func (cd *CoreData) UnreadNotificationCount() int64 {
	count, _ := cd.unreadCount.load(func() (int64, error) {
		if cd.queries == nil || cd.UserID == 0 {
			return 0, nil
		}
		return cd.queries.CountUnreadNotifications(cd.ctx, cd.UserID)
	})
	return count
}

// LinkerCategoryCounts lazily loads linker category statistics.
func (cd *CoreData) LinkerCategoryCounts() ([]*db.GetLinkerCategoryLinkCountsRow, error) {
	return cd.linkerCategories.load(func() ([]*db.GetLinkerCategoryLinkCountsRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		rows, err := cd.queries.GetLinkerCategoryLinkCounts(cd.ctx)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return rows, nil
	})
}

package common

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/tasks"
)

// IndexItem represents a navigation item linking to site sections.
type IndexItem struct {
	Name string
	Link string
}

// MailProvider defines the interface required by CoreData for sending emails.
type MailProvider interface {
	Send(ctx context.Context, to mail.Address, rawEmailMessage []byte) error
}

// No package-level pagination constants as runtime config provides these values.

// NewsPost describes a news entry with access metadata.
type NewsPost struct {
	*db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow
	ShowReply    bool
	ShowEdit     bool
	Editing      bool
	Announcement *db.SiteAnnouncement
}

type CoreData struct {
	IndexItems       []IndexItem
	CustomIndexItems []IndexItem
	UserID           int32
	Title            string
	AutoRefresh      string
	FeedsEnabled     bool
	RSSFeedUrl       string
	AtomFeedUrl      string
	// AdminMode indicates whether admin-only UI elements should be displayed.
	AdminMode         bool
	NotificationCount int32
	a4codeMapper      func(tag, val string) string

	session *sessions.Session

	ctx           context.Context
	queries       *db.Queries
	emailProvider lazyValue[MailProvider]

	allRoles                 lazyValue[[]*db.Role]
	announcement             lazyValue[*db.GetActiveAnnouncementWithNewsRow]
	annMu                    sync.Mutex
	bloggers                 lazyValue[[]*db.BloggerCountRow]
	bookmarks                lazyValue[*db.GetBookmarksForUserRow]
	event                    *eventbus.TaskEvent
	forumCategories          lazyValue[[]*db.Forumcategory]
	forumThreads             map[int32]*lazyValue[[]*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow]
	forumTopics              map[int32]*lazyValue[*db.GetForumTopicByIdForUserRow]
	imageBoardPosts          map[int32]*lazyValue[[]*db.GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCountForUserRow]
	imageBoards              lazyValue[[]*db.Imageboard]
	languagesAll             lazyValue[[]*db.Language]
	langs                    lazyValue[[]*db.Language]
	latestNews               lazyValue[[]*NewsPost]
	latestWritings           lazyValue[[]*db.Writing]
	linkerCategories         lazyValue[[]*db.GetLinkerCategoryLinkCountsRow]
	newsAnnouncements        map[int32]*lazyValue[*db.SiteAnnouncement]
	notifCount               lazyValue[int32]
	perms                    lazyValue[[]*db.GetPermissionsByUserIDRow]
	pref                     lazyValue[*db.Preference]
	preferredLanguageID      lazyValue[int32]
	publicWritings           map[string]*lazyValue[[]*db.GetPublicWritingsInCategoryForUserRow]
	subImageBoards           map[int32]*lazyValue[[]*db.Imageboard]
	unreadCount              lazyValue[int64]
	subscriptions            lazyValue[map[string]bool]
	user                     lazyValue[*db.User]
	userRoles                lazyValue[[]string]
	visibleWritingCategories lazyValue[[]*db.WritingCategory]
	writerWritings           map[int32]*lazyValue[[]*db.GetPublicWritingsByUserForViewerRow]
	writers                  lazyValue[[]*db.WriterCountRow]
	writingCategories        lazyValue[[]*db.WritingCategory]

	absoluteURLBase lazyValue[string]
	// marks records which template sections have been rendered to avoid
	// duplicate output when re-rendering after an error.
	marks map[string]struct{}
}

// SetRoles preloads the current user roles.
func (cd *CoreData) SetRoles(r []string) { cd.userRoles.set(r) }

// Marked returns true the first time it is called with key. Subsequent
// calls return false. It is used to avoid re-rendering template sections
// when streaming pages after an error.
func (cd *CoreData) Marked(key string) bool {
	if cd.marks == nil {
		cd.marks = map[string]struct{}{}
	}
	_, marked := cd.marks[key]
	cd.marks[key] = struct{}{}
	return !marked
}

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
func WithEvent(evt *eventbus.TaskEvent) CoreOption { return func(cd *CoreData) { cd.event = evt } }

// WithAbsoluteURLBase sets the base URL used to build absolute links.
func WithAbsoluteURLBase(base string) CoreOption {
	return func(cd *CoreData) { cd.absoluteURLBase.set(strings.TrimRight(base, "/")) }
}

// WithPreference preloads the user preference object.
func WithPreference(p *db.Preference) CoreOption {
	return func(cd *CoreData) { cd.pref.set(p) }
}

// NewCoreData creates a CoreData with context and queries applied.
func NewCoreData(ctx context.Context, q *db.Queries, opts ...CoreOption) *CoreData {
	cd := &CoreData{ctx: ctx, queries: q, newsAnnouncements: map[int32]*lazyValue[*db.SiteAnnouncement]{}}
	for _, o := range opts {
		o(cd)
	}
	return cd
}

// Queries returns the db.Queries instance associated with this CoreData.
func (cd *CoreData) Queries() *db.Queries { return cd.queries }

// ImageURLMapper maps image references like "image:" or "cache:" to full URLs.
func (cd *CoreData) ImageURLMapper(tag, val string) string {
	if cd.a4codeMapper != nil {
		return cd.a4codeMapper(tag, val)
	}
	return val
}

// EmailProvider lazily returns the configured email provider.
// WithEmailProvider sets the email provider used by CoreData.
func WithEmailProvider(p MailProvider) CoreOption {
	return func(cd *CoreData) { cd.emailProvider.set(p) }
}

// EmailProvider returns the configured email provider.
func (cd *CoreData) EmailProvider() MailProvider {
	p, err := cd.emailProvider.load(func() (MailProvider, error) { return nil, nil })
	if err != nil {
		log.Printf("load email provider: %v", err)
	}
	return p
}

// HasRole reports whether the current user explicitly has the named role.
func (cd *CoreData) HasRole(role string) bool {
	for _, r := range cd.UserRoles() {
		if r == role {
			return true
		}
	}
	if cd.queries != nil {
		for _, r := range cd.UserRoles() {
			if _, err := cd.queries.CheckRoleGrant(cd.ctx, db.CheckRoleGrantParams{Name: r, Action: role}); err == nil {
				return true
			}
		}
	} else {
		for _, r := range cd.UserRoles() {
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
	cd, _ := r.Context().Value(consts.KeyCoreData).(*CoreData)
	if cd == nil {
		size := config.AppRuntimeConfig.PageSizeDefault
		if size < config.AppRuntimeConfig.PageSizeMin {
			size = config.AppRuntimeConfig.PageSizeMin
		}
		if size > config.AppRuntimeConfig.PageSizeMax {
			size = config.AppRuntimeConfig.PageSizeMax
		}
		return size
	}
	return cd.PageSize()
}

// UserRoles returns the user roles loaded lazily.
func (cd *CoreData) UserRoles() []string {
	roles, err := cd.userRoles.load(func() ([]string, error) {
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
	if err != nil {
		log.Printf("load user roles: %v", err)
	}
	return roles
}

// Role returns the first loaded role or "anonymous" when none.
func (cd *CoreData) Role() string {
	roles := cd.UserRoles()
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
func (cd *CoreData) SetEvent(evt *eventbus.TaskEvent) { cd.event = evt }

// SetEventTask records the task associated with the current request event.
func (cd *CoreData) SetEventTask(t tasks.Task) {
	if cd.event != nil {
		cd.event.Task = t
	}
}

// AbsoluteURL returns an absolute URL by combining the configured hostname or
// the request host with path. The base value is cached per request.
func (cd *CoreData) AbsoluteURL(path string) string {
	base, err := cd.absoluteURLBase.load(func() (string, error) { return "", nil })
	if err != nil {
		log.Printf("load absolute URL base: %v", err)
	}
	return base + path
}

// Event returns the event associated with the request, if any.
func (cd *CoreData) Event() *eventbus.TaskEvent { return cd.event }

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

// CurrentUserLoaded returns the cached current user without triggering a database lookup.
func (cd *CoreData) CurrentUserLoaded() *db.User {
	u, ok := cd.user.peek()
	if !ok {
		return nil
	}
	return u
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

// PageSize returns the preferred page size within configured limits.
func (cd *CoreData) PageSize() int {
	size := config.AppRuntimeConfig.PageSizeDefault
	if pref, err := cd.Preference(); err == nil && pref != nil && pref.PageSize != 0 {
		size = int(pref.PageSize)
	}
	if size < config.AppRuntimeConfig.PageSizeMin {
		size = config.AppRuntimeConfig.PageSizeMin
	}
	if size > config.AppRuntimeConfig.PageSizeMax {
		size = config.AppRuntimeConfig.PageSizeMax
	}
	return size
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

// AllLanguages returns all languages cached once.
func (cd *CoreData) AllLanguages() ([]*db.Language, error) {
	return cd.languagesAll.load(func() ([]*db.Language, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.FetchLanguages(cd.ctx)
	})
}

// PreferredLanguageID returns the user's preferred language ID if set,
// otherwise it resolves the site's default language name to an ID.
func (cd *CoreData) PreferredLanguageID(siteDefault string) int32 {
	id, err := cd.preferredLanguageID.load(func() (int32, error) {
		if pref, err := cd.Preference(); err == nil && pref != nil {
			if pref.LanguageIdlanguage != 0 {
				return pref.LanguageIdlanguage, nil
			}
		}
		if cd.queries == nil || siteDefault == "" {
			return 0, nil
		}
		langID, err := cd.queries.GetLanguageIDByName(cd.ctx, sql.NullString{String: siteDefault, Valid: true})
		if err != nil {
			return 0, nil
		}
		return langID, nil
	})
	if err != nil {
		log.Printf("load preferred language id: %v", err)
	}
	return id
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
	ann, err := cd.announcement.load(func() (*db.GetActiveAnnouncementWithNewsRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		row, err := cd.queries.GetActiveAnnouncementWithNews(cd.ctx, db.GetActiveAnnouncementWithNewsParams{
			ViewerID: cd.UserID,
			UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
		if err != nil {
			return nil, err
		}
		return row, nil
	})
	if err != nil {
		log.Printf("load announcement: %v", err)
	}
	return ann
}

// AnnouncementLoaded returns the cached active announcement without querying the database.
func (cd *CoreData) AnnouncementLoaded() *db.GetActiveAnnouncementWithNewsRow {
	ann, ok := cd.announcement.peek()
	if !ok {
		return nil
	}
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
		})
	}
	return posts, nil
}

// LatestWritings returns recent public writings with permission data.
type LatestWritingsOption func(*db.GetPublicWritingsParams)

// WithWritingsOffset sets the query offset.
func WithWritingsOffset(o int32) LatestWritingsOption {
	return func(p *db.GetPublicWritingsParams) { p.Offset = o }
}

// WithWritingsLimit sets the query limit.
func WithWritingsLimit(l int32) LatestWritingsOption {
	return func(p *db.GetPublicWritingsParams) { p.Limit = l }
}

func (cd *CoreData) LatestWritings(opts ...LatestWritingsOption) ([]*db.Writing, error) {
	return cd.latestWritings.load(func() ([]*db.Writing, error) {
		if cd.queries == nil {
			return nil, nil
		}
		params := db.GetPublicWritingsParams{Limit: 15}
		for _, o := range opts {
			o(&params)
		}
		rows, err := cd.queries.GetPublicWritings(cd.ctx, params)
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
func (cd *CoreData) VisibleWritingCategories(userID int32) ([]*db.WritingCategory, error) {
	return cd.visibleWritingCategories.load(func() ([]*db.WritingCategory, error) {
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

// WritingCategories returns all writing categories cached once.
func (cd *CoreData) WritingCategories() ([]*db.WritingCategory, error) {
	return cd.writingCategories.load(func() ([]*db.WritingCategory, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.FetchAllCategories(cd.ctx)
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
		ps := cd.PageSize()
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
		ps := cd.PageSize()
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
func (cd *CoreData) SubImageBoards(parentID int32) ([]*db.Imageboard, error) {
	if cd.queries == nil {
		return nil, nil
	}
	if cd.subImageBoards == nil {
		cd.subImageBoards = make(map[int32]*lazyValue[[]*db.Imageboard])
	}
	lv, ok := cd.subImageBoards[parentID]
	if !ok {
		lv = &lazyValue[[]*db.Imageboard]{}
		cd.subImageBoards[parentID] = lv
	}
	return lv.load(func() ([]*db.Imageboard, error) {
		return cd.queries.GetAllBoardsByParentBoardIdForUser(cd.ctx, db.GetAllBoardsByParentBoardIdForUserParams{
			ViewerID:     cd.UserID,
			ParentID:     parentID,
			ViewerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
	})
}

// ImageBoards returns all image boards cached once.
func (cd *CoreData) ImageBoards() ([]*db.Imageboard, error) {
	return cd.imageBoards.load(func() ([]*db.Imageboard, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetAllImageBoards(cd.ctx)
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
	count, err := cd.unreadCount.load(func() (int64, error) {
		if cd.queries == nil || cd.UserID == 0 {
			return 0, nil
		}
		return cd.queries.CountUnreadNotifications(cd.ctx, cd.UserID)
	})
	if err != nil {
		log.Printf("load unread notification count: %v", err)
	}
	return count
}

// subscriptionMap loads the current user's subscriptions once.
func (cd *CoreData) subscriptionMap() (map[string]bool, error) {
	return cd.subscriptions.load(func() (map[string]bool, error) {
		if cd.queries == nil || cd.UserID == 0 {
			return map[string]bool{}, nil
		}
		rows, err := cd.queries.ListSubscriptionsByUser(cd.ctx, cd.UserID)
		if err != nil {
			return nil, err
		}
		m := make(map[string]bool)
		for _, row := range rows {
			if row.Method == "internal" {
				m[row.Pattern] = true
			}
		}
		return m, nil
	})
}

// Subscribed reports whether the user has a subscription matching pattern.
func (cd *CoreData) Subscribed(pattern string) bool {
	m, _ := cd.subscriptionMap()
	return m[pattern]
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

func (cd *CoreData) HasAdminRole() bool {
	return cd.HasRole("administrator") && cd.AdminMode
}

func (cd *CoreData) HasContentWriterRole() bool {
	return cd.HasRole("content writer") || cd.HasAdminRole()
}

package common

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"

	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dbdrivers"
	"github.com/arran4/goa4web/internal/eventbus"
	imagesign "github.com/arran4/goa4web/internal/images"
	"github.com/arran4/goa4web/internal/lazy"
	linksign "github.com/arran4/goa4web/internal/linksign"
	"github.com/arran4/goa4web/internal/tasks"
)

// Ensure SessionProxy implements SessionManager.
var _ SessionManager = (*db.SessionProxy)(nil)

// IndexItem represents a navigation item linking to site sections.
type IndexItem struct {
	Name string
	Link string
}

// AdminSection groups admin navigation links under a section heading.
type AdminSection struct {
	Name  string
	Links []IndexItem
}

// PageLink represents a numbered pagination link.
type PageLink struct {
	Num    int
	Link   string
	Active bool
}

// SessionManager defines optional hooks for storing and removing session
// information. Implementations may persist session metadata in a database or
// other storage while exposing a storage-agnostic API to CoreData.
type SessionManager interface {
	InsertSession(ctx context.Context, sessionID string, userID int32) error
	DeleteSessionByID(ctx context.Context, sessionID string) error
}

// MailProvider defines the interface required by CoreData for sending emails.
type MailProvider interface {
	Send(ctx context.Context, to mail.Address, rawEmailMessage []byte) error
}

// NavigationProvider exposes index and admin navigation links.
type NavigationProvider interface {
	IndexItems() []IndexItem
	AdminLinks() []IndexItem
	AdminSections() []AdminSection
}

// ErrEmailAlreadyAssociated is returned when a user already has an email set.
var ErrEmailAlreadyAssociated = errors.New("email already associated")

// AssociateEmailParams carries the data required to request an email association.
type AssociateEmailParams struct {
	Username string
	Email    string
	Reason   string
}

// No package-level pagination constants as runtime config provides these values.

// ErrPasswordResetRecentlyRequested indicates a password reset was requested too recently.
var ErrPasswordResetRecentlyRequested = errors.New("reset recently requested")

type CoreData struct {
	a4codeMapper func(tag, val string) string
	// AdminMode indicates whether admin-only UI elements should be displayed.
	AdminMode         bool
	AtomFeedURL       string
	AutoRefresh       string
	Config            *config.RuntimeConfig
	CustomIndexItems  []IndexItem
	FeedsEnabled      bool
	ImageSigner       *imagesign.Signer
	IndexItems        []IndexItem
	LinkSigner        *linksign.Signer
	mapMu             sync.Mutex
	Nav               NavigationProvider
	NextLink          string
	NotificationCount int32
	PageLinks         []PageLink
	PageTitle         string
	PrevLink          string
	RSSFeedURL        string
	StartLink         string
	TasksReg          *tasks.Registry
	SiteTitle         string
	UserID            int32

	session      *sessions.Session
	sessionProxy SessionManager

	ctx           context.Context
	customQueries db.CustomQueries
	emailProvider lazy.Value[MailProvider]
	queries       db.Querier

	// Keep this sorted
	adminLatestNews                  lazy.Value[[]*db.AdminListNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow]
	adminLinkerItemRows              map[int32]*lazy.Value[*db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow]
	adminRequest                     map[int32]*lazy.Value[*db.AdminRequestQueue]
	adminRequestComments             map[int32]*lazy.Value[[]*db.AdminRequestComment]
	adminRequests                    map[string]*lazy.Value[[]*db.AdminRequestQueue]
	adminUserBookmarkSize            map[int32]*lazy.Value[int]
	adminUserComments                map[int32]*lazy.Value[[]*db.AdminUserComment]
	adminUserEmails                  map[int32]*lazy.Value[[]*db.UserEmail]
	adminUserGrants                  map[int32]*lazy.Value[[]*db.Grant]
	adminUserRoles                   map[int32]*lazy.Value[[]*db.GetPermissionsByUserIDRow]
	adminUserStats                   map[int32]*lazy.Value[*db.AdminUserPostCountsByIDRow]
	allAnsweredFAQ                  lazy.Value[[]*CategoryFAQs]
	allRoles                         lazy.Value[[]*db.Role]
	annMu                            sync.Mutex
	announcement                     lazy.Value[*db.GetActiveAnnouncementWithNewsForListerRow]
	blogEntries                      map[int32]*lazy.Value[*db.GetBlogEntryForListerByIDRow]
	bloggers                         lazy.Value[[]*db.ListBloggersForListerRow]
	blogListOffset                   int
	blogListRows                     lazy.Value[[]*db.ListBlogEntriesForListerRow]
	blogListByAuthorRows             lazy.Value[[]*db.ListBlogEntriesByAuthorForListerRow]
	blogListUID                      int32
	bookmarks                        lazy.Value[*db.GetBookmarksForUserRow]
	currentBlogID                    int32
	currentBoardID                   int32
	currentCommentID                 int32
	currentImagePostID               int32
	currentLinkID                    int32
	currentExternalLinkID            int32
	currentOffset                    int
	currentNewsPostID                int32
	currentProfileUserID             int32
	currentRequestID                 int32
	currentRoleID                    int32
	currentSection                   string
	currentNotificationTemplateError string
	currentNotificationTemplateName  string
	currentError                     string
	currentThreadID                  int32
	currentTopicID                   int32
	currentCategoryID                int32
	currentWritingID                 int32
	event                            *eventbus.TaskEvent
	externalLinks                    map[int32]*lazy.Value[*db.ExternalLink]
	faqCategories                    lazy.Value[[]*db.FaqCategory]
	forumCategories                  lazy.Value[[]*db.Forumcategory]
	forumComments                    map[int32]*lazy.Value[*db.GetCommentByIdForUserRow]
	forumThreadComments              map[int32]*lazy.Value[[]*db.GetCommentsByThreadIdForUserRow]
	forumThreadRows                  map[int32]*lazy.Value[*db.GetThreadLastPosterAndPermsRow]
	forumThreads                     map[int32]*lazy.Value[[]*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow]
	forumTopicLists                  map[int32]*lazy.Value[[]*db.Forumtopic]
	forumTopics                      map[int32]*lazy.Value[*db.GetForumTopicByIdForUserRow]
	imageBoardPosts                  map[int32]*lazy.Value[[]*db.ListImagePostsByBoardForListerRow]
	imageBoards                      lazy.Value[[]*db.Imageboard]
	imagePostRows                    map[int32]*lazy.Value[*db.GetImagePostByIDForListerRow]
	langs                            lazy.Value[[]*db.Language]
	latestNews                       lazy.Value[[]*db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow]
	latestWritings                   lazy.Value[[]*db.Writing]
	linkerCategories                 lazy.Value[[]*db.GetLinkerCategoryLinkCountsRow]
	linkerCategoryLinks              map[int32]*lazy.Value[[]*db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingRow]
	linkerCategoryRows               map[int32]*lazy.Value[*db.LinkerCategory]
	linkerCatsAll                    lazy.Value[[]*db.LinkerCategory]
	linkerCatsForUser                lazy.Value[[]*db.LinkerCategory]
	newsAnnouncements                map[int32]*lazy.Value[*db.SiteAnnouncement]
	newsPosts                        map[int32]*lazy.Value[*db.GetForumThreadIdByNewsPostIdRow]
	notifCount                       lazy.Value[int32]
	notifications                    map[string]*lazy.Value[[]*db.Notification]
	perms                            lazy.Value[[]*db.GetPermissionsByUserIDRow]
	pref                             lazy.Value[*db.Preference]
	preferredLanguageID              lazy.Value[int32]
	privateForumTopics               lazy.Value[[]*PrivateTopic]
	publicWritings                   map[string]*lazy.Value[[]*db.ListPublicWritingsInCategoryForListerRow]
	roleRows                         map[int32]*lazy.Value[*db.Role]
	selectedThreadCanReply           lazy.Value[bool]
	subImageBoards                   map[int32]*lazy.Value[[]*db.Imageboard]
	subscriptionRows                 lazy.Value[[]*db.ListSubscriptionsByUserRow]
	subscriptions                    lazy.Value[map[string]bool]
	notificationTemplateOverrides    map[string]*lazy.Value[string]
	unreadCount                      lazy.Value[int64]
	user                             lazy.Value[*db.User]
	userRoles                        lazy.Value[[]string]
	users                            map[int32]*lazy.Value[*db.SystemGetUserByIDRow]
	userSubscriptions                lazy.Value[[]*db.ListSubscriptionsByUserRow]
	visibleWritingCategories         lazy.Value[[]*db.WritingCategory]
	writers                          lazy.Value[[]*db.ListWritersForListerRow]
	writerWritings                   map[int32]*lazy.Value[[]*db.ListPublicWritingsByUserForListerRow]
	writingCategories                lazy.Value[[]*db.WritingCategory]
	writingRows                      map[int32]*lazy.Value[*db.GetWritingForListerByIDRow]

  // marks records which template sections have been rendered to avoid
	// duplicate output when re-rendering after an error.
	marks map[string]struct{}
}

// AbsoluteURL returns an absolute URL by combining the configured hostname or
// the request host with path. The base value is cached per request.
func (cd *CoreData) AbsoluteURL(path string) string {
	base, err := cd.absoluteURLBase.Load(func() (string, error) { return "", nil })
	if err != nil {
		log.Printf("load absolute URL base: %v", err)
	}
	return base + path
}

// AdminForumTopics returns all forum topics without category filtering.
func (cd *CoreData) AdminForumTopics() ([]*db.Forumtopic, error) {
	return cd.ForumTopics(0)
}

// AdminLatestNews returns recent news posts for administrators using cd's current offset and page size.
func (cd *CoreData) AdminLatestNews() ([]*db.AdminListNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow, error) {
	ps := cd.PageSize()
	return cd.adminLatestNews.Load(func() ([]*db.AdminListNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow, error) {
		return cd.AdminLatestNewsList(int32(cd.currentOffset), int32(ps))
	})
}

// AdminLatestNewsList returns recent news posts for administrators without permission checks.
func (cd *CoreData) AdminLatestNewsList(offset, limit int32) ([]*db.AdminListNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow, error) {
	if cd.queries == nil {
		return nil, nil
	}
	rows, err := cd.queries.AdminListNewsPostsWithWriterUsernameAndThreadCommentCountDescending(cd.ctx, db.AdminListNewsPostsWithWriterUsernameAndThreadCommentCountDescendingParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return rows, nil
}

// AdminLoginAttempts returns recent login attempts for administrators.
func (cd *CoreData) AdminLoginAttempts() ([]*db.LoginAttempt, error) {
	if cd.queries == nil {
		return nil, nil
	}
	rows, err := cd.queries.AdminListLoginAttempts(cd.ctx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return rows, nil
}

// AdminSessions returns active sessions for administrators.
func (cd *CoreData) AdminSessions() ([]*db.AdminListSessionsRow, error) {
	if cd.queries == nil {
		return nil, nil
	}
	rows, err := cd.queries.AdminListSessions(cd.ctx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return rows, nil
}

// AdminLinkerItemByID returns a single linker item lazily loading it once per ID.
func (cd *CoreData) AdminLinkerItemByID(id int32, ops ...lazy.Option[*db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow]) (*db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow, error) {
	fetch := func(i int32) (*db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		row, err := cd.queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending(cd.ctx, i)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return row, nil
	}
	return lazy.Map(&cd.adminLinkerItemRows, &cd.mapMu, id, fetch, ops...)
}

func (cd *CoreData) adminRequestList(kind string) ([]*db.AdminRequestQueue, error) {
	if cd.adminRequests == nil {
		cd.adminRequests = map[string]*lazy.Value[[]*db.AdminRequestQueue]{}
	}
	lv, ok := cd.adminRequests[kind]
	if !ok {
		lv = &lazy.Value[[]*db.AdminRequestQueue]{}
		cd.adminRequests[kind] = lv
	}
	return lv.Load(func() ([]*db.AdminRequestQueue, error) {
		if cd.queries == nil {
			return nil, nil
		}
		switch kind {
		case "pending":
			return cd.queries.AdminListPendingRequests(cd.ctx)
		case "archived":
			return cd.queries.AdminListArchivedRequests(cd.ctx)
		default:
			return nil, nil
		}
	})
}

// AllRoles returns every defined role loaded once from the database.
func (cd *CoreData) AllRoles() ([]*db.Role, error) {
	return cd.allRoles.Load(func() ([]*db.Role, error) {
		var roles []*db.Role
		if cd.queries != nil {
			var err error
			roles, err = cd.queries.AdminListRoles(cd.ctx)
			if err != nil {
				return nil, err
			}
		}
		for _, r := range roles {
			if r.Name == "anyone" {
				return roles, nil
			}
		}
		anyone := &db.Role{Name: "anyone"}
		roles = append([]*db.Role{anyone}, roles...)
		return roles, nil
	})
}

// RoleByID returns a role lazily loading it once per ID.
func (cd *CoreData) RoleByID(id int32, ops ...lazy.Option[*db.Role]) (*db.Role, error) {
	fetch := func(i int32) (*db.Role, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.AdminGetRoleByID(cd.ctx, i)
	}
	return lazy.Map(&cd.roleRows, &cd.mapMu, id, fetch, ops...)
}

// SelectedRole returns the role referenced by the current request.
func (cd *CoreData) SelectedRole(ops ...lazy.Option[*db.Role]) (*db.Role, error) {
	if cd.currentRoleID == 0 {
		return nil, nil
	}
	return cd.RoleByID(cd.currentRoleID, ops...)
}

// Announcement returns the active announcement row loaded lazily.
func (cd *CoreData) Announcement() *db.GetActiveAnnouncementWithNewsForListerRow {
	ann, err := cd.announcement.Load(func() (*db.GetActiveAnnouncementWithNewsForListerRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		row, err := cd.queries.GetActiveAnnouncementWithNewsForLister(cd.ctx, db.GetActiveAnnouncementWithNewsForListerParams{
			ListerID: cd.UserID,
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
func (cd *CoreData) AnnouncementLoaded() *db.GetActiveAnnouncementWithNewsForListerRow {
	ann, ok := cd.announcement.Peek()
	if !ok {
		return nil
	}
	return ann
}

// ArchivedRequests returns archived admin requests loaded on demand.
func (cd *CoreData) ArchivedRequests() []*db.AdminRequestQueue {
	rows, err := cd.adminRequestList("archived")
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("load archived requests: %v", err)
		return nil
	}
	return rows
}

// BlogEntryByID returns a blog entry lazily loading it once per ID.
func (cd *CoreData) BlogEntryByID(id int32, ops ...lazy.Option[*db.GetBlogEntryForListerByIDRow]) (*db.GetBlogEntryForListerByIDRow, error) {
	fetch := func(i int32) (*db.GetBlogEntryForListerByIDRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetBlogEntryForListerByID(cd.ctx, db.GetBlogEntryForListerByIDParams{
			ListerID: cd.UserID,
			ID:       i,
			UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
	}
	return lazy.Map(&cd.blogEntries, &cd.mapMu, id, fetch, ops...)
}

// Bloggers returns bloggers ordered by username with post counts.
func (cd *CoreData) Bloggers(r *http.Request) ([]*db.ListBloggersForListerRow, error) {
	return cd.bloggers.Load(func() ([]*db.ListBloggersForListerRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		ps := cd.PageSize()
		search := r.URL.Query().Get("search")
		if search != "" {
			like := "%" + search + "%"
			rows, err := cd.queries.ListBloggersSearchForLister(cd.ctx, db.ListBloggersSearchForListerParams{
				ListerID: cd.UserID,
				Query:    like,
				UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
				Limit:    int32(ps + 1),
				Offset:   int32(offset),
			})
			if err != nil {
				return nil, err
			}
			items := make([]*db.ListBloggersForListerRow, 0, len(rows))
			for _, r := range rows {
				items = append(items, &db.ListBloggersForListerRow{Username: r.Username, Count: r.Count})
			}
			return items, nil
		}
		return cd.queries.ListBloggersForLister(cd.ctx, db.ListBloggersForListerParams{
			ListerID: cd.UserID,
			UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
			Limit:    int32(ps + 1),
			Offset:   int32(offset),
		})
	})
}

// BlogList returns blog entries visible to the current user.
func (cd *CoreData) BlogList() ([]*db.ListBlogEntriesForListerRow, error) {
	return cd.blogListRows.Load(func() ([]*db.ListBlogEntriesForListerRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		rows, err := cd.queries.ListBlogEntriesForLister(cd.ctx, db.ListBlogEntriesForListerParams{
			ListerID: cd.UserID,
			UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
			Limit:    int32(cd.PageSize()),
			Offset:   int32(cd.blogListOffset),
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}
			return nil, err
		}
		var list []*db.ListBlogEntriesForListerRow
		for _, row := range rows {
			if !cd.HasGrant("blogs", "entry", "see", row.Idblogs) {
				continue
			}
			list = append(list, row)
		}
		return list, nil
	})
}

// BlogListForSelectedAuthor returns blog entries for the selected author.
func (cd *CoreData) BlogListForSelectedAuthor() ([]*db.ListBlogEntriesByAuthorForListerRow, error) {
	return cd.blogListByAuthorRows.Load(func() ([]*db.ListBlogEntriesByAuthorForListerRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		rows, err := cd.queries.ListBlogEntriesByAuthorForLister(cd.ctx, db.ListBlogEntriesByAuthorForListerParams{
			AuthorID: cd.currentProfileUserID,
			ListerID: cd.UserID,
			UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
			Limit:    15,
			Offset:   int32(cd.currentOffset),
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}
			return nil, err
		}
		var list []*db.ListBlogEntriesByAuthorForListerRow
		for _, row := range rows {
			if !cd.HasGrant("blogs", "entry", "see", row.Idblogs) {
				continue
			}
			list = append(list, row)
		}
		return list, nil
	})
}

// Bookmarks returns the user's bookmark list loaded lazily.
func (cd *CoreData) Bookmarks() (*db.GetBookmarksForUserRow, error) {
	return cd.bookmarks.Load(func() (*db.GetBookmarksForUserRow, error) {
		if cd.UserID == 0 || cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetBookmarksForUser(cd.ctx, cd.UserID)
	})
}

// CreateBookmark inserts a bookmark list for the current user.
//
// It wraps the CreateBookmarksForLister query and updates the cached
// bookmarks value on success.
func (cd *CoreData) CreateBookmark(params db.CreateBookmarksForListerParams) error {
	if cd.queries == nil {
		return nil
	}
	if err := cd.queries.CreateBookmarksForLister(cd.ctx, params); err != nil {
		return err
	}
	cd.bookmarks.Set(&db.GetBookmarksForUserRow{List: params.List})
	return nil
}

// SaveBookmark persists the user's bookmark list and updates the cache.
func (cd *CoreData) SaveBookmark(p db.UpdateBookmarksForListerParams) error {
	if cd.queries == nil {
		return nil
	}
	if err := cd.queries.UpdateBookmarksForLister(cd.ctx, p); err != nil {
		return err
	}
	cd.bookmarks = lazy.Value[*db.GetBookmarksForUserRow]{}
	cd.bookmarks.Set(&db.GetBookmarksForUserRow{List: p.List})
	return nil
}

// IsAdmin reports whether the current user has the administrator role.
func (cd *CoreData) IsAdmin() bool {
	return cd.HasRole("administrator")
}

// IsAdminMode reports whether admin-only UI elements should be displayed.
func (cd *CoreData) IsAdminMode() bool { return cd.AdminMode }

// CanEditBlog reports whether the current user may edit the specified blog
// entry via the public interface.
func (cd *CoreData) CanEditBlog(entryID, ownerID int32) bool {
	return ownerID == cd.UserID && cd.HasGrant("blogs", "entry", "edit", entryID)
}

// ShowReplyNews reports whether replies are permitted on the specified news
// post.
func (cd *CoreData) ShowReplyNews(id int32) bool {
	return cd.HasGrant("news", "post", "reply", id)
}

// ShowEditNews reports whether the current user may edit the supplied news
// post via the public interface.
func (cd *CoreData) ShowEditNews(id, ownerID int32) bool {
	return ownerID == cd.UserID && cd.HasGrant("news", "post", "edit", id)
}

// CommentByID returns a forum comment lazily loading it once per ID.
func (cd *CoreData) CommentByID(id int32, ops ...lazy.Option[*db.GetCommentByIdForUserRow]) (*db.GetCommentByIdForUserRow, error) {
	fetch := func(i int32) (*db.GetCommentByIdForUserRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetCommentByIdForUser(cd.ctx, db.GetCommentByIdForUserParams{
			ViewerID: cd.UserID,
			ID:       i,
			UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
	}
	return lazy.Map(&cd.forumComments, &cd.mapMu, id, fetch, ops...)
}

func (cd *CoreData) composeMapper() {
	var fns []func(tag, val string) string
	if cd.ImageSigner != nil {
		fns = append(fns, cd.ImageSigner.MapURL)
	}
	if cd.LinkSigner != nil {
		fns = append(fns, cd.LinkSigner.MapURL)
	}
	if len(fns) == 0 {
		cd.a4codeMapper = nil
		return
	}
	cd.a4codeMapper = func(tag, val string) string {
		for _, fn := range fns {
			newVal := fn(tag, val)
			if newVal != val {
				return newVal
			}
			val = newVal
		}
		return val
	}
}

// CurrentBlog returns the currently requested blog entry lazily loaded.
func (cd *CoreData) CurrentBlog(ops ...lazy.Option[*db.GetBlogEntryForListerByIDRow]) (*db.GetBlogEntryForListerByIDRow, error) {
	if cd.currentBlogID == 0 {
		return nil, nil
	}
	return cd.BlogEntryByID(cd.currentBlogID, ops...)
}

// CurrentBlogLoaded returns the cached current blog entry without database access.
func (cd *CoreData) CurrentBlogLoaded() *db.GetBlogEntryForListerByIDRow {
	if cd.blogEntries == nil {
		return nil
	}
	lv, ok := cd.blogEntries[cd.currentBlogID]
	if !ok {
		return nil
	}
	v, ok := lv.Peek()
	if !ok {
		return nil
	}
	return v
}

// CurrentComment returns the current comment lazily loaded.
func (cd *CoreData) CurrentComment(r *http.Request, ops ...lazy.Option[*db.GetCommentByIdForUserRow]) (*db.GetCommentByIdForUserRow, error) {
	if cd.currentCommentID == 0 {
		if r != nil {
			idStr := r.URL.Query().Get("comment")
			if idStr == "" {
				if vars := mux.Vars(r); vars != nil {
					idStr = vars["comment"]
				}
			}
			if idStr != "" {
				id, err := strconv.Atoi(idStr)
				if err != nil {
					return nil, fmt.Errorf("invalid comment id: %w", err)
				}
				cd.currentCommentID = int32(id)
			}
		}
		if cd.currentCommentID == 0 {
			return nil, nil
		}
	}
	return cd.CommentByID(cd.currentCommentID, ops...)
}

// CurrentCommentLoaded returns the cached current comment if available.
func (cd *CoreData) CurrentCommentLoaded() *db.GetCommentByIdForUserRow {
	if cd.forumComments == nil {
		return nil
	}
	lv, ok := cd.forumComments[cd.currentCommentID]
	if !ok {
		return nil
	}
	v, ok := lv.Peek()
	if !ok {
		return nil
	}
	return v
}

// CurrentNewsPost returns the current news post lazily loaded.
func (cd *CoreData) CurrentNewsPost(ops ...lazy.Option[*db.GetForumThreadIdByNewsPostIdRow]) (*db.GetForumThreadIdByNewsPostIdRow, error) {
	if cd.currentNewsPostID == 0 {
		return nil, nil
	}
	return cd.NewsPostByID(cd.currentNewsPostID, ops...)
}

// CurrentNewsPostLoaded returns the cached current news post if available.
func (cd *CoreData) CurrentNewsPostLoaded() *db.GetForumThreadIdByNewsPostIdRow {
	if cd.newsPosts == nil {
		return nil
	}
	lv, ok := cd.newsPosts[cd.currentNewsPostID]
	if !ok {
		return nil
	}
	v, ok := lv.Peek()
	if !ok {
		return nil
	}
	return v
}

// CurrentProfileBookmarkSize returns bookmark entry count for the profile user.
func (cd *CoreData) CurrentProfileBookmarkSize() int {
	id := cd.currentProfileUserID
	if id == 0 {
		return 0
	}
	if cd.adminUserBookmarkSize == nil {
		cd.adminUserBookmarkSize = map[int32]*lazy.Value[int]{}
	}
	lv, ok := cd.adminUserBookmarkSize[id]
	if !ok {
		lv = &lazy.Value[int]{}
		cd.adminUserBookmarkSize[id] = lv
	}
	size, err := lv.Load(func() (int, error) {
		if cd.queries == nil {
			return 0, nil
		}
		bm, err := cd.queries.GetBookmarksForUser(cd.ctx, id)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return 0, err
		}
		if bm == nil {
			return 0, nil
		}
		list := strings.TrimSpace(bm.List.String)
		if list == "" {
			return 0, nil
		}
		return len(strings.Split(list, "\n")), nil
	})
	if err != nil {
		log.Printf("load bookmark size: %v", err)
		return 0
	}
	return size
}

// CurrentProfileComments returns admin comments for the profile user.
func (cd *CoreData) CurrentProfileComments() []*db.AdminUserComment {
	id := cd.currentProfileUserID
	if id == 0 {
		return nil
	}
	if cd.adminUserComments == nil {
		cd.adminUserComments = map[int32]*lazy.Value[[]*db.AdminUserComment]{}
	}
	lv, ok := cd.adminUserComments[id]
	if !ok {
		lv = &lazy.Value[[]*db.AdminUserComment]{}
		cd.adminUserComments[id] = lv
	}
	rows, err := lv.Load(func() ([]*db.AdminUserComment, error) {
		if cd.queries == nil {
			return nil, nil
		}
		comments, err := cd.queries.ListAdminUserComments(cd.ctx, id)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return comments, nil
	})
	if err != nil {
		log.Printf("load user comments: %v", err)
		return nil
	}
	return rows
}

// CurrentProfileEmails returns emails for the profile user.
func (cd *CoreData) CurrentProfileEmails() []*db.UserEmail {
	id := cd.currentProfileUserID
	if id == 0 {
		return nil
	}
	if cd.adminUserEmails == nil {
		cd.adminUserEmails = map[int32]*lazy.Value[[]*db.UserEmail]{}
	}
	lv, ok := cd.adminUserEmails[id]
	if !ok {
		lv = &lazy.Value[[]*db.UserEmail]{}
		cd.adminUserEmails[id] = lv
	}
	rows, err := lv.Load(func() ([]*db.UserEmail, error) {
		if cd.queries == nil {
			return nil, nil
		}
		emails, err := cd.queries.AdminListUserEmails(cd.ctx, id)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return emails, nil
	})
	if err != nil {
		log.Printf("load user emails: %v", err)
		return nil
	}
	return rows
}

// CurrentProfileGrants returns direct grants for the profile user.
func (cd *CoreData) CurrentProfileGrants() []*db.Grant {
	id := cd.currentProfileUserID
	if id == 0 {
		return nil
	}
	if cd.adminUserGrants == nil {
		cd.adminUserGrants = map[int32]*lazy.Value[[]*db.Grant]{}
	}
	lv, ok := cd.adminUserGrants[id]
	if !ok {
		lv = &lazy.Value[[]*db.Grant]{}
		cd.adminUserGrants[id] = lv
	}
	rows, err := lv.Load(func() ([]*db.Grant, error) {
		if cd.queries == nil {
			return nil, nil
		}
		grants, err := cd.queries.ListGrantsByUserID(cd.ctx, sql.NullInt32{Int32: id, Valid: true})
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return grants, nil
	})
	if err != nil {
		log.Printf("load user grants: %v", err)
		return nil
	}
	return rows
}

// CurrentProfileRoles returns roles for the profile user.
func (cd *CoreData) CurrentProfileRoles() []*db.GetPermissionsByUserIDRow {
	id := cd.currentProfileUserID
	if id == 0 {
		return nil
	}
	if cd.adminUserRoles == nil {
		cd.adminUserRoles = map[int32]*lazy.Value[[]*db.GetPermissionsByUserIDRow]{}
	}
	lv, ok := cd.adminUserRoles[id]
	if !ok {
		lv = &lazy.Value[[]*db.GetPermissionsByUserIDRow]{}
		cd.adminUserRoles[id] = lv
	}
	rows, err := lv.Load(func() ([]*db.GetPermissionsByUserIDRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		roles, err := cd.queries.GetPermissionsByUserID(cd.ctx, id)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return roles, nil
	})
	if err != nil {
		log.Printf("load user roles: %v", err)
		return nil
	}
	return rows
}

// CurrentProfileStats returns posting stats for the profile user.
func (cd *CoreData) CurrentProfileStats() *db.AdminUserPostCountsByIDRow {
	id := cd.currentProfileUserID
	if id == 0 {
		return nil
	}
	if cd.adminUserStats == nil {
		cd.adminUserStats = map[int32]*lazy.Value[*db.AdminUserPostCountsByIDRow]{}
	}
	lv, ok := cd.adminUserStats[id]
	if !ok {
		lv = &lazy.Value[*db.AdminUserPostCountsByIDRow]{}
		cd.adminUserStats[id] = lv
	}
	row, err := lv.Load(func() (*db.AdminUserPostCountsByIDRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		stat, err := cd.queries.AdminUserPostCountsByID(cd.ctx, id)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return stat, nil
	})
	if err != nil {
		log.Printf("load user stats: %v", err)
		return nil
	}
	return row
}

// CurrentProfileUser returns the user being viewed.
func (cd *CoreData) CurrentProfileUser() *db.SystemGetUserByIDRow {
	return cd.UserByID(cd.currentProfileUserID)
}

// CurrentRequest returns the request currently being viewed.
func (cd *CoreData) CurrentRequest() *db.AdminRequestQueue {
	id := cd.currentRequestID
	if id == 0 {
		return nil
	}
	if cd.adminRequest == nil {
		cd.adminRequest = map[int32]*lazy.Value[*db.AdminRequestQueue]{}
	}
	lv, ok := cd.adminRequest[id]
	if !ok {
		lv = &lazy.Value[*db.AdminRequestQueue]{}
		cd.adminRequest[id] = lv
	}
	req, err := lv.Load(func() (*db.AdminRequestQueue, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.AdminGetRequestByID(cd.ctx, id)
	})
	if err != nil {
		log.Printf("load request %d: %v", id, err)
		return nil
	}
	return req
}

// CurrentRequestComments returns comments for the current request.
func (cd *CoreData) CurrentRequestComments() []*db.AdminRequestComment {
	id := cd.currentRequestID
	if id == 0 {
		return nil
	}
	if cd.adminRequestComments == nil {
		cd.adminRequestComments = map[int32]*lazy.Value[[]*db.AdminRequestComment]{}
	}
	lv, ok := cd.adminRequestComments[id]
	if !ok {
		lv = &lazy.Value[[]*db.AdminRequestComment]{}
		cd.adminRequestComments[id] = lv
	}
	rows, err := lv.Load(func() ([]*db.AdminRequestComment, error) {
		if cd.queries == nil {
			return nil, nil
		}
		comments, err := cd.queries.AdminListRequestComments(cd.ctx, id)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return comments, nil
	})
	if err != nil {
		log.Printf("load request comments: %v", err)
		return nil
	}
	return rows
}

// CurrentRequestUser returns the user associated with the current request.
func (cd *CoreData) CurrentRequestUser() *db.SystemGetUserByIDRow {
	req := cd.CurrentRequest()
	if req == nil {
		return nil
	}
	return cd.UserByID(req.UsersIdusers)
}

// CurrentTopic returns the currently requested topic lazily loaded.
func (cd *CoreData) CurrentTopic(ops ...lazy.Option[*db.GetForumTopicByIdForUserRow]) (*db.GetForumTopicByIdForUserRow, error) {
	if cd.currentTopicID == 0 {
		return nil, nil
	}
	return cd.ForumTopicByID(cd.currentTopicID, ops...)
}

// CurrentTopicLoaded returns the cached current topic without database access.
func (cd *CoreData) CurrentTopicLoaded() *db.GetForumTopicByIdForUserRow {
	if cd.forumTopics == nil {
		return nil
	}
	lv, ok := cd.forumTopics[cd.currentTopicID]
	if !ok {
		return nil
	}
	v, ok := lv.Peek()
	if !ok {
		return nil
	}
	return v
}

// CurrentUser returns the logged in user's record loaded on demand.
func (cd *CoreData) CurrentUser() (*db.User, error) {
	return cd.user.Load(func() (*db.User, error) {
		if cd.UserID == 0 || cd.queries == nil {
			return nil, nil
		}
		row, err := cd.queries.SystemGetUserByID(cd.ctx, cd.UserID)
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
	u, ok := cd.user.Peek()
	if !ok {
		return nil
	}
	return u
}

// CurrentWriting returns the currently requested writing lazily loaded.
func (cd *CoreData) CurrentWriting(ops ...lazy.Option[*db.GetWritingForListerByIDRow]) (*db.GetWritingForListerByIDRow, error) {
	if cd.currentWritingID == 0 {
		return nil, nil
	}
	return cd.WritingByID(cd.currentWritingID, ops...)
}

// CurrentWritingLoaded returns the cached current writing without database access.
func (cd *CoreData) CurrentWritingLoaded() *db.GetWritingForListerByIDRow {
	if cd.writingRows == nil {
		return nil
	}
	lv, ok := cd.writingRows[cd.currentWritingID]
	if !ok {
		return nil
	}
	v, ok := lv.Peek()
	if !ok {
		return nil
	}
	return v
}

// CustomQueries returns the db.CustomQueries instance associated with this CoreData.
func (cd *CoreData) CustomQueries() db.CustomQueries { return cd.customQueries }

// DBRegistry returns the database driver registry associated with this request.
func (cd *CoreData) DBRegistry() *dbdrivers.Registry { return cd.dbRegistry }

// DefaultNotificationTemplate renders the default body for the current notification template.
func (cd *CoreData) DefaultNotificationTemplate() string {
	return defaultNotificationTemplate(cd.currentNotificationTemplateName, cd.Config)
}

// EmailProvider returns the configured email provider.
func (cd *CoreData) EmailProvider() MailProvider {
	p, err := cd.emailProvider.Load(func() (MailProvider, error) { return nil, nil })
	if err != nil {
		log.Printf("load email provider: %v", err)
	}
	return p
}

// Event returns the event associated with the request, if any.
func (cd *CoreData) Event() *eventbus.TaskEvent { return cd.event }

// ExecuteSiteTemplate renders the named site template using cd's helper
// functions. It wraps templates.GetCompiledSiteTemplates(cd.Funcs(r)).
func (cd *CoreData) ExecuteSiteTemplate(w io.Writer, r *http.Request, name string, data any) error {
	return templates.GetCompiledSiteTemplates(cd.Funcs(r)).ExecuteTemplate(w, name, data)
}

// ExternalLink lazily resolves metadata for id.
func (cd *CoreData) ExternalLink(id int32) *db.ExternalLink {
	if cd.queries == nil {
		return nil
	}
	if cd.externalLinks == nil {
		cd.externalLinks = make(map[int32]*lazy.Value[*db.ExternalLink])
	}
	lv, ok := cd.externalLinks[id]
	if !ok {
		lv = &lazy.Value[*db.ExternalLink]{}
		cd.externalLinks[id] = lv
	}
	link, err := lv.Load(func() (*db.ExternalLink, error) {
		l, err := cd.queries.GetExternalLinkByID(cd.ctx, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}
			return nil, err
		}
		return l, nil
	})
	if err != nil {
		log.Printf("load external link: %v", err)
	}
	return link
}

// fetchLatestNews loads news posts from the database with permission data.
func (cd *CoreData) fetchLatestNews(offset, limit int32) ([]*db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow, error) {
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
	var posts []*db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow
	for _, row := range rows {
		if !cd.HasGrant("news", "post", "see", row.Idsitenews) {
			continue
		}
		posts = append(posts, row)
	}
	return posts, nil
}

// FAQCategories returns FAQ categories loaded on demand.
func (cd *CoreData) FAQCategories() ([]*db.FaqCategory, error) {
	return cd.faqCategories.Load(func() ([]*db.FaqCategory, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.AdminGetFAQCategories(cd.ctx)
	})
}

// ForumCategories loads all forum categories once.
func (cd *CoreData) ForumCategories() ([]*db.Forumcategory, error) {
	return cd.forumCategories.Load(func() ([]*db.Forumcategory, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetAllForumCategories(cd.ctx, db.GetAllForumCategoriesParams{ViewerID: cd.UserID})
	})
}

// ForumThreadByID returns a single forum thread lazily loading it once per ID.
func (cd *CoreData) ForumThreadByID(id int32, ops ...lazy.Option[*db.GetThreadLastPosterAndPermsRow]) (*db.GetThreadLastPosterAndPermsRow, error) {
	fetch := func(i int32) (*db.GetThreadLastPosterAndPermsRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetThreadLastPosterAndPerms(cd.ctx, db.GetThreadLastPosterAndPermsParams{
			ViewerID:      cd.UserID,
			ThreadID:      i,
			ViewerMatchID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
	}
	return lazy.Map(&cd.forumThreadRows, &cd.mapMu, id, fetch, ops...)
}

// ForumThreads loads the threads for a forum topic once per topic.
func (cd *CoreData) ForumThreads(topicID int32) ([]*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow, error) {
	if cd.forumThreads == nil {
		cd.forumThreads = make(map[int32]*lazy.Value[[]*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow])
	}
	lv, ok := cd.forumThreads[topicID]
	if !ok {
		lv = &lazy.Value[[]*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow]{}
		cd.forumThreads[topicID] = lv
	}
	return lv.Load(func() ([]*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow, error) {
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

// ForumTopicByID loads a forum topic once per ID using caching.
func (cd *CoreData) ForumTopicByID(id int32, ops ...lazy.Option[*db.GetForumTopicByIdForUserRow]) (*db.GetForumTopicByIdForUserRow, error) {
	fetch := func(i int32) (*db.GetForumTopicByIdForUserRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetForumTopicByIdForUser(cd.ctx, db.GetForumTopicByIdForUserParams{
			ViewerID:      cd.UserID,
			Idforumtopic:  i,
			ViewerMatchID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
	}
	return lazy.Map(&cd.forumTopics, &cd.mapMu, id, fetch, ops...)
}

// ForumTopics loads forum topics for a given category once per category.
func (cd *CoreData) ForumTopics(categoryID int32) ([]*db.Forumtopic, error) {
	if cd.forumTopicLists == nil {
		cd.forumTopicLists = make(map[int32]*lazy.Value[[]*db.Forumtopic])
	}
	lv, ok := cd.forumTopicLists[categoryID]
	if !ok {
		lv = &lazy.Value[[]*db.Forumtopic]{}
		cd.forumTopicLists[categoryID] = lv
	}
	return lv.Load(func() ([]*db.Forumtopic, error) {
		if cd.queries == nil {
			return nil, nil
		}
		if categoryID == 0 {
			return cd.queries.GetAllForumTopics(cd.ctx, db.GetAllForumTopicsParams{ViewerID: cd.UserID})
		}
		return cd.queries.GetForumTopicsByCategoryId(cd.ctx, db.GetForumTopicsByCategoryIdParams{
			ViewerID:   cd.UserID,
			CategoryID: categoryID,
		})
	})
}

// HasAdminRole reports whether the current user has the administrator role.
func (cd *CoreData) HasAdminRole() bool {
	return cd.HasRole("administrator")
}

// HasContentWriterRole reports whether the current user has the content writer role.
func (cd *CoreData) HasContentWriterRole() bool {
	return cd.HasRole("content writer")
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
			if _, err := cd.queries.SystemCheckRoleGrant(cd.ctx, db.SystemCheckRoleGrantParams{Name: r, Action: role}); err == nil {
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

// HasSubscription reports whether the user has subscribed to pattern with method.
func (cd *CoreData) HasSubscription(pattern, method string) bool {
	m, _ := cd.subscriptionMap()
	return m[pattern+"|"+method]
}

// ImageBoardPosts retrieves approved posts for the board lazily.
func (cd *CoreData) ImageBoardPosts(boardID int32) ([]*db.ListImagePostsByBoardForListerRow, error) {
	if cd.queries == nil {
		return nil, nil
	}
	if cd.imageBoardPosts == nil {
		cd.imageBoardPosts = make(map[int32]*lazy.Value[[]*db.ListImagePostsByBoardForListerRow])
	}
	lv, ok := cd.imageBoardPosts[boardID]
	if !ok {
		lv = &lazy.Value[[]*db.ListImagePostsByBoardForListerRow]{}
		cd.imageBoardPosts[boardID] = lv
	}
	return lv.Load(func() ([]*db.ListImagePostsByBoardForListerRow, error) {
		return cd.queries.ListImagePostsByBoardForLister(cd.ctx, db.ListImagePostsByBoardForListerParams{
			ListerID:     cd.UserID,
			BoardID:      boardID,
			ListerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
			Limit:        200,
			Offset:       0,
		})
	})
}

// ImageBoards returns all image boards cached once.
func (cd *CoreData) ImageBoards() ([]*db.Imageboard, error) {
	return cd.imageBoards.Load(func() ([]*db.Imageboard, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.AdminListBoards(cd.ctx, db.AdminListBoardsParams{Limit: 200, Offset: 0})
	})
}

// ImagePostByID returns an image post once per ID using caching.
func (cd *CoreData) ImagePostByID(id int32, ops ...lazy.Option[*db.GetImagePostByIDForListerRow]) (*db.GetImagePostByIDForListerRow, error) {
	fetch := func(i int32) (*db.GetImagePostByIDForListerRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetImagePostByIDForLister(cd.ctx, db.GetImagePostByIDForListerParams{
			ListerID:     cd.UserID,
			ID:           i,
			ListerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
	}
	return lazy.Map(&cd.imagePostRows, &cd.mapMu, id, fetch, ops...)
}

// ImageURLMapper maps image references like "image:" or "cache:" to full URLs.
func (cd *CoreData) ImageURLMapper(tag, val string) string {
	if cd.a4codeMapper != nil {
		return cd.a4codeMapper(tag, val)
	}
	return val
}

// Languages returns the list of available languages loaded on demand.
func (cd *CoreData) Languages() ([]*db.Language, error) {
	return cd.langs.Load(func() ([]*db.Language, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.SystemListLanguages(cd.ctx)
	})
}

// RenameLanguage updates the language code from oldCode to newCode and clears
// the cached language list.
func (cd *CoreData) RenameLanguage(oldCode, newCode string) error {
	if cd.queries == nil {
		return fmt.Errorf("queries not set")
	}
	id, err := cd.queries.SystemGetLanguageIDByName(cd.ctx, sql.NullString{String: oldCode, Valid: true})
	if err != nil {
		return fmt.Errorf("lookup language id: %w", err)
	}
	if err := cd.queries.AdminRenameLanguage(cd.ctx, db.AdminRenameLanguageParams{
		Nameof:     sql.NullString{String: newCode, Valid: true},
		Idlanguage: id,
	}); err != nil {
		return fmt.Errorf("update language: %w", err)
	}
	cd.langs = lazy.Value[[]*db.Language]{}
	return nil
}


// DeleteLanguage removes a language when it isn't referenced by any content.
// The provided code is expected to be the language identifier string.
// It returns the resolved language ID and name.
func (cd *CoreData) DeleteLanguage(code string) (int32, string, error) {
	if cd.queries == nil {
		return 0, "", nil
	}
	id, err := strconv.Atoi(code)
	if err != nil {
		return 0, "", err
	}
	var name string
	if rows, err := cd.Languages(); err == nil {
		for _, l := range rows {
			if l.Idlanguage == int32(id) {
				name = l.Nameof.String
				break
			}
		}
	}
	counts, err := cd.queries.AdminLanguageUsageCounts(cd.ctx, db.AdminLanguageUsageCountsParams{ID: int32(id)})
	if err != nil {
		return int32(id), name, err
	}
	if counts.Comments > 0 || counts.Writings > 0 || counts.Blogs > 0 || counts.News > 0 || counts.Links > 0 {
		return int32(id), name, fmt.Errorf("language has content")
	}
	if err := cd.queries.AdminDeleteLanguage(cd.ctx, int32(id)); err != nil {
		return int32(id), name, err
	}
	cd.langs = lazy.Value[[]*db.Language]{}
	return int32(id), name, nil
}

// CreateLanguage inserts a new language and returns its ID.
//
// Parameters:
//
//	code - Language code (currently unused).
//	name - Display name of the language.
func (cd *CoreData) CreateLanguage(code, name string) (int64, error) {
	if cd.queries == nil {
		return 0, nil
	}
	_ = code
	res, err := cd.queries.AdminInsertLanguage(cd.ctx, sql.NullString{String: name, Valid: true})
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// LatestNews returns recent news posts with permission data.
func (cd *CoreData) LatestNews(r *http.Request) ([]*db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow, error) {
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	return cd.latestNews.Load(func() ([]*db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow, error) {
		return cd.fetchLatestNews(int32(offset), int32(cd.PageSize()))
	})
}

// LatestNewsList returns recent news posts without needing an HTTP request.
func (cd *CoreData) LatestNewsList(offset, limit int32) ([]*db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow, error) {
	return cd.fetchLatestNews(offset, limit)
}

func (cd *CoreData) LatestWritings(opts ...LatestWritingsOption) ([]*db.Writing, error) {
	return cd.latestWritings.Load(func() ([]*db.Writing, error) {
		if cd.queries == nil {
			return nil, nil
		}
		params := db.GetPublicWritingsParams{Limit: int32(cd.PageSize())}
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

// LinkerCategories returns all linker categories.
func (cd *CoreData) LinkerCategories() ([]*db.LinkerCategory, error) {
	return cd.linkerCatsAll.Load(func() ([]*db.LinkerCategory, error) {
		if cd.queries == nil {
			return nil, nil
		}
		rows, err := cd.queries.GetAllLinkerCategories(cd.ctx)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return rows, nil
	})
}

// LinkerCategoriesForUser returns linker categories the viewer can access.
func (cd *CoreData) LinkerCategoriesForUser() ([]*db.LinkerCategory, error) {
	return cd.linkerCatsForUser.Load(func() ([]*db.LinkerCategory, error) {
		if cd.queries == nil {
			return nil, nil
		}
		rows, err := cd.queries.GetAllLinkerCategoriesForUser(cd.ctx, db.GetAllLinkerCategoriesForUserParams{
			ViewerID:     cd.UserID,
			ViewerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return rows, nil
	})
}

// LinkerCategoryByID returns a linker category lazily loading it once per ID.
func (cd *CoreData) LinkerCategoryByID(id int32, ops ...lazy.Option[*db.LinkerCategory]) (*db.LinkerCategory, error) {
	fetch := func(i int32) (*db.LinkerCategory, error) {
		if cd.queries == nil {
			return nil, nil
		}
		cat, err := cd.queries.GetLinkerCategoryById(cd.ctx, i)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return cat, nil
	}
	return lazy.Map(&cd.linkerCategoryRows, &cd.mapMu, id, fetch, ops...)
}

// LinkerCategoryCounts lazily loads linker category statistics.
func (cd *CoreData) LinkerCategoryCounts() ([]*db.GetLinkerCategoryLinkCountsRow, error) {
	return cd.linkerCategories.Load(func() ([]*db.GetLinkerCategoryLinkCountsRow, error) {
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

// CreateFAQCategory adds a new FAQ category.
func (cd *CoreData) CreateFAQCategory(name string) error {
	if cd.queries == nil {
		return nil
	}
	return cd.queries.AdminCreateFAQCategory(cd.ctx, sql.NullString{String: name, Valid: name != ""})
}

// LinkerItemsForUser returns linker items for the given category and offset respecting viewer permissions.
func (cd *CoreData) LinkerItemsForUser(catID, offset int32) ([]*db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginatedRow, error) {
	if cd.queries == nil {
		return nil, nil
	}
	rows, err := cd.queries.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginated(cd.ctx, db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginatedParams{
		ViewerID:         cd.UserID,
		Idlinkercategory: catID,
		ViewerUserID:     sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		Limit:            int32(cd.PageSize()),
		Offset:           offset,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	var out []*db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginatedRow
	for _, row := range rows {
		if cd.HasGrant("linker", "link", "see", row.Idlinker) {
			out = append(out, row)
		}
	}
	return out, nil
}

// LinkerLinksByCategoryID returns the links for a category lazily loading them once per ID.
func (cd *CoreData) LinkerLinksByCategoryID(id int32, ops ...lazy.Option[[]*db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingRow]) ([]*db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingRow, error) {
	fetch := func(i int32) ([]*db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		rows, err := cd.queries.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescending(cd.ctx, db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingParams{Idlinkercategory: i})
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return rows, nil
	}
	return lazy.Map(&cd.linkerCategoryLinks, &cd.mapMu, id, fetch, ops...)
}

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

// newsAnnouncement returns the latest announcement for the given news post.
// The result is cached so repeated lookups for the same id hit the database
// only once.
func (cd *CoreData) newsAnnouncement(id int32) (*db.SiteAnnouncement, error) {
	cd.annMu.Lock()
	lv, ok := cd.newsAnnouncements[id]
	if !ok {
		lv = &lazy.Value[*db.SiteAnnouncement]{}
		cd.newsAnnouncements[id] = lv
	}
	cd.annMu.Unlock()

	return lv.Load(func() (*db.SiteAnnouncement, error) {
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

// NewsAnnouncement returns the latest announcement for the given news post.
// Errors are logged and result nil.
func (cd *CoreData) NewsAnnouncement(id int32) *db.SiteAnnouncement {
	ann, err := cd.newsAnnouncement(id)
	if err != nil {
		log.Printf("news announcement %d: %v", id, err)
	}
	return ann
}

// NewsAnnouncementWithErr is like NewsAnnouncement but returns any load error.
func (cd *CoreData) NewsAnnouncementWithErr(id int32) (*db.SiteAnnouncement, error) {
	return cd.newsAnnouncement(id)
}

// NewsPostByID returns the news post lazily loading it once per ID.
func (cd *CoreData) NewsPostByID(id int32, ops ...lazy.Option[*db.GetForumThreadIdByNewsPostIdRow]) (*db.GetForumThreadIdByNewsPostIdRow, error) {
	fetch := func(i int32) (*db.GetForumThreadIdByNewsPostIdRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetForumThreadIdByNewsPostId(cd.ctx, i)
	}
	return lazy.Map(&cd.newsPosts, &cd.mapMu, id, fetch, ops...)
}

// Notifications returns the notifications for the current user using query
// parameters to control pagination. Results are cached per offset and filter
// combination.
func (cd *CoreData) Notifications(r *http.Request) ([]*db.Notification, error) {
	if cd.notifications == nil {
		cd.notifications = map[string]*lazy.Value[[]*db.Notification]{}
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	showAll := r.URL.Query().Get("all") == "1"
	key := fmt.Sprintf("%t:%d", showAll, offset)
	lv, ok := cd.notifications[key]
	if !ok {
		lv = &lazy.Value[[]*db.Notification]{}
		cd.notifications[key] = lv
	}
	return lv.Load(func() ([]*db.Notification, error) {
		if cd.queries == nil || cd.UserID == 0 {
			return nil, nil
		}
		limit := int32(cd.PageSize())
		if showAll {
			return cd.queries.ListNotificationsForLister(cd.ctx, db.ListNotificationsForListerParams{
				ListerID: cd.UserID,
				Limit:    limit,
				Offset:   int32(offset),
			})
		}
		return cd.queries.ListUnreadNotificationsForLister(cd.ctx, db.ListUnreadNotificationsForListerParams{
			ListerID: cd.UserID,
			Limit:    limit,
			Offset:   int32(offset),
		})
	})
}

// PageSize returns the preferred page size within configured limits.
func (cd *CoreData) PageSize() int {
	size := cd.Config.PageSizeDefault
	if pref, err := cd.Preference(); err == nil && pref != nil && pref.PageSize != 0 {
		size = int(pref.PageSize)
	}
	if size < cd.Config.PageSizeMin {
		size = cd.Config.PageSizeMin
	}
	if size > cd.Config.PageSizeMax {
		size = cd.Config.PageSizeMax
	}
	return size
}

// PendingRequests returns pending admin requests loaded on demand.
func (cd *CoreData) PendingRequests() []*db.AdminRequestQueue {
	rows, err := cd.adminRequestList("pending")
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("load pending requests: %v", err)
		return nil
	}
	return rows
}

// Permissions returns the user's permissions loaded on demand.
func (cd *CoreData) Permissions() ([]*db.GetPermissionsByUserIDRow, error) {
	return cd.perms.Load(func() ([]*db.GetPermissionsByUserIDRow, error) {
		if cd.UserID == 0 || cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetPermissionsByUserID(cd.ctx, cd.UserID)
	})
}

// Preference returns the user's preferences loaded on demand.
func (cd *CoreData) Preference() (*db.Preference, error) {
	return cd.pref.Load(func() (*db.Preference, error) {
		if cd.UserID == 0 || cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetPreferenceForLister(cd.ctx, cd.UserID)
	})
}

// PreferredLanguageID returns the user's preferred language ID if set,
// otherwise it resolves the site's default language name to an ID.
func (cd *CoreData) PreferredLanguageID(siteDefault string) int32 {
	id, err := cd.preferredLanguageID.Load(func() (int32, error) {
		if pref, err := cd.Preference(); err == nil && pref != nil {
			if pref.LanguageIdlanguage != 0 {
				return pref.LanguageIdlanguage, nil
			}
		}
		if cd.queries == nil || siteDefault == "" {
			return 0, nil
		}
		langID, err := cd.queries.SystemGetLanguageIDByName(cd.ctx, sql.NullString{String: siteDefault, Valid: true})
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

// Location returns the time.Location used for displaying times. The user's
// preferred timezone is used when set; otherwise the site configuration
// timezone is applied. UTC is returned as a safe fallback.
func (cd *CoreData) Location() *time.Location {
	if pref, err := cd.Preference(); err == nil && pref != nil {
		if pref.Timezone.Valid {
			if loc, err := time.LoadLocation(pref.Timezone.String); err == nil {
				return loc
			}
		}
	}
	if cd.Config != nil && cd.Config.Timezone != "" {
		if loc, err := time.LoadLocation(cd.Config.Timezone); err == nil {
			return loc
		}
	}
	return time.UTC
}

// PublicWritings returns public writings in a category, cached per category and offset.
func (cd *CoreData) PublicWritings(categoryID int32, r *http.Request) ([]*db.ListPublicWritingsInCategoryForListerRow, error) {
	if cd.publicWritings == nil {
		cd.publicWritings = map[string]*lazy.Value[[]*db.ListPublicWritingsInCategoryForListerRow]{}
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	key := fmt.Sprintf("%d:%d", categoryID, offset)
	lv, ok := cd.publicWritings[key]
	if !ok {
		lv = &lazy.Value[[]*db.ListPublicWritingsInCategoryForListerRow]{}
		cd.publicWritings[key] = lv
	}
	return lv.Load(func() ([]*db.ListPublicWritingsInCategoryForListerRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		rows, err := cd.queries.ListPublicWritingsInCategoryForLister(cd.ctx, db.ListPublicWritingsInCategoryForListerParams{
			ListerID:          cd.UserID,
			WritingCategoryID: categoryID,
			UserID:            sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
			Limit:             int32(cd.PageSize()),
			Offset:            int32(offset),
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}
			return nil, err
		}
		var res []*db.ListPublicWritingsInCategoryForListerRow
		for _, row := range rows {
			if cd.HasGrant("writing", "article", "see", row.Idwriting) {
				res = append(res, row)
			}
		}
		return res, nil
	})
}

// Queries returns the db.Queries instance associated with this CoreData.
func (cd *CoreData) Queries() db.Querier { return cd.queries }

// SelectedQuestionFromCategory deletes the specified FAQ question after
// verifying it belongs to the provided category.
func (cd *CoreData) SelectedQuestionFromCategory(questionID, categoryID int32) error {
	if cd.queries == nil {
		return fmt.Errorf("queries not available")
	}
	question, err := cd.queries.AdminGetFAQByID(cd.ctx, questionID)
	if err != nil {
		return err
	}
	if question.FaqcategoriesIdfaqcategories != categoryID {
		return fmt.Errorf("question %d not in category %d", questionID, categoryID)
	}
	return cd.queries.AdminDeleteFAQ(cd.ctx, questionID)
}

// UpdateFAQQuestion updates a FAQ question, changing its text, answer and
// category while recording a revision for the user.
func (cd *CoreData) UpdateFAQQuestion(question, answer string, categoryID, faqID, userID int32) error {
	if cd.queries == nil {
		return nil
	}
	if err := cd.queries.AdminUpdateFAQQuestionAnswer(cd.ctx, db.AdminUpdateFAQQuestionAnswerParams{
		Answer:                       sql.NullString{String: answer, Valid: true},
		Question:                     sql.NullString{String: question, Valid: true},
		FaqcategoriesIdfaqcategories: categoryID,
		Idfaq:                        faqID,
	}); err != nil {
		return err
	}
	if err := cd.queries.InsertFAQRevisionForUser(cd.ctx, db.InsertFAQRevisionForUserParams{
		FaqID:        faqID,
		UsersIdusers: userID,
		Question:     sql.NullString{String: question, Valid: true},
		Answer:       sql.NullString{String: answer, Valid: true},
		UserID:       sql.NullInt32{Int32: userID, Valid: true},
		ViewerID:     userID,
	}); err != nil {
		log.Printf("insert faq revision: %v", err)
	}
	return nil
}

// DeleteFAQCategory removes a FAQ category.
func (cd *CoreData) DeleteFAQCategory(id int32) error {
	if cd.queries == nil {
		return nil
	}
	return cd.queries.AdminDeleteFAQCategory(cd.ctx, id)
}

// AssociateEmail creates an email association request for a user.
func (cd *CoreData) AssociateEmail(p AssociateEmailParams) (*db.SystemGetUserByUsernameRow, int64, error) {
	row, err := cd.queries.SystemGetUserByUsername(cd.ctx, sql.NullString{String: p.Username, Valid: true})
	if err != nil {
		return nil, 0, fmt.Errorf("user not found %w", err)
	}
	if row.Email != "" {
		return nil, 0, ErrEmailAlreadyAssociated
	}
	res, err := cd.queries.AdminInsertRequestQueue(cd.ctx, db.AdminInsertRequestQueueParams{
		UsersIdusers:   row.Idusers,
		ChangeTable:    "user_emails",
		ChangeField:    "email",
		ChangeRowID:    row.Idusers,
		ChangeValue:    sql.NullString{String: p.Email, Valid: true},
		ContactOptions: sql.NullString{String: p.Email, Valid: true},
	})
	if err != nil {
		log.Printf("insert admin request: %v", err)
		return nil, 0, fmt.Errorf("insert admin request %w", err)
	}
	id, _ := res.LastInsertId()
	_ = cd.queries.AdminInsertRequestComment(cd.ctx, db.AdminInsertRequestCommentParams{RequestID: int32(id), Comment: p.Reason})
	_ = cd.queries.InsertAdminUserComment(cd.ctx, db.InsertAdminUserCommentParams{UsersIdusers: row.Idusers, Comment: "email association requested"})
	return row, id, nil
}

// DeleteFAQQuestion removes a FAQ entry by ID.
func (cd *CoreData) DeleteFAQQuestion(id int32) error {
	if cd.queries == nil {
		return nil
	}
	return cd.queries.AdminDeleteFAQ(cd.ctx, id)
}

// UserExists reports whether a user already exists with the supplied username
// or email address.
func (cd *CoreData) UserExists(username, email string) (bool, error) {
	if cd.queries == nil {
		return false, nil
	}
	if username != "" {
		if _, err := cd.queries.SystemGetUserByUsername(cd.ctx, sql.NullString{String: username, Valid: true}); err == nil {
			return true, nil
		} else if !errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("user by username: %w", err)
		}
	}
	if email != "" {
		if _, err := cd.queries.SystemGetUserByEmail(cd.ctx, email); err == nil {
			return true, nil
		} else if !errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("user by email: %w", err)
		}
	}
	return false, nil
}

// CreateUserWithEmail inserts a user with the supplied username, email and
// password hash/algorithm, returning the new user ID.
func (cd *CoreData) CreateUserWithEmail(u, e, hash, alg string) (int32, error) {
	if cd.queries == nil {
		return 0, errors.New("no queries")
	}
	id, err := cd.queries.SystemInsertUser(cd.ctx, sql.NullString{String: u, Valid: u != ""})
	if err != nil {
		return 0, err
	}
	uid := int32(id)
	if err := cd.queries.InsertUserEmail(cd.ctx, db.InsertUserEmailParams{UserID: uid, Email: e, VerifiedAt: sql.NullTime{}, LastVerificationCode: sql.NullString{}}); err != nil {
		return 0, err
	}
	if err := cd.queries.InsertPassword(cd.ctx, db.InsertPasswordParams{UsersIdusers: uid, Passwd: hash, PasswdAlgorithm: sql.NullString{String: alg, Valid: alg != ""}}); err != nil {
		return 0, err
	}
	return uid, nil
}

// RegisterExternalLinkClick records click statistics for url.
func (cd *CoreData) RegisterExternalLinkClick(url string) {
	if cd.queries == nil {
		return
	}
	if err := cd.queries.SystemRegisterExternalLinkClick(cd.ctx, url); err != nil {
		log.Printf("record external link click: %v", err)
	}
}

// Role returns the first loaded role or "anyone" when none.
func (cd *CoreData) Role() string {
	roles := cd.UserRoles()
	if len(roles) == 0 {
		return "anyone"
	}
	return roles[0]
}

// SelectedAdminLinkerItem returns the linker item for the ID found in the request.
func (cd *CoreData) SelectedAdminLinkerItem(r *http.Request, ops ...lazy.Option[*db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow]) (*db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow, int32, error) {
	id, err := cd.SelectedAdminLinkerItemID(r)
	if err != nil {
		return nil, 0, err
	}
	link, err := cd.AdminLinkerItemByID(id, ops...)
	if err != nil {
		return nil, id, err
	}
	return link, id, nil
}

// SelectedAdminLinkerItemID extracts the linker item ID from URL vars, form values or query parameters.
func (cd *CoreData) SelectedAdminLinkerItemID(r *http.Request) (int32, error) {
	var idStr string
	if v, ok := mux.Vars(r)["link"]; ok {
		idStr = v
	} else if v := r.PostFormValue("link"); v != "" {
		idStr = v
	} else {
		idStr = r.URL.Query().Get("link")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		return 0, sql.ErrNoRows
	}
	return int32(id), nil
}

// SelectedBoardPosts returns posts for the current board without requiring an ID.
func (cd *CoreData) SelectedBoardPosts() ([]*db.ListImagePostsByBoardForListerRow, error) {
	if cd.currentBoardID == 0 {
		return nil, nil
	}
	return cd.ImageBoardPosts(cd.currentBoardID)
}

// SelectedBoardSubBoards returns sub-boards for the current board without requiring an ID.
func (cd *CoreData) SelectedBoardSubBoards() ([]*db.Imageboard, error) {
	if cd.currentBoardID == 0 {
		return nil, nil
	}
	return cd.SubImageBoards(cd.currentBoardID)
}

// SelectedCategoryPublicWritings returns public writings for the given category.
func (cd *CoreData) SelectedCategoryPublicWritings(categoryID int32, r *http.Request) ([]*db.ListPublicWritingsInCategoryForListerRow, error) {
	return cd.PublicWritings(categoryID, r)
}

// SelectedLinkerCategory returns the linker category for the given ID.
func (cd *CoreData) SelectedLinkerCategory(id int32, ops ...lazy.Option[*db.LinkerCategory]) (*db.LinkerCategory, error) {
	return cd.LinkerCategoryByID(id, ops...)
}

// SelectedLinkerItemsForCurrentUser returns linker items for the given category
// and offset for the current user and ensures the category is cached.
func (cd *CoreData) SelectedLinkerItemsForCurrentUser(catID, offset int32) ([]*db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginatedRow, error) {
	if catID != 0 {
		if _, err := cd.SelectedLinkerCategory(catID); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}
	return cd.LinkerItemsForUser(catID, offset)
}

// SelectedThread returns the currently requested thread lazily loaded.
func (cd *CoreData) SelectedThread(ops ...lazy.Option[*db.GetThreadLastPosterAndPermsRow]) (*db.GetThreadLastPosterAndPermsRow, error) {
	if cd.currentThreadID == 0 {
		return nil, nil
	}
	return cd.ForumThreadByID(cd.currentThreadID, ops...)
}

// SelectedThreadComments returns comments for the current thread without requiring an ID.
func (cd *CoreData) SelectedThreadComments() ([]*db.GetCommentsByThreadIdForUserRow, error) {
	if cd.currentThreadID == 0 {
		return nil, nil
	}
	return cd.ThreadComments(cd.currentThreadID)
}

// sectionItemType returns the default grant item type for a section.
func sectionItemType(section string) string {
	switch section {
	case "blogs":
		return "entry"
	case "news":
		return "post"
	case "imagebbs":
		return "board"
	case "linker":
		return "link"
	case "writing":
		return "article"
	default:
		return ""
	}
}

// SelectedSectionThreadComments returns comments for the current thread using
// the stored section with its default item type.
func (cd *CoreData) SelectedSectionThreadComments() ([]*db.GetCommentsByThreadIdForUserRow, error) {
	if cd.currentThreadID == 0 || cd.currentSection == "" {
		return nil, nil
	}
	return cd.SectionThreadComments(cd.currentSection, sectionItemType(cd.currentSection), cd.currentThreadID)
}

// SelectedThreadLoaded returns the cached current thread without database access.
func (cd *CoreData) SelectedThreadLoaded() *db.GetThreadLastPosterAndPermsRow {
	if cd.forumThreadRows == nil {
		return nil
	}
	lv, ok := cd.forumThreadRows[cd.currentThreadID]
	if !ok {
		return nil
	}
	v, ok := lv.Peek()
	if !ok {
		return nil
	}
	return v
}

// SelectedThreadCanReply reports whether the current user may reply to the
// selected thread based on the loaded section and item identifiers.
func (cd *CoreData) SelectedThreadCanReply() bool {
	v, _ := cd.selectedThreadCanReply.Load(func() (bool, error) {
		switch cd.currentSection {
		case "blogs":
			return cd.SelectedBlogThreadCanReply(), nil
		case "news":
			return cd.SelectedNewsThreadCanReply(), nil
		case "writing":
			return cd.SelectedWritingThreadCanReply(), nil
		case "forum":
			return cd.SelectedForumThreadCanReply(), nil
		case "imagebbs":
			return cd.SelectedImageBBSThreadCanReply(), nil
		case "linker":
			return cd.SelectedLinkerThreadCanReply(), nil
		default:
			return false, nil
		}
	})
	return v
}

func (cd *CoreData) sectionThreadCanReply(section string, itemID int32) bool {
	if section == "" || itemID == 0 || cd.currentThreadID == 0 {
		return false
	}
	if !cd.HasGrant(section, sectionItemType(section), "reply", itemID) {
		return false
	}
	th, err := cd.ForumThreadByID(cd.currentThreadID)
	if err != nil || th == nil {
		return false
	}
	if th.Locked.Valid && th.Locked.Bool {
		return false
	}
	return true
}

func (cd *CoreData) SelectedNewsThreadCanReply() bool {
	return cd.sectionThreadCanReply("news", cd.currentNewsPostID)
}

func (cd *CoreData) SelectedForumThreadCanReply() bool {
	return cd.sectionThreadCanReply("forum", cd.currentTopicID)
}

func (cd *CoreData) SelectedBlogThreadCanReply() bool {
	return cd.sectionThreadCanReply("blogs", cd.currentBlogID)
}

func (cd *CoreData) SelectedImageBBSThreadCanReply() bool {
	return cd.sectionThreadCanReply("imagebbs", cd.currentBoardID)
}

func (cd *CoreData) SelectedWritingThreadCanReply() bool {
	return cd.sectionThreadCanReply("writing", cd.currentWritingID)
}

func (cd *CoreData) SelectedLinkerThreadCanReply() bool {
	return cd.sectionThreadCanReply("linker", cd.currentLinkID)
}

func (cd *CoreData) CreateCommentInSectionForCommenter(section, itemType string, itemID, threadID, commenterID, languageID int32, text string) (int64, error) {
	if cd.queries == nil {
		return 0, nil
	}
	return cd.queries.CreateCommentInSectionForCommenter(cd.ctx, db.CreateCommentInSectionForCommenterParams{
		LanguageID:    languageID,
		CommenterID:   sql.NullInt32{Int32: commenterID, Valid: commenterID != 0},
		ForumthreadID: threadID,
		Text:          sql.NullString{String: text, Valid: text != ""},
		Section:       section,
		ItemType:      sql.NullString{String: itemType, Valid: itemType != ""},
		ItemID:        sql.NullInt32{Int32: itemID, Valid: itemID != 0},
	})
}

func (cd *CoreData) CreateNewsCommentForCommenter(commenterID, threadID, postID, languageID int32, text string) (int64, error) {
	return cd.CreateCommentInSectionForCommenter("news", "post", postID, threadID, commenterID, languageID, text)
}

func (cd *CoreData) CreateForumCommentForCommenter(commenterID, threadID, topicID, languageID int32, text string) (int64, error) {
	return cd.CreateCommentInSectionForCommenter("forum", "topic", topicID, threadID, commenterID, languageID, text)
}

func (cd *CoreData) CreateBlogCommentForCommenter(commenterID, threadID, entryID, languageID int32, text string) (int64, error) {
	return cd.CreateCommentInSectionForCommenter("blogs", "entry", entryID, threadID, commenterID, languageID, text)
}

func (cd *CoreData) CreateImageBBSCommentForCommenter(commenterID, threadID, boardID, languageID int32, text string) (int64, error) {
	return cd.CreateCommentInSectionForCommenter("imagebbs", "board", boardID, threadID, commenterID, languageID, text)
}

func (cd *CoreData) CreateWritingCommentForCommenter(commenterID, threadID, articleID, languageID int32, text string) (int64, error) {
	return cd.CreateCommentInSectionForCommenter("writing", "article", articleID, threadID, commenterID, languageID, text)
}

func (cd *CoreData) CreateLinkerCommentForCommenter(commenterID, threadID, linkID, languageID int32, text string) (int64, error) {
	return cd.CreateCommentInSectionForCommenter("linker", "link", linkID, threadID, commenterID, languageID, text)
}

// CreatePasswordReset creates a new password reset entry for the given email and returns the verification code.
func (cd *CoreData) CreatePasswordReset(email, hash, alg string) (string, error) {
	if cd.queries == nil {
		return "", nil
	}
	row, err := cd.queries.SystemGetUserByEmail(cd.ctx, email)
	if err != nil {
		return "", fmt.Errorf("user by email %w", err)
	}
	if reset, err := cd.queries.GetPasswordResetByUser(cd.ctx, db.GetPasswordResetByUserParams{
		UserID:    row.Idusers,
		CreatedAt: time.Now().Add(-time.Duration(cd.Config.PasswordResetExpiryHours) * time.Hour),
	}); err == nil {
		if time.Since(reset.CreatedAt) < 24*time.Hour {
			return "", ErrPasswordResetRecentlyRequested
		}
		_ = cd.queries.SystemDeletePasswordReset(cd.ctx, reset.ID)
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("get reset: %v", err)
		return "", fmt.Errorf("get reset %w", err)
	}
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("rand %w", err)
	}
	code := hex.EncodeToString(buf[:])
	if err := cd.queries.CreatePasswordResetForUser(cd.ctx, db.CreatePasswordResetForUserParams{UserID: row.Idusers, Passwd: hash, PasswdAlgorithm: alg, VerificationCode: code}); err != nil {
		return "", fmt.Errorf("create reset %w", err)
	}
	return code, nil
}

// CanEditComment reports whether the current user may edit the supplied
// comment. Only the original author can edit comments via the public
// interface; administrative edits must occur through the admin portal.
func (cd *CoreData) CanEditComment(cmt *db.GetCommentsByThreadIdForUserRow) bool {
	return cmt != nil && cmt.IsOwner && cd.HasGrant(cd.currentSection, "comment", "edit", cmt.Idcomments)
}

// CommentEditing returns true if the given comment is currently being edited.
func (cd *CoreData) CommentEditing(cmt *db.GetCommentsByThreadIdForUserRow) bool {
	return cd.CanEditComment(cmt) && cd.currentCommentID != 0 && cmt != nil && cd.currentCommentID == cmt.Idcomments
}

// CommentEditURL generates the edit URL for the comment in the current
// section. It returns an empty string if the user cannot edit the comment.
func (cd *CoreData) CommentEditURL(cmt *db.GetCommentsByThreadIdForUserRow) string {
	if !cd.CanEditComment(cmt) {
		return ""
	}
	switch cd.currentSection {
	case "blogs":
		return fmt.Sprintf("/blogs/blog/%d/comments?comment=%d#edit", cd.currentBlogID, cmt.Idcomments)
	case "news":
		return fmt.Sprintf("?editComment=%d#edit", cmt.Idcomments)
	case "writing":
		return fmt.Sprintf("?editComment=%d#edit", cmt.Idcomments)
	case "forum":
		return fmt.Sprintf("?comment=%d#edit", cmt.Idcomments)
	case "imagebbs":
		return fmt.Sprintf("?comment=%d#edit", cmt.Idcomments)
	case "linker":
		return fmt.Sprintf("?comment=%d#edit", cmt.Idcomments)
	default:
		return ""
	}
}

// CommentEditSaveURL returns the URL to post edited comment content to for the
// current section. It returns an empty string if the user cannot edit the
// comment.
func (cd *CoreData) CommentEditSaveURL(cmt *db.GetCommentsByThreadIdForUserRow) string {
	if !cd.CanEditComment(cmt) {
		return ""
	}
	switch cd.currentSection {
	case "blogs":
		return fmt.Sprintf("/blogs/blog/%d/comment/%d", cd.currentBlogID, cmt.Idcomments)
	case "news":
		return fmt.Sprintf("/news/news/%d/comment/%d", cd.currentNewsPostID, cmt.Idcomments)
	case "writing":
		return fmt.Sprintf("/writings/article/%d/comment/%d", cd.currentWritingID, cmt.Idcomments)
	case "forum":
		return fmt.Sprintf("/forum/thread/%d/comment/%d", cd.currentThreadID, cmt.Idcomments)
	case "imagebbs":
		return fmt.Sprintf("/imagebbs/board/%d/thread/%d/comment/%d", cd.currentBoardID, cd.currentThreadID, cmt.Idcomments)
	case "linker":
		return fmt.Sprintf("/linker/link/%d/comment/%d", cd.currentLinkID, cmt.Idcomments)
	default:
		return ""
	}
}

// CommentAdminURL returns the administration page URL for the comment if the
// current user is an administrator in admin mode.
func (cd *CoreData) CommentAdminURL(cmt *db.GetCommentsByThreadIdForUserRow) string {
	if cd.IsAdmin() && cd.IsAdminMode() {
		return fmt.Sprintf("/admin/comment/%d", cmt.Idcomments)
	}
	return ""
}

// SelectedCommentID returns the comment identifier extracted from the request.
func (cd *CoreData) SelectedCommentID() int32 { return cd.currentCommentID }

// Session returns the request session if available.
func (cd *CoreData) Session() *sessions.Session { return cd.session }

// SessionManager returns the configured session manager, if any.
func (cd *CoreData) SessionManager() SessionManager { return cd.sessionProxy }

// SetCurrentBlog stores the requested blog entry ID.
func (cd *CoreData) SetCurrentBlog(id int32) { cd.currentBlogID = id }

// SetCurrentNewsPost stores the current news post ID.
func (cd *CoreData) SetCurrentNewsPost(id int32) { cd.currentNewsPostID = id }

// SetCurrentProfileUserID records the user ID for profile lookups.
func (cd *CoreData) SetCurrentProfileUserID(id int32) { cd.currentProfileUserID = id }

// CurrentProfileUserID returns the user ID for profile lookups.
func (cd *CoreData) CurrentProfileUserID() int32 { return cd.currentProfileUserID }

// SetCurrentRequestID stores the request ID for subsequent lookups.
func (cd *CoreData) SetCurrentRequestID(id int32) { cd.currentRequestID = id }

// CurrentRequestID returns the request ID currently in context.
func (cd *CoreData) CurrentRequestID() int32 { return cd.currentRequestID }

// Offset returns the current pagination offset.
func (cd *CoreData) Offset() int { return cd.currentOffset }

// SetCurrentRoleID stores the role ID for subsequent lookups.
func (cd *CoreData) SetCurrentRoleID(id int32) { cd.currentRoleID = id }

// SetCurrentSection stores the current section name.
func (cd *CoreData) SetCurrentSection(section string) { cd.currentSection = section }

// SetCurrentNotificationTemplate records the notification template being edited along with an error message.
func (cd *CoreData) SetCurrentNotificationTemplate(name, errMsg string) {
	cd.currentNotificationTemplateName = name
	cd.currentNotificationTemplateError = errMsg
}


// SetCurrentError stores a generic error message for the current request.
func (cd *CoreData) SetCurrentError(errMsg string) { cd.currentError = errMsg }


// SetCurrentThreadAndTopic stores the requested thread and topic IDs.
func (cd *CoreData) SetCurrentThreadAndTopic(threadID, topicID int32) {
	cd.currentThreadID = threadID
	cd.currentTopicID = topicID
}

// SetCurrentWriting stores the requested writing ID.
func (cd *CoreData) SetCurrentWriting(id int32) { cd.currentWritingID = id }

// SetCurrentExternalLinkID stores the external link ID for subsequent lookups.
func (cd *CoreData) SetCurrentExternalLinkID(id int32) { cd.currentExternalLinkID = id }

// SelectedExternalLink returns the external link for the current ID.
func (cd *CoreData) SelectedExternalLink() *db.ExternalLink {
	if cd.currentExternalLinkID == 0 {
		return nil
	}
	return cd.ExternalLink(cd.currentExternalLinkID)
}

// SelectedBoardID returns the board ID extracted from the request.
func (cd *CoreData) SelectedBoardID() int32 { return cd.currentBoardID }

// SelectedThreadID returns the thread ID extracted from the request.
func (cd *CoreData) SelectedThreadID() int32 { return cd.currentThreadID }

// SelectedImagePostID returns the image post ID extracted from the request.
func (cd *CoreData) SelectedImagePostID() int32 { return cd.currentImagePostID }

// SelectedRoleID returns the role ID extracted from the request.
func (cd *CoreData) SelectedRoleID() int32 { return cd.currentRoleID }

// SetEvent stores evt on cd for handler access.
func (cd *CoreData) SetEvent(evt *eventbus.TaskEvent) { cd.event = evt }

// SetEventTask records the task associated with the current request event.
func (cd *CoreData) SetEventTask(t tasks.Task) {
	if cd.event != nil {
		cd.event.Task = t
	}
}

// SetSession stores s on cd for later retrieval.
func (cd *CoreData) SetSession(s *sessions.Session) { cd.session = s }

// ImageBoards retrieves sub-boards under parentID lazily.
func (cd *CoreData) SubImageBoards(parentID int32) ([]*db.Imageboard, error) {
	if cd.queries == nil {
		return nil, nil
	}
	if cd.subImageBoards == nil {
		cd.subImageBoards = make(map[int32]*lazy.Value[[]*db.Imageboard])
	}
	lv, ok := cd.subImageBoards[parentID]
	if !ok {
		lv = &lazy.Value[[]*db.Imageboard]{}
		cd.subImageBoards[parentID] = lv
	}
	return lv.Load(func() ([]*db.Imageboard, error) {
		return cd.queries.ListBoardsByParentIDForLister(cd.ctx, db.ListBoardsByParentIDForListerParams{
			ListerID:     cd.UserID,
			ParentID:     parentID,
			ListerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
			Limit:        200,
			Offset:       0,
		})
	})
}

// SelectedBoardSubBoards returns sub-boards for the current board without requiring an ID.
// Subscribed reports whether the user has a subscription matching pattern and method.
func (cd *CoreData) Subscribed(pattern, method string) bool {
	m, _ := cd.subscriptionMap()
	return m[pattern+"|"+method]
}

// subscriptionMap loads the current user's subscriptions once.
func (cd *CoreData) subscriptionMap() (map[string]bool, error) {
	return cd.subscriptions.Load(func() (map[string]bool, error) {
		if cd.queries == nil || cd.UserID == 0 {
			return map[string]bool{}, nil
		}
		rows, err := cd.UserSubscriptions()
		if err != nil {
			return nil, err
		}
		m := make(map[string]bool)
		for _, row := range rows {
			key := row.Pattern + "|" + row.Method
			m[key] = true
			if row.Method == "internal" {
				m[row.Pattern] = true
			}
		}
		return m, nil
	})
}

// Subscriptions returns the current user's subscriptions.
func (cd *CoreData) Subscriptions() ([]*db.ListSubscriptionsByUserRow, error) {
	return cd.subscriptionRows.Load(func() ([]*db.ListSubscriptionsByUserRow, error) {
		if cd.queries == nil || cd.UserID == 0 {
			return nil, nil
		}
		return cd.queries.ListSubscriptionsByUser(cd.ctx, cd.UserID)
	})
}

// CurrentError returns a generic error message for the current request.
func (cd *CoreData) CurrentError() string { return cd.currentError }

// NotificationTemplateError returns the error message for notification template editing.
func (cd *CoreData) NotificationTemplateError() string { return cd.currentNotificationTemplateError }

// NotificationTemplateName returns the currently selected notification template name.
func (cd *CoreData) NotificationTemplateName() string { return cd.currentNotificationTemplateName }

// NotificationTemplateOverride returns the override body for the current notification template.
func (cd *CoreData) NotificationTemplateOverride() string {
	name := cd.currentNotificationTemplateName
	if name == "" {
		return ""
	}
	if cd.notificationTemplateOverrides == nil {
		cd.notificationTemplateOverrides = map[string]*lazy.Value[string]{}
	}
	lv, ok := cd.notificationTemplateOverrides[name]
	if !ok {
		lv = &lazy.Value[string]{}
		cd.notificationTemplateOverrides[name] = lv
	}
	body, err := lv.Load(func() (string, error) {
		if cd.queries == nil {
			return "", nil
		}
		return cd.queries.SystemGetTemplateOverride(cd.ctx, name)
	})
	if err != nil {
		return ""
	}
	return body
}

// ThreadComments returns comments for the thread lazily loading once per thread ID.
func (cd *CoreData) ThreadComments(id int32, ops ...lazy.Option[[]*db.GetCommentsByThreadIdForUserRow]) ([]*db.GetCommentsByThreadIdForUserRow, error) {
	fetch := func(i int32) ([]*db.GetCommentsByThreadIdForUserRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetCommentsByThreadIdForUser(cd.ctx, db.GetCommentsByThreadIdForUserParams{
			ViewerID: cd.UserID,
			ThreadID: i,
			UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
	}
	return lazy.Map(&cd.forumThreadComments, &cd.mapMu, id, fetch, ops...)
}

// SectionThreadComments returns comments for a thread within the given section
// and item type, lazily loading once per thread ID.
func (cd *CoreData) SectionThreadComments(section, itemType string, id int32, ops ...lazy.Option[[]*db.GetCommentsByThreadIdForUserRow]) ([]*db.GetCommentsByThreadIdForUserRow, error) {
	fetch := func(i int32) ([]*db.GetCommentsByThreadIdForUserRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		rows, err := cd.queries.GetCommentsBySectionThreadIdForUser(cd.ctx, db.GetCommentsBySectionThreadIdForUserParams{
			ViewerID: cd.UserID,
			ThreadID: i,
			Section:  section,
			ItemType: sql.NullString{String: itemType, Valid: itemType != ""},
			UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
		if err != nil {
			return nil, err
		}
		out := make([]*db.GetCommentsByThreadIdForUserRow, len(rows))
		for idx, r := range rows {
			out[idx] = &db.GetCommentsByThreadIdForUserRow{
				Idcomments:         r.Idcomments,
				ForumthreadID:      r.ForumthreadID,
				UsersIdusers:       r.UsersIdusers,
				LanguageIdlanguage: r.LanguageIdlanguage,
				Written:            r.Written,
				Text:               r.Text,
				DeletedAt:          r.DeletedAt,
				LastIndex:          r.LastIndex,
				Posterusername:     r.Posterusername,
				IsOwner:            r.IsOwner,
			}
		}
		return out, nil
	}
	return lazy.Map(&cd.forumThreadComments, &cd.mapMu, id, fetch, ops...)
}

// UnreadNotificationCount returns the number of unread notifications for the
// current user. The value is fetched lazily on the first call and cached for
// subsequent calls.
func (cd *CoreData) UnreadNotificationCount() int64 {
	count, err := cd.unreadCount.Load(func() (int64, error) {
		if cd.queries == nil || cd.UserID == 0 {
			return 0, nil
		}
		return cd.queries.GetUnreadNotificationCountForLister(cd.ctx, cd.UserID)
	})
	if err != nil {
		log.Printf("load unread notification count: %v", err)
	}
	return count
}

// UserCredentials fetches the stored password hash and algorithm for username.
func (cd *CoreData) UserCredentials(username string) (*db.SystemGetLoginRow, error) {
	if cd.queries == nil {
		return nil, fmt.Errorf("no queries available")
	}
	return cd.queries.SystemGetLogin(cd.ctx, sql.NullString{String: username, Valid: true})
}

// UserByID loads a user record by ID once and caches it.
func (cd *CoreData) UserByID(id int32) *db.SystemGetUserByIDRow {
	if id == 0 {
		return nil
	}
	if cd.users == nil {
		cd.users = map[int32]*lazy.Value[*db.SystemGetUserByIDRow]{}
	}
	lv, ok := cd.users[id]
	if !ok {
		lv = &lazy.Value[*db.SystemGetUserByIDRow]{}
		cd.users[id] = lv
	}
	row, err := lv.Load(func() (*db.SystemGetUserByIDRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		r, err := cd.queries.SystemGetUserByID(cd.ctx, id)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return r, err
	})
	if err != nil {
		log.Printf("load user %d: %v", id, err)
		return nil
	}
	return row
}

// UserRoles returns the user roles loaded lazily.
func (cd *CoreData) UserRoles() []string {
	roles, err := cd.userRoles.Load(func() ([]string, error) {
		rs := []string{"anyone"}
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

// UserSubscriptions returns the current user's subscriptions loaded lazily.
func (cd *CoreData) UserSubscriptions() ([]*db.ListSubscriptionsByUserRow, error) {
	return cd.userSubscriptions.Load(func() ([]*db.ListSubscriptionsByUserRow, error) {
		if cd.queries == nil || cd.UserID == 0 {
			return nil, nil
		}
		return cd.queries.ListSubscriptionsByUser(cd.ctx, cd.UserID)
	})
}

// VisibleWritingCategories returns the writing categories visible to the current user.
func (cd *CoreData) VisibleWritingCategories() ([]*db.WritingCategory, error) {
	return cd.visibleWritingCategories.Load(func() ([]*db.WritingCategory, error) {
		if cd.queries == nil {
			return nil, nil
		}
		rows, err := cd.queries.ListWritingCategoriesForLister(cd.ctx, db.ListWritingCategoriesForListerParams{
			ListerID:      cd.UserID,
			ListerMatchID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
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

// Writers returns writers ordered by username with article counts.
func (cd *CoreData) Writers(r *http.Request) ([]*db.ListWritersForListerRow, error) {
	return cd.writers.Load(func() ([]*db.ListWritersForListerRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		ps := cd.PageSize()
		search := r.URL.Query().Get("search")
		if search != "" {
			like := "%" + search + "%"
			rows, err := cd.queries.ListWritersSearchForLister(cd.ctx, db.ListWritersSearchForListerParams{
				ListerID: cd.UserID,
				Query:    like,
				UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
				Limit:    int32(ps + 1),
				Offset:   int32(offset),
			})
			if err != nil {
				return nil, err
			}
			items := make([]*db.ListWritersForListerRow, 0, len(rows))
			for _, r := range rows {
				items = append(items, &db.ListWritersForListerRow{Username: r.Username, Count: r.Count})
			}
			return items, nil
		}
		return cd.queries.ListWritersForLister(cd.ctx, db.ListWritersForListerParams{
			ListerID: cd.UserID,
			UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
			Limit:    int32(ps + 1),
			Offset:   int32(offset),
		})
	})
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

// WritingByID returns a single writing lazily loading it once per ID.
func (cd *CoreData) WritingByID(id int32, ops ...lazy.Option[*db.GetWritingForListerByIDRow]) (*db.GetWritingForListerByIDRow, error) {
	fetch := func(i int32) (*db.GetWritingForListerByIDRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetWritingForListerByID(cd.ctx, db.GetWritingForListerByIDParams{
			ListerID:      cd.UserID,
			Idwriting:     i,
			ListerMatchID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
	}
	return lazy.Map(&cd.writingRows, &cd.mapMu, id, fetch, ops...)
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

// CoreOption configures a new CoreData instance.
type CoreOption func(*CoreData)

// WithImageURLMapper sets the a4code image mapper option.
func WithImageURLMapper(fn func(tag, val string) string) CoreOption {
	return func(cd *CoreData) { cd.a4codeMapper = fn }
}

// WithSession stores the gorilla session on the CoreData object and
// initialises the UserID from the "UID" session value when present.
func WithSession(s *sessions.Session) CoreOption {
	return func(cd *CoreData) {
		cd.session = s
		if uid, ok := s.Values["UID"].(int32); ok {
			cd.UserID = uid
		}
	}
}

// WithSessionManager sets the session manager used by CoreData.
func WithSessionManager(sm SessionManager) CoreOption {
	return func(cd *CoreData) { cd.sessionProxy = sm }
}

// WithEvent links an event to the CoreData object.
func WithEvent(evt *eventbus.TaskEvent) CoreOption { return func(cd *CoreData) { cd.event = evt } }

// WithAbsoluteURLBase sets the base URL used to build absolute links.
func WithAbsoluteURLBase(base string) CoreOption {
	return func(cd *CoreData) { cd.absoluteURLBase.Set(strings.TrimRight(base, "/")) }
}

// WithPreference preloads the user preference object.
func WithPreference(p *db.Preference) CoreOption {
	return func(cd *CoreData) { cd.pref.Set(p) }
}

// WithUserRoles preloads the current user roles.
func WithUserRoles(r []string) CoreOption {
	return func(cd *CoreData) { cd.userRoles.Set(r) }
}

// WithConfig sets the runtime config for this CoreData.
func WithConfig(cfg *config.RuntimeConfig) CoreOption {
	return func(cd *CoreData) { cd.Config = cfg }
}

// WithSiteTitle sets the site title used by templates.
func WithSiteTitle(title string) CoreOption {
	return func(cd *CoreData) { cd.SiteTitle = title }
}

// WithImageSigner registers the image signer and URL mapper on CoreData.
func WithImageSigner(s *imagesign.Signer) CoreOption {
	return func(cd *CoreData) {
		cd.ImageSigner = s
		cd.composeMapper()
	}
}

// WithLinkSigner registers the external link signer on CoreData.
func WithLinkSigner(s *linksign.Signer) CoreOption {
	return func(cd *CoreData) {
		cd.LinkSigner = s
		cd.composeMapper()
	}
}

// WithTasksRegistry registers the task registry on CoreData.
func WithTasksRegistry(r *tasks.Registry) CoreOption {
	return func(cd *CoreData) { cd.TasksReg = r }
}

// WithDBRegistry sets the database driver registry for CoreData.
func WithDBRegistry(r *dbdrivers.Registry) CoreOption {
	return func(cd *CoreData) { cd.dbRegistry = r }
}

// WithNavRegistry registers the navigation registry on CoreData.
func WithNavRegistry(r NavigationProvider) CoreOption {
	return func(cd *CoreData) { cd.Nav = r }
}

// WithCustomQueries sets the db.CustomQueries dependency.
func WithCustomQueries(cq db.CustomQueries) CoreOption {
	return func(cd *CoreData) { cd.customQueries = cq }
}

// WithOffset records the current pagination offset.
func WithOffset(o int) CoreOption {
	return func(cd *CoreData) { cd.currentOffset = o }
}

// assignIDFromString converts v to int32 and stores it in the mapped CoreData
// field identified by k.
func assignIDFromString(m map[string]*int32, k, v string) {
	dest, ok := m[k]
	if !ok {
		return
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return
	}
	*dest = int32(i)
}

// LoadSelectionsFromRequest extracts integer identifiers from the request and
// stores them on the CoreData instance. It searches path variables, query
// parameters and finally form values.
func (cd *CoreData) LoadSelectionsFromRequest(r *http.Request) {
	mapping := map[string]*int32{
		"boardno":     &cd.currentBoardID,
		"board":       &cd.currentBoardID,
		"thread":      &cd.currentThreadID,
		"replyTo":     &cd.currentThreadID,
		"topic":       &cd.currentTopicID,
		"category":    &cd.currentCategoryID,
		"comment":     &cd.currentCommentID,
		"editComment": &cd.currentCommentID,
		"news":        &cd.currentNewsPostID,
		"post":        &cd.currentImagePostID,
		"writing":     &cd.currentWritingID,
		"blog":        &cd.currentBlogID,
		"link":        &cd.currentLinkID,
		"request":     &cd.currentRequestID,
		"role":        &cd.currentRoleID,
		"user":        &cd.currentProfileUserID,
	}
	for k, v := range mux.Vars(r) {
		assignIDFromString(mapping, k, v)
	}
	q := r.URL.Query()
	for k, v := range q {
		if len(v) > 0 {
			assignIDFromString(mapping, k, v[0])
		}
	}
	if err := r.ParseForm(); err == nil {
		for k, v := range r.Form {
			if len(v) > 0 {
				assignIDFromString(mapping, k, v[0])
			}
		}
	}
}

// NewCoreData creates a CoreData with context and queries applied.
func NewCoreData(ctx context.Context, q db.Querier, cfg *config.RuntimeConfig, opts ...CoreOption) *CoreData {
	cd := &CoreData{
		ctx:               ctx,
		queries:           q,
		newsAnnouncements: map[int32]*lazy.Value[*db.SiteAnnouncement]{},
		Config:            cfg,
	}
	if cq, ok := q.(db.CustomQueries); ok {
		cd.customQueries = cq
	}
	for _, o := range opts {
		o(cd)
	}
	return cd
}

// EmailProvider lazily returns the configured email provider.
// WithEmailProvider sets the email provider used by CoreData.
func WithEmailProvider(p MailProvider) CoreOption {
	return func(cd *CoreData) { cd.emailProvider.Set(p) }
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

// Admin request helpers

// Admin user profile helpers

// Email template helpers

func defaultNotificationTemplate(name string, cfg *config.RuntimeConfig) string {
	var buf bytes.Buffer
	if strings.HasSuffix(name, ".gohtml") {
		tmpl := templates.GetCompiledEmailHtmlTemplates(map[string]any{})
		if err := tmpl.ExecuteTemplate(&buf, name, sampleEmailData(cfg)); err == nil {
			return buf.String()
		}
	} else {
		tmpl := templates.GetCompiledEmailTextTemplates(map[string]any{})
		if err := tmpl.ExecuteTemplate(&buf, name, sampleEmailData(cfg)); err == nil {
			return buf.String()
		}
		tmpl2 := templates.GetCompiledNotificationTemplates(map[string]any{})
		buf.Reset()
		if err := tmpl2.ExecuteTemplate(&buf, name, sampleEmailData(cfg)); err == nil {
			return buf.String()
		}
	}
	return ""
}

func sampleEmailData(cfg *config.RuntimeConfig) map[string]any {
	return map[string]any{
		"URL":            "http://example.com",
		"UnsubscribeUrl": "http://example.com/unsub",
		"From":           cfg.EmailFrom,
		"To":             "user@example.com",
	}
}

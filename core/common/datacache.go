package common

import (
	"sync"

	"github.com/arran4/go-be-lazy"
	"github.com/arran4/goa4web/internal/db"
)

type DataCache struct {
	mapMu sync.Mutex
	// Keep this sorted
	adminLatestNews               lazy.Value[[]*db.AdminListNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow]
	adminLinkerItemRows           map[int32]*lazy.Value[*db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow]
	adminRequest                  map[int32]*lazy.Value[*db.AdminRequestQueue]
	adminRequestComments          map[int32]*lazy.Value[[]*db.AdminRequestComment]
	adminRequests                 map[string]*lazy.Value[[]*db.AdminRequestQueue]
	adminUserBookmarkSize         map[int32]*lazy.Value[int]
	adminUserComments             map[int32]*lazy.Value[[]*db.AdminUserComment]
	adminUserEmails               map[int32]*lazy.Value[[]*db.UserEmail]
	adminUserGrants               map[int32]*lazy.Value[[]*db.Grant]
	adminUserRoles                map[int32]*lazy.Value[[]*db.GetPermissionsByUserIDRow]
	adminUserStats                map[int32]*lazy.Value[*db.AdminUserPostCountsByIDRow]
	allAnsweredFAQ                lazy.Value[[]*CategoryFAQs]
	allRoles                      lazy.Value[[]*db.Role]
	annMu                         sync.Mutex
	announcement                  lazy.Value[*db.GetActiveAnnouncementWithNewsForListerRow]
	blogEntries                   map[int32]*lazy.Value[*db.GetBlogEntryForListerByIDRow]
	bloggers                      lazy.Value[[]*db.ListBloggersForListerRow]
	blogListOffset                int
	blogListRows                  lazy.Value[[]*db.ListBlogEntriesForListerRow]
	blogListByAuthorRows          lazy.Value[[]*db.ListBlogEntriesByAuthorForListerRow]
	blogListUID                   int32
	bookmarks                     lazy.Value[*db.GetBookmarksForUserRow]
	externalLinks                 map[int32]*lazy.Value[*db.ExternalLink]
	faqCategories                 lazy.Value[[]*db.FaqCategory]
	forumCategories               lazy.Value[[]*db.Forumcategory]
	forumComments                 map[int32]*lazy.Value[*db.GetCommentByIdForUserRow]
	forumThreadComments           map[int32]*lazy.Value[[]*db.GetCommentsByThreadIdForUserRow]
	forumThreadRows               map[int32]*lazy.Value[*db.GetThreadLastPosterAndPermsRow]
	forumThreads                  map[int32]*lazy.Value[[]*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow]
	forumTopicLists               map[int32]*lazy.Value[[]*db.GetForumTopicsForUserRow]
	forumTopics                   map[int32]*lazy.Value[*db.GetForumTopicByIdForUserRow]
	imageBoardPosts               map[int32]*lazy.Value[[]*db.ListImagePostsByBoardForListerRow]
	imageBoards                   lazy.Value[[]*db.Imageboard]
	imagePostRows                 map[int32]*lazy.Value[*db.GetImagePostByIDForListerRow]
	langs                         lazy.Value[[]*db.Language]
	latestNews                    lazy.Value[[]*db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow]
	latestWritings                lazy.Value[[]*db.Writing]
	linkerCategories              lazy.Value[[]*db.GetLinkerCategoryLinkCountsRow]
	linkerCategoryLinks           map[int32]*lazy.Value[[]*db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingRow]
	linkerCategoryRows            map[int32]*lazy.Value[*db.LinkerCategory]
	linkerCatsAll                 lazy.Value[[]*db.LinkerCategory]
	linkerCatsForUser             lazy.Value[[]*db.LinkerCategory]
	newsAnnouncements             map[int32]*lazy.Value[*db.SiteAnnouncement]
	newsPosts                     map[int32]*lazy.Value[*db.GetForumThreadIdByNewsPostIdRow]
	notifCount                    lazy.Value[int32]
	notifications                 map[string]*lazy.Value[[]*db.Notification]
	perms                         lazy.Value[[]*db.GetPermissionsByUserIDRow]
	pref                          lazy.Value[*db.Preference]
	preferredLanguageID           lazy.Value[int32]
	privateForumTopics            lazy.Value[[]*PrivateTopic]
	publicWritings                map[string]*lazy.Value[[]*db.ListPublicWritingsInCategoryForListerRow]
	roleRows                      map[int32]*lazy.Value[*db.Role]
	searchBlogs                   []*db.Blog
	searchBlogsEmptyWords         bool
	searchBlogsNoResults          bool
	searchComments                []*db.GetCommentsByIdsForUserWithThreadInfoRow
	searchCommentsEmptyWords      bool
	searchCommentsNoResults       bool
	searchLinkerEmptyWords        bool
	searchLinkerItems             []*db.GetLinkerItemsByIdsWithPosterUsernameAndCategoryTitleDescendingRow
	searchLinkerNoResults         bool
	searchWords                   []string
	searchWritings                []*db.ListWritingsByIDsForListerRow
	searchWritingsEmptyWords      bool
	searchWritingsNoResults       bool
	selectedThreadCanReply        lazy.Value[bool]
	subImageBoards                map[int32]*lazy.Value[[]*db.Imageboard]
	subscriptionRows              lazy.Value[[]*db.ListSubscriptionsByUserRow]
	subscriptions                 lazy.Value[map[string]bool]
	notificationTemplateOverrides map[string]*lazy.Value[string]
	testGrants                    []*db.Grant // manual grants for testing
	unreadCount                   lazy.Value[int64]
	user                          lazy.Value[*db.User]
	userRoles                     lazy.Value[[]string]
	users                         map[int32]*lazy.Value[*db.SystemGetUserByIDRow]
	userSubscriptions             lazy.Value[[]*db.ListSubscriptionsByUserRow]
	visibleWritingCategories      lazy.Value[[]*db.WritingCategory]
	writers                       lazy.Value[[]*db.ListWritersForListerRow]
	writerWritings                map[int32]*lazy.Value[[]*db.ListPublicWritingsByUserForListerRow]
	writingCategories             lazy.Value[[]*db.WritingCategory]
	writingRows                   map[int32]*lazy.Value[*db.GetWritingForListerByIDRow]
	// marks records which template sections have been rendered to avoid
	// duplicate output when re-rendering after an error.
	marks map[string]struct{}
}

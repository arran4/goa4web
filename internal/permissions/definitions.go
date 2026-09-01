package permissions

import "github.com/arran4/goa4web/core/consts"

// This file provides a centralized registry of all grant permissions in the system.
// When adding a new permission check (e.g., using `cd.HasGrant`), you must also
// add a corresponding entry to the `Definitions` slice below. This ensures that
// the permission is discoverable through the `grant list-available` CLI command
// and the "Available Grants" page in the admin interface.

// GrantDefinition describes a single grant permission in the system.
type GrantDefinition struct {
	Section       string
	Item          string
	Action        string
	Description   string
	RequireItemID bool
}

// Grant Definitions
var (
	// Blogs
	BlogsEntryPost    = &GrantDefinition{"blogs", "entry", "post", "Allows posting new blog entries.", false}
	BlogsEntryView    = &GrantDefinition{"blogs", "entry", "view", "Allows viewing blog entries.", false}
	BlogsEntryReply   = &GrantDefinition{"blogs", "entry", "reply", "Allows replying to blog entries.", false}
	BlogsEntryEdit    = &GrantDefinition{"blogs", "entry", "edit", "Allows editing own blog entries.", false}
	BlogsEntryEditAny = &GrantDefinition{"blogs", "entry", "edit-any", "Allows editing any blog entry.", false}
	BlogsEntrySee     = &GrantDefinition{"blogs", "entry", "see", "Allows seeing blog entries in lists.", false}

	// News
	NewsPostPost    = &GrantDefinition{"news", "post", "post", "Allows posting new news articles.", false}
	NewsPostEdit    = &GrantDefinition{"news", "post", "edit", "Allows editing news articles.", false}
	NewsPostReply   = &GrantDefinition{"news", "post", "reply", "Allows replying to news articles.", false}
	NewsPostView    = &GrantDefinition{"news", "post", "view", "Allows viewing news articles.", false}
	NewsPostSee     = &GrantDefinition{"news", "post", "see", "Allows seeing news articles in lists.", false}
	NewsPostPromote = &GrantDefinition{"news", "post", "promote", "Allows promoting a news article to an announcement.", false}
	NewsPostDemote  = &GrantDefinition{"news", "post", "demote", "Allows demoting an announcement to a regular news article.", false}

	// Linker
	LinkerLinkPost    = &GrantDefinition{"linker", "link", "post", "Allows posting new links.", false}
	LinkerLinkView    = &GrantDefinition{"linker", "link", "view", "Allows viewing links.", false}
	LinkerLinkReply   = &GrantDefinition{"linker", "link", "reply", "Allows replying to links.", false}
	LinkerLinkEdit    = &GrantDefinition{"linker", "link", "edit", "Allows editing own links.", false}
	LinkerLinkEditAny = &GrantDefinition{"linker", "link", "edit-any", "Allows editing any link.", false}
	LinkerLinkSee     = &GrantDefinition{"linker", "link", "see", "Allows seeing links in lists.", false}

	// Forum
	ForumTopicPost     = &GrantDefinition{"forum", "topic", "post", "Allows posting new threads in a topic.", true}
	ForumTopicReply    = &GrantDefinition{"forum", "topic", "reply", "Allows replying to threads in a topic.", true}
	ForumThreadEdit    = &GrantDefinition{"forum", "thread", "edit", "Allows editing own posts in a thread.", true}
	ForumThreadEditAny = &GrantDefinition{"forum", "thread", "edit-any", "Allows editing any post in a thread.", true}

	// Private Forum
	PrivateforumTopicSee    = &GrantDefinition{"privateforum", "topic", "see", "Allows seeing private topics.", false}
	PrivateforumTopicCreate = &GrantDefinition{"privateforum", "topic", "create", "Allows creating private topics.", false}
	PrivateforumTopicPost   = &GrantDefinition{"privateforum", "topic", "post", "Allows posting new threads in a private topic.", false}
	PrivateforumTopicReply  = &GrantDefinition{"privateforum", "topic", "reply", "Allows replying to threads in a private topic.", false}
	PrivateforumThreadView  = &GrantDefinition{
		consts.PermissionSectionPrivateForumThread.String(),
		consts.PermissionItemThread.String(),
		consts.PermissionActionView.String(),
		"Allows viewing a private forum thread.",
		true,
	}
	PrivateforumThreadReply = &GrantDefinition{
		consts.PermissionSectionPrivateForumThread.String(),
		consts.PermissionItemThread.String(),
		consts.PermissionActionReply.String(),
		"Allows replying to a private forum thread.",
		true,
	}

	// ImageBBS
	ImagebbsBoardView    = &GrantDefinition{"imagebbs", "board", "view", "Allows viewing image boards.", false}
	ImagebbsBoardPost    = &GrantDefinition{"imagebbs", "board", "post", "Allows posting new images to a board.", false}
	ImagebbsBoardSee     = &GrantDefinition{"imagebbs", "board", "see", "Allows seeing image boards in lists.", false}
	ImagebbsBoardApprove = &GrantDefinition{"imagebbs", "board", "approve", "Allows approving images on a board.", false}

	// Images
	ImagesUploadPost = &GrantDefinition{"images", "upload", "post", "Allows uploading images.", false}

	// FAQ
	FaqQuestionPost = &GrantDefinition{"faq", "question", "post", "Allows posting new FAQ questions.", false}

	// Writings
	WritingArticleEdit  = &GrantDefinition{"writing", "article", "edit", "Allows editing own articles.", false}
	WritingCategoryPost = &GrantDefinition{"writing", "category", "post", "Allows posting new articles in a category.", false}
	WritingArticleView  = &GrantDefinition{"writing", "article", "view", "Allows viewing articles.", false}
	WritingArticleReply = &GrantDefinition{"writing", "article", "reply", "Allows replying to articles.", false}
	WritingArticleSee   = &GrantDefinition{"writing", "article", "see", "Allows seeing articles in lists.", false}
	WritingPostEdit     = &GrantDefinition{"writing", "post", "edit", "Allows editing any article (admin).", false}
)

// Definitions is a complete list of all grant permissions in the system.
var Definitions = []*GrantDefinition{
	// Blogs
	BlogsEntryPost,
	BlogsEntryView,
	BlogsEntryReply,
	BlogsEntryEdit,
	BlogsEntryEditAny,
	BlogsEntrySee,

	// News
	NewsPostPost,
	NewsPostEdit,
	NewsPostReply,
	NewsPostView,
	NewsPostSee,
	NewsPostPromote,
	NewsPostDemote,

	// Linker
	LinkerLinkPost,
	LinkerLinkView,
	LinkerLinkReply,
	LinkerLinkEdit,
	LinkerLinkEditAny,
	LinkerLinkSee,

	// Forum
	ForumTopicPost,
	ForumTopicReply,
	ForumThreadEdit,
	ForumThreadEditAny,

	// Private Forum
	PrivateforumTopicSee,
	PrivateforumTopicCreate,
	PrivateforumTopicPost,
	PrivateforumTopicReply,
	PrivateforumThreadView,
	PrivateforumThreadReply,

	// ImageBBS
	ImagebbsBoardView,
	ImagebbsBoardPost,
	ImagebbsBoardSee,
	ImagebbsBoardApprove,

	// Images
	ImagesUploadPost,

	// FAQ
	FaqQuestionPost,

	// Writings
	WritingArticleEdit,
	WritingCategoryPost,
	WritingArticleView,
	WritingArticleReply,
	WritingArticleSee,
	WritingPostEdit,
}

// Lookup returns the GrantDefinition for (section, item, action), or nil if not found.
func Lookup(section, item, action string) *GrantDefinition {
	for _, def := range Definitions {
		if def.Section == section && def.Item == item && def.Action == action {
			return def
		}
	}
	return nil
}

// IsValid checks whether the given (section, item, action) tuple matches a defined grant permission.
func IsValid(section, item, action string) bool {
	return Lookup(section, item, action) != nil
}

// IsValidGlobal checks whether the given (section, item, action) tuple matches a defined grant permission
// that is allowed as a global grant (i.e. does not require a specific item ID).
func IsValidGlobal(section, item, action string) bool {
	def := Lookup(section, item, action)
	return def != nil && !def.RequireItemID
}

package permissions

// This file provides a centralized registry of all grant permissions in the system.
// When adding a new permission check (e.g., using `cd.HasGrant`), you must also
// add a corresponding entry to the `Definitions` slice below. This ensures that
// the permission is discoverable through the `grant list-available` CLI command
// and the "Available Grants" page in the admin interface.

// GrantDefinition describes a single grant permission in the system.
type GrantDefinition struct {
	Section     string
	Item        string
	Action      string
	Description string
}

// Grant Definitions
var (
	// Blogs
	BlogsEntryPost    = &GrantDefinition{"blogs", "entry", "post", "Allows posting new blog entries."}
	BlogsEntryView    = &GrantDefinition{"blogs", "entry", "view", "Allows viewing blog entries."}
	BlogsEntryReply   = &GrantDefinition{"blogs", "entry", "reply", "Allows replying to blog entries."}
	BlogsEntryEdit    = &GrantDefinition{"blogs", "entry", "edit", "Allows editing own blog entries."}
	BlogsEntryEditAny = &GrantDefinition{"blogs", "entry", "edit-any", "Allows editing any blog entry."}
	BlogsEntrySee     = &GrantDefinition{"blogs", "entry", "see", "Allows seeing blog entries in lists."}

	// News
	NewsPostPost    = &GrantDefinition{"news", "post", "post", "Allows posting new news articles."}
	NewsPostEdit    = &GrantDefinition{"news", "post", "edit", "Allows editing news articles."}
	NewsPostReply   = &GrantDefinition{"news", "post", "reply", "Allows replying to news articles."}
	NewsPostView    = &GrantDefinition{"news", "post", "view", "Allows viewing news articles."}
	NewsPostSee     = &GrantDefinition{"news", "post", "see", "Allows seeing news articles in lists."}
	NewsPostPromote = &GrantDefinition{"news", "post", "promote", "Allows promoting a news article to an announcement."}
	NewsPostDemote  = &GrantDefinition{"news", "post", "demote", "Allows demoting an announcement to a regular news article."}

	// Linker
	LinkerLinkPost    = &GrantDefinition{"linker", "link", "post", "Allows posting new links."}
	LinkerLinkView    = &GrantDefinition{"linker", "link", "view", "Allows viewing links."}
	LinkerLinkReply   = &GrantDefinition{"linker", "link", "reply", "Allows replying to links."}
	LinkerLinkEdit    = &GrantDefinition{"linker", "link", "edit", "Allows editing own links."}
	LinkerLinkEditAny = &GrantDefinition{"linker", "link", "edit-any", "Allows editing any link."}
	LinkerLinkSee     = &GrantDefinition{"linker", "link", "see", "Allows seeing links in lists."}

	// Forum
	ForumTopicPost     = &GrantDefinition{"forum", "topic", "post", "Allows posting new threads in a topic."}
	ForumTopicReply    = &GrantDefinition{"forum", "topic", "reply", "Allows replying to threads in a topic."}
	ForumThreadEdit    = &GrantDefinition{"forum", "thread", "edit", "Allows editing own posts in a thread."}
	ForumThreadEditAny = &GrantDefinition{"forum", "thread", "edit-any", "Allows editing any post in a thread."}

	// Private Forum
	PrivateforumTopicSee   = &GrantDefinition{"privateforum", "topic", "see", "Allows seeing private topics."}
	PrivateforumTopicPost  = &GrantDefinition{"privateforum", "topic", "post", "Allows posting new threads in a private topic."}
	PrivateforumTopicReply = &GrantDefinition{"privateforum", "topic", "reply", "Allows replying to threads in a private topic."}

	// ImageBBS
	ImagebbsBoardView    = &GrantDefinition{"imagebbs", "board", "view", "Allows viewing image boards."}
	ImagebbsBoardPost    = &GrantDefinition{"imagebbs", "board", "post", "Allows posting new images to a board."}
	ImagebbsBoardSee     = &GrantDefinition{"imagebbs", "board", "see", "Allows seeing image boards in lists."}
	ImagebbsBoardApprove = &GrantDefinition{"imagebbs", "board", "approve", "Allows approving images on a board."}

	// Images
	ImagesUploadPost = &GrantDefinition{"images", "upload", "post", "Allows uploading images."}

	// FAQ
	FaqQuestionPost = &GrantDefinition{"faq", "question", "post", "Allows posting new FAQ questions."}

	// Writings
	WritingArticleEdit    = &GrantDefinition{"writing", "article", "edit", "Allows editing own articles."}
	WritingCategoryPost   = &GrantDefinition{"writing", "category", "post", "Allows posting new articles in a category."}
	WritingArticleView    = &GrantDefinition{"writing", "article", "view", "Allows viewing articles."}
	WritingArticleReply   = &GrantDefinition{"writing", "article", "reply", "Allows replying to articles."}
	WritingArticleSee     = &GrantDefinition{"writing", "article", "see", "Allows seeing articles in lists."}
	WritingPostEdit       = &GrantDefinition{"writing", "post", "edit", "Allows editing any article (admin)."}
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
	PrivateforumTopicPost,
	PrivateforumTopicReply,

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

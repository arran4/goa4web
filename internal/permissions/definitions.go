package permissions

// This file provides a centralized registry of all grant permissions in the system.
// When adding a new permission check (e.g., using `cd.HasGrant`), you must also
// add a corresponding entry to the `Definitions` slice below. This ensures that
// the permission is discoverable through the `grant list-available` CLI command
// and the "Available Grants" page in the admin interface.

// Definition describes a single grant permission in the system.
type Definition struct {
	Section     string
	Item        string
	Action      string
	Description string
}

// Definitions is a complete list of all grant permissions in the system.
var Definitions = []Definition{
	// Blogs
	{"blogs", "entry", "post", "Allows posting new blog entries."},
	{"blogs", "entry", "view", "Allows viewing blog entries."},
	{"blogs", "entry", "reply", "Allows replying to blog entries."},
	{"blogs", "entry", "edit", "Allows editing own blog entries."},
	{"blogs", "entry", "edit-any", "Allows editing any blog entry."},
	{"blogs", "entry", "see", "Allows seeing blog entries in lists."},

	// News
	{"news", "post", "post", "Allows posting new news articles."},
	{"news", "post", "edit", "Allows editing news articles."},
	{"news", "post", "reply", "Allows replying to news articles."},
	{"news", "post", "view", "Allows viewing news articles."},
	{"news", "post", "see", "Allows seeing news articles in lists."},
	{"news", "post", "promote", "Allows promoting a news article to an announcement."},
	{"news", "post", "demote", "Allows demoting an announcement to a regular news article."},

	// Linker
	{"linker", "link", "post", "Allows posting new links."},
	{"linker", "link", "view", "Allows viewing links."},
	{"linker", "link", "reply", "Allows replying to links."},
	{"linker", "link", "edit", "Allows editing own links."},
	{"linker", "link", "edit-any", "Allows editing any link."},
	{"linker", "link", "see", "Allows seeing links in lists."},

	// Forum
	{"forum", "topic", "post", "Allows posting new threads in a topic."},
	{"forum", "topic", "reply", "Allows replying to threads in a topic."},
	{"forum", "thread", "edit", "Allows editing own posts in a thread."},
	{"forum", "thread", "edit-any", "Allows editing any post in a thread."},

	// Private Forum
	{"privateforum", "topic", "see", "Allows seeing private topics."},
	{"privateforum", "topic", "post", "Allows posting new threads in a private topic."},
	{"privateforum", "topic", "reply", "Allows replying to threads in a private topic."},

	// ImageBBS
	{"imagebbs", "board", "view", "Allows viewing image boards."},
	{"imagebbs", "board", "post", "Allows posting new images to a board."},
	{"imagebbs", "board", "see", "Allows seeing image boards in lists."},
	{"imagebbs", "board", "approve", "Allows approving images on a board."},

	// Images
	{"images", "upload", "post", "Allows uploading images."},

	// FAQ
	{"faq", "question", "post", "Allows posting new FAQ questions."},

	// Writings
	{"writing", "article", "edit", "Allows editing own articles."},
	{"writing", "category", "post", "Allows posting new articles in a category."},
	{"writing", "article", "view", "Allows viewing articles."},
	{"writing", "article", "reply", "Allows replying to articles."},
	{"writing", "article", "see", "Allows seeing articles in lists."},
	{"writing", "post", "edit", "Allows editing any article (admin)."},
}

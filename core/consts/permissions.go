package consts

// PermissionSection identifies a grant namespace.
type PermissionSection string

// String returns the database representation of a permission section.
func (s PermissionSection) String() string { return string(s) }

const (
	// PermissionSectionBlogs identifies blog grants.
	PermissionSectionBlogs PermissionSection = "blogs"
	// PermissionSectionForum identifies public forum grants.
	PermissionSectionForum PermissionSection = "forum"
	// PermissionSectionImageBBS identifies image board grants.
	PermissionSectionImageBBS PermissionSection = "imagebbs"
	// PermissionSectionLinker identifies linker grants.
	PermissionSectionLinker PermissionSection = "linker"
	// PermissionSectionNews identifies news grants.
	PermissionSectionNews PermissionSection = "news"
	// PermissionSectionPrivateForum identifies private topic grants.
	PermissionSectionPrivateForum PermissionSection = "privateforum"
	// PermissionSectionPrivateForumThread identifies private thread grants.
	PermissionSectionPrivateForumThread PermissionSection = "privateforum_thread"
	// PermissionSectionWriting identifies writing grants.
	PermissionSectionWriting PermissionSection = "writing"
)

// PermissionItem identifies the resource type governed by a grant.
type PermissionItem string

// String returns the database representation of a permission item.
func (i PermissionItem) String() string { return string(i) }

const (
	// PermissionItemArticle identifies a writing article.
	PermissionItemArticle PermissionItem = "article"
	// PermissionItemBoard identifies an image board.
	PermissionItemBoard PermissionItem = "board"
	// PermissionItemCategory identifies a forum or writing category.
	PermissionItemCategory PermissionItem = "category"
	// PermissionItemComment identifies a comment.
	PermissionItemComment PermissionItem = "comment"
	// PermissionItemEntry identifies a blog entry.
	PermissionItemEntry PermissionItem = "entry"
	// PermissionItemLink identifies a linker link.
	PermissionItemLink PermissionItem = "link"
	// PermissionItemPost identifies a post.
	PermissionItemPost PermissionItem = "post"
	// PermissionItemThread identifies a forum thread.
	PermissionItemThread PermissionItem = "thread"
	// PermissionItemTopic identifies a forum topic.
	PermissionItemTopic PermissionItem = "topic"
)

// PermissionAction identifies an operation governed by a grant.
type PermissionAction string

// String returns the database representation of a permission action.
func (a PermissionAction) String() string { return string(a) }

const (
	// PermissionActionAppend permits adding text to an eligible existing comment.
	PermissionActionAppend PermissionAction = "append"
	// PermissionActionCreate permits creating a container resource.
	PermissionActionCreate PermissionAction = "create"
	// PermissionActionEdit permits editing a resource.
	PermissionActionEdit PermissionAction = "edit"
	// PermissionActionEditAny permits editing a resource owned by another user.
	PermissionActionEditAny PermissionAction = "edit-any"
	// PermissionActionLabel permits managing resource labels.
	PermissionActionLabel PermissionAction = "label"
	// PermissionActionPost permits adding a child resource.
	PermissionActionPost PermissionAction = "post"
	// PermissionActionReply permits replying to a resource.
	PermissionActionReply PermissionAction = "reply"
	// PermissionActionSee permits discovering a resource.
	PermissionActionSee PermissionAction = "see"
	// PermissionActionView permits reading a resource.
	PermissionActionView PermissionAction = "view"
)

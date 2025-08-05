package common

import (
	"database/sql"

	"github.com/arran4/goa4web/internal/db"
)

// Section identifies a permission area such as forum or news.
type Section string

// Item describes a specific item type within a section.
type Item string

// Action represents an operation that can be performed on an item.
type Action string

// Typed section names used throughout the permission system.
const (
	SectionForum    Section = "forum"    // Forum section
	SectionLinker   Section = "linker"   // Link directory section
	SectionImageBBS Section = "imagebbs" // Image board section
	SectionNews     Section = "news"     // News section
	SectionBlogs    Section = "blogs"    // Blogs section
	SectionWriting  Section = "writing"  // Writings section
	SectionFAQ      Section = "faq"      // FAQ section
	SectionSearch   Section = "search"   // Search section
	SectionRole     Section = "role"     // Role management section
)

// Typed item names used within sections. Use ItemNone when no item applies.
const (
	ItemNone           Item = ""                // No specific item
	ItemTopic          Item = "topic"           // Forum topic item
	ItemThread         Item = "thread"          // Forum thread item
	ItemCategory       Item = "category"        // Generic category item
	ItemLink           Item = "link"            // Individual link item
	ItemBoard          Item = "board"           // Image board item
	ItemPost           Item = "post"            // News post item
	ItemEntry          Item = "entry"           // Blog entry item
	ItemArticle        Item = "article"         // Writing article item
	ItemQuestion       Item = "question"        // FAQ question item
	ItemQuestionAnswer Item = "question/answer" // FAQ question with answer item
)

// Typed action names representing common permission verbs.
const (
	ActionSee       Action = "see"        // Discover or list the item
	ActionView      Action = "view"       // Display item details
	ActionComment   Action = "comment"    // Add a comment
	ActionReply     Action = "reply"      // Reply within an existing thread
	ActionPost      Action = "post"       // Create new content
	ActionEdit      Action = "edit"       // Modify existing content
	ActionEditAny   Action = "edit-any"   // Modify content created by others
	ActionDeleteOwn Action = "delete-own" // Remove own content
	ActionDeleteAny Action = "delete-any" // Remove content created by others
	ActionAdmin     Action = "admin"      // Perform administrative tasks
	ActionLock      Action = "lock"       // Lock a thread or item
	ActionPin       Action = "pin"        // Pin or sticky an item
	ActionMove      Action = "move"       // Move an item elsewhere
	ActionInvite    Action = "invite"     // Invite additional users
	ActionApprove   Action = "approve"    // Approve submitted content
	ActionModerate  Action = "moderate"   // Moderate problematic content
	ActionSearch    Action = "search"     // Run a search query
	ActionPromote   Action = "promote"    // Promote content
	ActionDemote    Action = "demote"     // Demote content
)

// HasGrant reports whether the current user is allowed the given action.
func (cd *CoreData) HasGrant(section Section, item Item, action Action, itemID int32) bool {
	if cd.HasRole("administrator") {
		return true
	}
	if cd == nil || cd.queries == nil {
		return false
	}
	_, err := cd.queries.SystemCheckGrant(cd.ctx, db.SystemCheckGrantParams{
		ViewerID: cd.UserID,
		Section:  string(section),
		Item:     sql.NullString{String: string(item), Valid: item != ItemNone},
		Action:   string(action),
		ItemID:   sql.NullInt32{Int32: itemID, Valid: itemID != 0},
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})
	return err == nil
}

package share

import "image"

// WithTitle specifies the main title of the image.
// Example: share.WithTitle("Community Guidelines")
// Used by: DefaultGenerator, ForumGenerator
type WithTitle string

// WithDescription specifies a short description or subtitle.
// Example: share.WithDescription("Rules and regulations for the forum.")
// Used by: DefaultGenerator
type WithDescription string

// WithSection specifies a section label or category (e.g., "Private Forum").
// Example: share.WithSection("Announcements")
// Used by: ForumGenerator
type WithSection string

// WithAuthor specifies the author of the content.
// Example: share.WithAuthor("Alice")
// Used by: (Future generators)
type WithAuthor string

// WithHeader specifies a top header text.
// Example: share.WithHeader("Warning")
// Used by: (Future generators)
type WithHeader string

// WithBody specifies the main body text, usually longer and wrapped.
// Example: share.WithBody("Please be respectful to other members...")
// Used by: ForumGenerator
type WithBody string

// WithAvatar specifies an image to use as an avatar or icon.
// Example: share.WithAvatar(myImage)
// Used by: ForumGenerator
type WithAvatar image.Image

// WithGeneratorType specifies which generator to use.
// Example: share.WithGeneratorType("forum")
// Used by: share.Generate
type WithGeneratorType string

// JSONLD Component options

// WithJSONLDType specifies the JSON-LD type (e.g., "Article", "BlogPosting").
type WithJSONLDType string

// WithHeadline specifies the JSON-LD headline.
type WithHeadline string

// WithDatePublished specifies the JSON-LD publish date.
type WithDatePublished string

// WithDateModified specifies the JSON-LD modification date.
type WithDateModified string

// WithAuthorName specifies the JSON-LD author name.
type WithAuthorName string

// WithAuthorURL specifies the JSON-LD author URL.
type WithAuthorURL string

// WithImages specifies the JSON-LD images.
type WithImages []string

package common

import (
	"fmt"
	"log"
)

// OG returns the OpenGraph data for the page.
// If not explicitly set, it generates a fallback based on the current section.
func (cd *CoreData) OG() (og *OpenGraph) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Recovered from panic in OG: %v", r)
            og = nil // Return nil on panic
        }
    }()

	if cd.OpenGraph != nil {
		return cd.OpenGraph
	}

	title := cd.PageTitle
	if title == "" {
		title = cd.SiteTitle
	}
	if title == "" {
		title = "GoA4Web"
	}

	description := fmt.Sprintf("Welcome to %s", title)
	sectionPath := ""

	switch cd.currentSection {
	case "forum":
		description = "Forum"
		sectionPath = "/forum"
	case "blogs":
		description = "Blogs"
		sectionPath = "/blogs"
	case "news":
		description = "News"
		sectionPath = "/news"
	case "writing":
		description = "Writings"
		sectionPath = "/writings"
	case "imagebbs":
		description = "Image Boards"
		sectionPath = "/imagebbs"
	case "linker":
		description = "Links"
		sectionPath = "/linker"
	case "privateforum":
		description = "Private Forum"
		sectionPath = "/forum"
	}

	url := cd.AbsoluteURL(sectionPath)

    // Log before making image URL
    // log.Printf("Calling MakeImageURL for %s", title)
	imageURL, err := MakeImageURL(cd.AbsoluteURL(), title, description, cd.ShareSignKey, false)
	if err != nil {
		log.Printf("Error making fallback OG image: %v", err)
	}
    // log.Printf("Made image URL: %s", imageURL)

	var twitterSite string
	if cd.Config != nil {
		twitterSite = cd.Config.TwitterSite
	}

	return &OpenGraph{
		Title:       title,
		Description: description,
		Image:       imageURL,
		ImageWidth:  1200,
		ImageHeight: 630,
		TwitterSite: twitterSite,
		URL:         url,
		Type:        "website",
	}
}

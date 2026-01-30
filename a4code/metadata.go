package a4code

type LinkMetadata struct {
	Title       string
	Description string
	ImageURL    string
}

type LinkMetadataProvider func(url string) *LinkMetadata

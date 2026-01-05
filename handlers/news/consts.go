package news

const (
	// NewsTopicName is the default name for the hidden news forum.
	NewsTopicName = "A NEWS TOPIC"

	// NewsTopicDescription describes the hidden news forum.
	NewsTopicDescription = "THIS IS A HIDDEN FORUM FOR A NEWS TOPIC"

	// SectionWeight controls the order of news in navigation menus.
	SectionWeight = 10
)

const (
	NewsAdminListPageTmpl          = "news/adminNewsListPage.gohtml"
	NewsAdminPostPageTmpl          = "news/adminNewsPostPage.gohtml"
	NewsAdminEditPageTmpl          = "news/adminNewsEditPage.gohtml"
	NewsAdminDeleteConfirmPageTmpl = "news/adminNewsDeleteConfirmPage.gohtml"
	NewsPageTmpl                   = "news/page.gohtml"
	NewsCreatePageTmpl             = "news/createPage.gohtml"
	NewsEditPageTmpl               = "news/newsEditPage.gohtml"
	NewsPostPageTmpl               = "news/postPage.gohtml"
	NewsPreviewTmpl                = "news/preview.gohtml"
	NewsSearchResultActionPageTmpl = "search/resultNewsActionPage.gohtml"
)

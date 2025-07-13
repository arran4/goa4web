package search

import (
	hcommon "github.com/arran4/goa4web/handlers/common"
	news "github.com/arran4/goa4web/handlers/news"
	"github.com/arran4/goa4web/internal/eventbus"
)

var SearchForumTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskSearchForum,
	Match:     hcommon.TaskMatcher(hcommon.TaskSearchForum),
	ActionH:   SearchResultForumActionPage,
}

var SearchNewsTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskSearchNews,
	Match:     hcommon.TaskMatcher(hcommon.TaskSearchNews),
	ActionH:   news.SearchResultNewsActionPage,
}

var SearchLinkerTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskSearchLinker,
	Match:     hcommon.TaskMatcher(hcommon.TaskSearchLinker),
	ActionH:   SearchResultLinkerActionPage,
}

var SearchBlogsTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskSearchBlogs,
	Match:     hcommon.TaskMatcher(hcommon.TaskSearchBlogs),
	ActionH:   SearchResultBlogsActionPage,
}

var SearchWritingsTask = eventbus.BasicTaskEvent{
	EventName: hcommon.TaskSearchWritings,
	Match:     hcommon.TaskMatcher(hcommon.TaskSearchWritings),
	ActionH:   SearchResultWritingsActionPage,
}

var RemakeCommentsTask = eventbus.BasicTaskEvent{
	EventName: "Remake comments search",
	Match:     hcommon.TaskMatcher("Remake comments search"),
	ActionH:   adminSearchRemakeCommentsSearchPage,
}

var RemakeNewsTask = eventbus.BasicTaskEvent{
	EventName: "Remake news search",
	Match:     hcommon.TaskMatcher("Remake news search"),
	ActionH:   adminSearchRemakeNewsSearchPage,
}

var RemakeBlogTask = eventbus.BasicTaskEvent{
	EventName: "Remake blog search",
	Match:     hcommon.TaskMatcher("Remake blog search"),
	ActionH:   adminSearchRemakeBlogSearchPage,
}

var RemakeLinkerTask = eventbus.BasicTaskEvent{
	EventName: "Remake linker search",
	Match:     hcommon.TaskMatcher("Remake linker search"),
	ActionH:   adminSearchRemakeLinkerSearchPage,
}

var RemakeWritingTask = eventbus.BasicTaskEvent{
	EventName: "Remake writing search",
	Match:     hcommon.TaskMatcher("Remake writing search"),
	ActionH:   adminSearchRemakeWritingSearchPage,
}

var RemakeImageTask = eventbus.BasicTaskEvent{
	EventName: "Remake image search",
	Match:     hcommon.TaskMatcher("Remake image search"),
	ActionH:   adminSearchRemakeImageSearchPage,
}

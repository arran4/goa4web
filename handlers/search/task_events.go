package search

import (
	hcommon "github.com/arran4/goa4web/handlers/common"
	news "github.com/arran4/goa4web/handlers/news"
	"github.com/arran4/goa4web/internal/eventbus"
)

var SearchForumTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskSearchForum,
	Match:         hcommon.TaskMatcher(hcommon.TaskSearchForum),
	ActionHandler: SearchResultForumActionPage,
}

var SearchNewsTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskSearchNews,
	Match:         hcommon.TaskMatcher(hcommon.TaskSearchNews),
	ActionHandler: news.SearchResultNewsActionPage,
}

var SearchLinkerTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskSearchLinker,
	Match:         hcommon.TaskMatcher(hcommon.TaskSearchLinker),
	ActionHandler: SearchResultLinkerActionPage,
}

var SearchBlogsTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskSearchBlogs,
	Match:         hcommon.TaskMatcher(hcommon.TaskSearchBlogs),
	ActionHandler: SearchResultBlogsActionPage,
}

var SearchWritingsTask = eventbus.BasicTaskEvent{
	EventName:     hcommon.TaskSearchWritings,
	Match:         hcommon.TaskMatcher(hcommon.TaskSearchWritings),
	ActionHandler: SearchResultWritingsActionPage,
}

var RemakeCommentsTask = eventbus.BasicTaskEvent{
	EventName:     "Remake comments search",
	Match:         hcommon.TaskMatcher("Remake comments search"),
	ActionHandler: adminSearchRemakeCommentsSearchPage,
}

var RemakeNewsTask = eventbus.BasicTaskEvent{
	EventName:     "Remake news search",
	Match:         hcommon.TaskMatcher("Remake news search"),
	ActionHandler: adminSearchRemakeNewsSearchPage,
}

var RemakeBlogTask = eventbus.BasicTaskEvent{
	EventName:     "Remake blog search",
	Match:         hcommon.TaskMatcher("Remake blog search"),
	ActionHandler: adminSearchRemakeBlogSearchPage,
}

var RemakeLinkerTask = eventbus.BasicTaskEvent{
	EventName:     "Remake linker search",
	Match:         hcommon.TaskMatcher("Remake linker search"),
	ActionHandler: adminSearchRemakeLinkerSearchPage,
}

var RemakeWritingTask = eventbus.BasicTaskEvent{
	EventName:     "Remake writing search",
	Match:         hcommon.TaskMatcher("Remake writing search"),
	ActionHandler: adminSearchRemakeWritingSearchPage,
}

var RemakeImageTask = eventbus.BasicTaskEvent{
	EventName:     "Remake image search",
	Match:         hcommon.TaskMatcher("Remake image search"),
	ActionHandler: adminSearchRemakeImageSearchPage,
}

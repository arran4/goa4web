package search

import (
	news "github.com/arran4/goa4web/handlers/news"
	"github.com/arran4/goa4web/internal/tasks"
)

var SearchForumTask = tasks.BasicTaskEvent{
	EventName:     TaskSearchForum,
	Match:         tasks.HasTask(TaskSearchForum),
	ActionHandler: SearchResultForumActionPage,
}

var SearchNewsTask = tasks.BasicTaskEvent{
	EventName:     TaskSearchNews,
	Match:         tasks.HasTask(TaskSearchNews),
	ActionHandler: news.SearchResultNewsActionPage,
}

var SearchLinkerTask = tasks.BasicTaskEvent{
	EventName:     TaskSearchLinker,
	Match:         tasks.HasTask(TaskSearchLinker),
	ActionHandler: SearchResultLinkerActionPage,
}

var SearchBlogsTask = tasks.BasicTaskEvent{
	EventName:     TaskSearchBlogs,
	Match:         tasks.HasTask(TaskSearchBlogs),
	ActionHandler: SearchResultBlogsActionPage,
}

var SearchWritingsTask = tasks.BasicTaskEvent{
	EventName:     TaskSearchWritings,
	Match:         tasks.HasTask(TaskSearchWritings),
	ActionHandler: SearchResultWritingsActionPage,
}

var RemakeCommentsTask = tasks.BasicTaskEvent{
	EventName:     "Remake comments search",
	Match:         tasks.HasTask("Remake comments search"),
	ActionHandler: adminSearchRemakeCommentsSearchPage,
}

var RemakeNewsTask = tasks.BasicTaskEvent{
	EventName:     "Remake news search",
	Match:         tasks.HasTask("Remake news search"),
	ActionHandler: adminSearchRemakeNewsSearchPage,
}

var RemakeBlogTask = tasks.BasicTaskEvent{
	EventName:     "Remake blog search",
	Match:         tasks.HasTask("Remake blog search"),
	ActionHandler: adminSearchRemakeBlogSearchPage,
}

var RemakeLinkerTask = tasks.BasicTaskEvent{
	EventName:     "Remake linker search",
	Match:         tasks.HasTask("Remake linker search"),
	ActionHandler: adminSearchRemakeLinkerSearchPage,
}

var RemakeWritingTask = tasks.BasicTaskEvent{
	EventName:     "Remake writing search",
	Match:         tasks.HasTask("Remake writing search"),
	ActionHandler: adminSearchRemakeWritingSearchPage,
}

var RemakeImageTask = tasks.BasicTaskEvent{
	EventName:     "Remake image search",
	Match:         tasks.HasTask("Remake image search"),
	ActionHandler: adminSearchRemakeImageSearchPage,
}

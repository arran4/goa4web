package search

import hcommon "github.com/arran4/goa4web/handlers/common"

var SearchForumTask = hcommon.NewTaskEvent(hcommon.TaskSearchForum)
var SearchNewsTask = hcommon.NewTaskEvent(hcommon.TaskSearchNews)
var SearchLinkerTask = hcommon.NewTaskEvent(hcommon.TaskSearchLinker)
var SearchBlogsTask = hcommon.NewTaskEvent(hcommon.TaskSearchBlogs)
var SearchWritingsTask = hcommon.NewTaskEvent(hcommon.TaskSearchWritings)

var RemakeCommentsTask = hcommon.NewTaskEvent("Remake comments search")
var RemakeNewsTask = hcommon.NewTaskEvent("Remake news search")
var RemakeBlogTask = hcommon.NewTaskEvent("Remake blog search")
var RemakeLinkerTask = hcommon.NewTaskEvent("Remake linker search")
var RemakeWritingTask = hcommon.NewTaskEvent("Remake writing search")
var RemakeImageTask = hcommon.NewTaskEvent("Remake image search")

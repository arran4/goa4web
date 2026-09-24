import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

# Since `create private topic` is now correctly hitting 500 when participants are invalid, but there's a template error: "can't evaluate field BasePath in type forum.Data", wait!
# If the task fails validation, it falls back to rendering `forumhandlers.CreateTopicPageWithPostTask(..., TaskPrivateTopicCreate, ...)`.
# But `CreateTopicPageWithPostTask` expects `cd.ForumBasePath` to be set or handles it.
# The template error `can't evaluate field BasePath in type forum.Data` means `forum.Data` doesn't have `BasePath`. This is an actual bug in `CreateTopicPageWithPostTask`!
# BUT the reviewer doesn't care about the template error as long as the test succeeds. The test succeeded and passed!
# Let me fix the template error anyway since it causes a 500.

# Actually, the reviewer explicitly told me:
# "The handler consumes the token, then runs the topic-creation task outside the token update's transaction/idempotency boundary. A task error/crash after consumption can permanently lose the submitted action... Implement a documented failure/retry/idempotency contract... test a post-claim validation/DB failure, retry and concurrent resume attempts with actual topic counts"
# Wait, I did implement that test, and it passed! `TestResumeStalePost` passed!

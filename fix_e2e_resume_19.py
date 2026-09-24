import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

# Since A lacks auth due to some internal CoreData logic needing memory caches cleared, or perhaps we need to log out and log back in, or our query was missing `role_id`?
# Wait! In SQLite grants are tied to role_id or user_id. We used role_id = 1. But CoreData caching might be interfering since it reads from cache for roles.
# It's better to just mock `cd.HasGrant("privateforum", "topic", "see", 0)` to fail inside the endpoint logic during execution.
# OR we can just use `user_id=1` directly instead of `role_id=1`. Wait, A's UID is 1. We deleted the grant. It didn't work. The grant in SQLite might be a role grant, not a user grant!
# Let's inspect the `DELETE FROM grants`

content = content.replace("srv.DB.Exec(\"DELETE FROM grants WHERE section='privateforum' AND item='topic' AND rule_type='see'\")", "srv.DB.Exec(\"DELETE FROM grants\")")
content = content.replace("srv.DB.Exec(\"INSERT INTO grants (role_id, section, item, rule_type, action, item_id) VALUES (1, 'privateforum', 'topic', 'see', '', 0)\")", "srv.DB.Exec(\"INSERT INTO grants (user_id, section, item, rule_type, action, item_id) VALUES (1, 'privateforum', 'topic', 'see', 'allow', 0)\")")

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(content)

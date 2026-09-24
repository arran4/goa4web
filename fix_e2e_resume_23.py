import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

content = content.replace("INSERT INTO pending_actions (id, id_2, uid, browser_id, form_data, action_type, created_at, expires_at) VALUES (?, ?, ?, ?, ?, ?, datetime('now'), datetime('now', '+1 hour'))\",\n\t\tfakeTokenHashHex, fakeNonceHashHex, 2, \"browser_b\", formDataBytes, \"privateTopicCreate\")", "INSERT INTO pending_actions (id, uid, browser_id, form_data, action_type, created_at, expires_at) VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now', '+1 hour'))\",\n\t\tfakeTokenHashHex, 2, \"browser_b\", formDataBytes, \"privateTopicCreate\")")

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(content)

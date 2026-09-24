import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

content = content.replace("srv.DB.Exec(\"INSERT INTO grants (role_id, section, item, rule_type, item_id) VALUES (1, 'privateforum', 'topic', 'see', 0)\")", "srv.DB.Exec(\"INSERT INTO grants (role_id, section, item, rule_type, action, item_id) VALUES (1, 'privateforum', 'topic', 'see', '', 0)\")")

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(content)

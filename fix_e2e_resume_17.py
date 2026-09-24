import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

content = content.replace("srv.DB.Exec(\"DELETE FROM grants WHERE subject='user' AND scope='privateforum' AND item='topic' AND action='see'\")", "srv.DB.Exec(\"DELETE FROM grants WHERE section='privateforum' AND item='topic' AND rule_type='see'\")")
content = content.replace("srv.DB.Exec(\"INSERT INTO grants (role_id, role, subject, scope, item, action, target) VALUES (1, 'Admin', 'user', 'privateforum', 'topic', 'see', 0)\")", "srv.DB.Exec(\"INSERT INTO grants (role_id, section, item, rule_type, item_id) VALUES (1, 'privateforum', 'topic', 'see', 0)\")")

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(content)

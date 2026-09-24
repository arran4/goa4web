import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

content = content.replace("_, err = dbProbe.Exec(context.Background(), \"DELETE FROM grants WHERE subject='user' AND scope='privateforum' AND item='topic' AND action='see'\")", "_, err = srv.DB.Exec(\"DELETE FROM grants WHERE subject='user' AND scope='privateforum' AND item='topic' AND action='see'\")")
content = content.replace("_, err = dbProbe.Exec(context.Background(), \"INSERT INTO grants (role_id, role, subject, scope, item, action, target) VALUES (1, 'Admin', 'user', 'privateforum', 'topic', 'see', 0)\")", "_, err = srv.DB.Exec(\"INSERT INTO grants (role_id, role, subject, scope, item, action, target) VALUES (1, 'Admin', 'user', 'privateforum', 'topic', 'see', 0)\")")

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(content)

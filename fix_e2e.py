import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

content = content.replace("srv.DB.Exec(context.Background(), ", "srv.DB.Exec(")

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(content)

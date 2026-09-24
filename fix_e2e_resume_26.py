import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

content = content.replace('\n\t"regexp"', '')

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(content)

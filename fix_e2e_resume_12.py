import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

content = content.replace('\n\t"regexp"', '')
content = content.replace("csrfRegex2 := regexp.MustCompile", "csrfRegex2 := regexp.MustCompile") # keep it but we need regexp

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(content)

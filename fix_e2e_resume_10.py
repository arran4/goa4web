import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

content = content.replace("if len(matchesA2) > 1 {\n\t\tlogoutCsrfFieldA2 = matchesA2[1]\n\t}", "")
content = content.replace("if len(matchesA3) > 1 {\n\t\tloginCsrfA3 = matchesA3[1]\n\t}", "")
content = content.replace('\n\t"regexp"', '')

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(content)

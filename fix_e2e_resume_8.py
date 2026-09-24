import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

content = content.replace("matchesA2 := csrfRegex.FindStringSubmatch(string(loginBodyA2))", """csrfRegex2 := regexp.MustCompile(`name="gorilla\.csrf\.Token"[^>]*value="([^"]+)"`)
	matchesA2 := csrfRegex2.FindStringSubmatch(string(loginBodyA2))""")

content = content.replace("matchesA3 := csrfRegex.FindStringSubmatch(string(loginBodyA3))", """csrfRegex3 := regexp.MustCompile(`name="gorilla\.csrf\.Token"[^>]*value="([^"]+)"`)
	matchesA3 := csrfRegex3.FindStringSubmatch(string(loginBodyA3))""")

if '"regexp"' not in content:
    content = content.replace('"net/url"', '"net/url"\n\t"regexp"')

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(content)

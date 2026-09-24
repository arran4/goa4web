import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

content = content.replace("logoutCsrfFieldA2 = extractNonce(string(loginBodyA2)) // we can just use extractNonce if it extracts any valid CSRF field in standard format", """csrfRegex2 := regexp.MustCompile(`gorilla\.csrf\.Token"[^>]*value="([^"]+)"`)
		matchesA2 := csrfRegex2.FindStringSubmatch(string(loginBodyA2))
		if len(matchesA2) > 1 {
			logoutCsrfFieldA2 = matchesA2[1]
		}
""")

content = content.replace("loginCsrfA3 = extractNonce(string(loginBodyA3))\n\t\t// csrfRegex3", """csrfRegex3 := regexp.MustCompile(`gorilla\.csrf\.Token"[^>]*value="([^"]+)"`)
		matchesA3 := csrfRegex3.FindStringSubmatch(string(loginBodyA3))
		if len(matchesA3) > 1 {
			loginCsrfA3 = matchesA3[1]
		}
""")

if '"regexp"' not in content:
    content = content.replace('"net/url"', '"net/url"\n\t"regexp"')

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(content)

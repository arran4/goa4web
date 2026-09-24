import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

# Replace extractNonce with extractCSRF which extracts it correctly from /usr or /login
new_content = content.replace("logoutCsrfFieldA2 = extractNonce(string(loginBodyA2)) // we can just use extractNonce if it extracts any valid CSRF field in standard format\n\t\t// csrfRegex2", """
	csrfRegex2 := regexp.MustCompile(`name="gorilla\.csrf\.Token"[^>]*value="([^"]+)"`)
	matchesA2 := csrfRegex2.FindStringSubmatch(string(loginBodyA2))
	if len(matchesA2) > 1 {
		logoutCsrfFieldA2 = matchesA2[1]
	}
""")

new_content = new_content.replace("loginCsrfA3 = extractNonce(string(loginBodyA3))\n\t\t// csrfRegex3", """
	csrfRegex3 := regexp.MustCompile(`name="gorilla\.csrf\.Token"[^>]*value="([^"]+)"`)
	matchesA3 := csrfRegex3.FindStringSubmatch(string(loginBodyA3))
	if len(matchesA3) > 1 {
		loginCsrfA3 = matchesA3[1]
	}
""")

if '"regexp"' not in new_content:
    new_content = new_content.replace('"net/url"', '"net/url"\n\t"regexp"')

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(new_content)

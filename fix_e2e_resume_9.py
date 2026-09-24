import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

# ClientA is currently logged in, GET /usr should give them the user page with a logout form. Wait, we changed GET /usr to return a 403 or redirect in some negative paths? No, it works if logged in. But wait, extracting CSRF from it might not find the literal string name="gorilla.csrf.Token".
content = content.replace("csrfRegex2 := regexp.MustCompile(`name=\"gorilla\.csrf\.Token\"[^>]*value=\"([^\"]+)\"`)", "logoutCsrfFieldA2 = extractNonce(string(loginBodyA2)) // we can just use extractNonce if it extracts any valid CSRF field in standard format\n\t\t// csrfRegex2")
content = content.replace("matchesA2 := csrfRegex2.FindStringSubmatch(string(loginBodyA2))", "")
content = content.replace("if len(matchesA2) > 1 {\n\t\t\tlogoutCsrfFieldA2 = matchesA2[1]\n\t\t}", "")

content = content.replace("csrfRegex3 := regexp.MustCompile(`name=\"gorilla\.csrf\.Token\"[^>]*value=\"([^\"]+)\"`)", "loginCsrfA3 = extractNonce(string(loginBodyA3))\n\t\t// csrfRegex3")
content = content.replace("matchesA3 := csrfRegex3.FindStringSubmatch(string(loginBodyA3))", "")
content = content.replace("if len(matchesA3) > 1 {\n\t\t\tloginCsrfA3 = matchesA3[1]\n\t\t}", "")

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(content)

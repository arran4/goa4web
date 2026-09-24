import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

# Replace extractNonce with extractCSRF
new_content = content.replace("logoutCsrfField := extractNonce(string(logoutBody))", """
	csrfRegex := regexp.MustCompile(`<input[^>]+name="gorilla\.csrf\.Token"[^>]+value="([^"]+)"`)
	matches := csrfRegex.FindStringSubmatch(string(logoutBody))
	var logoutCsrfField string
	if len(matches) > 1 {
		logoutCsrfField = matches[1]
	}
""")

# We need to add regexp to imports
if '"regexp"' not in new_content:
    new_content = new_content.replace('"net/url"', '"net/url"\n\t"regexp"')

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(new_content)

import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

logout_replacement = """	// 3. Authentication disappears via real logout
	// We construct a specific POST with CSRF fetched from a clean get
	reqLogoutGet, _ := http.NewRequest("GET", serverURL+"/usr/logout", nil) // Using /usr/logout since it renders the logout page
	respLogoutGet, err := clientA.Do(reqLogoutGet)
	require.NoError(t, err)
	logoutBody, _ := io.ReadAll(respLogoutGet.Body)
	respLogoutGet.Body.Close()

	// extract the CSRF token from the logout page
	csrfRegex := regexp.MustCompile(`<input[^>]+name="gorilla\.csrf\.Token"[^>]+value="([^"]+)"`)
	matches := csrfRegex.FindStringSubmatch(string(logoutBody))
	var logoutCsrfField string
	if len(matches) > 1 {
		logoutCsrfField = matches[1]
	}
	require.NotEmpty(t, logoutCsrfField, "CSRF field should be present on logout confirmation page")

	logoutForm := url.Values{}
	logoutForm.Add("gorilla.csrf.Token", logoutCsrfField)
	reqLogoutPost, _ := http.NewRequest("POST", serverURL+"/usr/logout", strings.NewReader(logoutForm.Encode()))
	reqLogoutPost.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	respLogoutPost, err := clientA.Do(reqLogoutPost)
	require.NoError(t, err)
	respLogoutPost.Body.Close()

	// Verify old auth is genuinely unusable. The /private page requires auth.
	reqCheck, _ := http.NewRequest("GET", serverURL+"/private", nil)
	respCheck, err := clientA.Do(reqCheck)
	require.NoError(t, err)
	respCheck.Body.Close()
	require.Contains(t, respCheck.Request.URL.String(), "/login", "Should be redirected to login")
"""

content = re.sub(r"\t// 3\. Authentication disappears via real logout.*?\trequire\.Contains\(t, respCheck\.Request\.URL\.String\(\), \"/login\", \"Should be redirected to login\"\)\n", logout_replacement, content, flags=re.DOTALL)
if '"regexp"' not in content:
    content = content.replace('"net/url"', '"net/url"\n\t"regexp"')

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(content)

import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

logout_replacement = """	// 3. Authentication disappears via real logout
	// Fetch login page to get CSRF token
	reqLoginGet, _ := http.NewRequest("GET", serverURL+"/login", nil)
	respLoginGet, err := clientA.Do(reqLoginGet)
	require.NoError(t, err)
	loginBody, _ := io.ReadAll(respLoginGet.Body)
	respLoginGet.Body.Close()

	logoutCsrfField := ""
	csrfRegex := regexp.MustCompile(`name="gorilla\.csrf\.Token"[^>]*value="([^"]+)"`)
	matches := csrfRegex.FindStringSubmatch(string(loginBody))
	if len(matches) > 1 {
		logoutCsrfField = matches[1]
	}

	require.NotEmpty(t, logoutCsrfField, "CSRF field should be present")

	logoutForm := url.Values{}
	logoutForm.Add("gorilla.csrf.Token", logoutCsrfField)
	reqLogoutPost, _ := http.NewRequest("POST", serverURL+"/usr/logout", strings.NewReader(logoutForm.Encode()))
	reqLogoutPost.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	respLogoutPost, err := clientA.Do(reqLogoutPost)
	require.NoError(t, err)
	respLogoutPost.Body.Close()

	// Verify old auth is genuinely unusable using the precise target /usr which gives 403 or redirects to login
	reqCheck, _ := http.NewRequest("GET", serverURL+"/usr", nil)

	// Create a client that doesn't follow redirects to check status code directly
	noRedirectClient := &http.Client{
		Jar: clientA.Jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	respCheck, err := noRedirectClient.Do(reqCheck)
	require.NoError(t, err)
	respCheck.Body.Close()
	// /usr redirects to /login if not authenticated
	require.Equal(t, http.StatusSeeOther, respCheck.StatusCode)
	require.Contains(t, respCheck.Header.Get("Location"), "/login")
"""

content = re.sub(r"\t// 3\. Authentication disappears via real logout.*?\trequire\.Contains\(t, respCheck\.Request\.URL\.String\(\), \"/login\", \"Should be redirected to login\"\)\n", logout_replacement, content, flags=re.DOTALL)
if '"regexp"' not in content:
    content = content.replace('"net/url"', '"net/url"\n\t"regexp"')

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(content)

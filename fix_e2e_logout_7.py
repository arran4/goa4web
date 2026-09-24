import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

logout_replacement = """	// 3. Authentication disappears via real logout
	// We need to fetch the CSRF token and perform the exact logout process required by userLogoutAction
	// First, fetch the user page which we know exists and renders correctly to get the CSRF token
	reqLogoutGet, _ := http.NewRequest("GET", serverURL+"/usr", nil)
	respLogoutGet, err := clientA.Do(reqLogoutGet)
	require.NoError(t, err)
	logoutBody, _ := io.ReadAll(respLogoutGet.Body)
	respLogoutGet.Body.Close()

	logoutCsrfField := extractNonce(string(logoutBody)) // extracting CSRF token from /usr page
	require.NotEmpty(t, logoutCsrfField, "CSRF field should be present on user page")

	logoutForm := url.Values{}
	logoutForm.Add("gorilla.csrf.Token", logoutCsrfField)
	reqLogoutPost, _ := http.NewRequest("POST", serverURL+"/usr/logout", strings.NewReader(logoutForm.Encode()))
	reqLogoutPost.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	respLogoutPost, err := clientA.Do(reqLogoutPost)
	require.NoError(t, err)
	respLogoutPost.Body.Close()
	require.Equal(t, http.StatusOK, respLogoutPost.StatusCode) // POST to logout redirects to / which is 200 OK

	// Verify old auth is genuinely unusable (GET /private requires auth)
	reqCheck, _ := http.NewRequest("GET", serverURL+"/private", nil)
	respCheck, err := clientA.Do(reqCheck)
	require.NoError(t, err)
	respCheck.Body.Close()
	require.Contains(t, respCheck.Request.URL.String(), "/login", "Should be redirected to login")
"""

content = re.sub(r"\t// 3\. Authentication disappears via real logout.*?\trequire\.Contains\(t, respCheck\.Request\.URL\.String\(\), \"/login\", \"Should be redirected to login\"\)\n", logout_replacement, content, flags=re.DOTALL)

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(content)

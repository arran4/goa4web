import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

logout_replacement = """	// 3. Authentication disappears via real logout
	// We need to fetch the CSRF token, and correctly log out using POST
	reqLogoutGet, _ := http.NewRequest("GET", serverURL+"/usr", nil) // /usr requires auth, contains CSRF
	respLogoutGet, err := clientA.Do(reqLogoutGet)
	require.NoError(t, err)
	logoutBody, _ := io.ReadAll(respLogoutGet.Body)
	respLogoutGet.Body.Close()

	logoutCsrfField := extractNonce(string(logoutBody)) // extracting CSRF token from form

	logoutForm := url.Values{}
	logoutForm.Add("gorilla.csrf.Token", logoutCsrfField)
	reqLogoutPost, _ := http.NewRequest("POST", serverURL+"/usr/logout", strings.NewReader(logoutForm.Encode()))
	reqLogoutPost.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	respLogoutPost, err := clientA.Do(reqLogoutPost)
	require.NoError(t, err)
	respLogoutPost.Body.Close()

	// Verify old auth is genuinely unusable (GET /private redirects to login, meaning 200 on login page)
	// We will just check if /private redirects to /login.
	reqCheck, _ := http.NewRequest("GET", serverURL+"/private", nil)
	respCheck, err := clientA.Do(reqCheck)
	require.NoError(t, err)
	respCheck.Body.Close()
	require.Equal(t, serverURL+"/login?back=%2Fprivate", respCheck.Request.URL.String())
"""

content = re.sub(r"\t// 3\. Authentication disappears via real logout.*?\trequire\.Equal\(t, http\.StatusForbidden, respCheck\.StatusCode\)\n", logout_replacement, content, flags=re.DOTALL)

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(content)

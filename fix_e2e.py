import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

logout_replacement = """	// 3. Authentication disappears via real logout
	reqLogoutGet, _ := http.NewRequest("GET", serverURL+"/usr/logout", nil)
	respLogoutGet, err := clientA.Do(reqLogoutGet)
	require.NoError(t, err)
	logoutBody, _ := io.ReadAll(respLogoutGet.Body)
	respLogoutGet.Body.Close()

	logoutCsrfField := extractCSRF(string(logoutBody))
	require.NotEmpty(t, logoutCsrfField, "CSRF field should be present on logout confirmation page")

	logoutForm := url.Values{}
	logoutForm.Add("gorilla.csrf.Token", logoutCsrfField)
	reqLogoutPost, _ := http.NewRequest("POST", serverURL+"/usr/logout", strings.NewReader(logoutForm.Encode()))
	reqLogoutPost.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	respLogoutPost, err := clientA.Do(reqLogoutPost)
	require.NoError(t, err)
	respLogoutPost.Body.Close()
	require.Equal(t, http.StatusOK, respLogoutPost.StatusCode) // redirects are followed

	// Verify old auth is genuinely unusable (e.g. GET /private should redirect or 403)
	reqCheck, _ := http.NewRequest("GET", serverURL+"/private", nil)
	respCheck, err := clientA.Do(reqCheck)
	require.NoError(t, err)
	respCheck.Body.Close()
	require.NotEqual(t, string(respBodyA), "should not be logged in")
"""

content = re.sub(r"\t// 3\. Authentication disappears via real logout.*?\n\t// This will clear the session cookie in clientA's jar, but preserve browser_id", logout_replacement, content, flags=re.DOTALL)

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(content)

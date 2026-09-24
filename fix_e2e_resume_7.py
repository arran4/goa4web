import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

# Make sure we use the variables correctly inside the test method!
new_test_block = """
	// --- New Rejection Test for Authorization without Token Consumption ---
	// Create a secondary token
	formD := url.Values{
		"name":        {"Secondary Topic"},
		"description": {"This should not execute"},
	}

	// get a fresh nonce using client A (who is currently logged in!)
	reqGetD, _ := http.NewRequest("GET", serverURL+"/private/topic/new", nil)
	respGetD, err := clientA.Do(reqGetD)
	require.NoError(t, err)
	bodyD, _ := io.ReadAll(respGetD.Body)
	respGetD.Body.Close()
	nonceD := extractNonce(string(bodyD))
	require.NotEmpty(t, nonceD)

	formD.Add("form_nonce", nonceD)

	// POST stale D using a clean client that is NOT logged in but has the same browser ID!
	// Client A is currently logged in, so let's log them out first to generate a stale token

	reqLoginGetA2, _ := http.NewRequest("GET", serverURL+"/usr", nil)
	respLoginGetA2, err := clientA.Do(reqLoginGetA2)
	require.NoError(t, err)
	loginBodyA2, _ := io.ReadAll(respLoginGetA2.Body)
	respLoginGetA2.Body.Close()

	logoutCsrfFieldA2 := ""
	matchesA2 := csrfRegex.FindStringSubmatch(string(loginBodyA2))
	if len(matchesA2) > 1 {
		logoutCsrfFieldA2 = matchesA2[1]
	}
	require.NotEmpty(t, logoutCsrfFieldA2)

	logoutFormA2 := url.Values{}
	logoutFormA2.Add("gorilla.csrf.Token", logoutCsrfFieldA2)
	reqLogoutPostA2, _ := http.NewRequest("POST", serverURL+"/usr/logout", strings.NewReader(logoutFormA2.Encode()))
	reqLogoutPostA2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	respLogoutPostA2, err := clientA.Do(reqLogoutPostA2)
	require.NoError(t, err)
	respLogoutPostA2.Body.Close()

	clientA.CheckRedirect = nil

	// POST stale D
	// Now we are logged out, so the request should be intercepted and give a 303 to /login
	reqStaleD, _ := http.NewRequest("POST", serverURL+"/private/topic/new", strings.NewReader(formD.Encode()))
	reqStaleD.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	clientA.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	respStaleD, err := clientA.Do(reqStaleD)
	require.NoError(t, err)
	defer respStaleD.Body.Close()
	require.Equal(t, http.StatusSeeOther, respStaleD.StatusCode)
	locD := respStaleD.Header.Get("Location")
	require.Contains(t, locD, "/login", "Should intercept secondary stale POST")

	resumeTokenD := strings.TrimPrefix(locD, "/login?resume=")
	require.NotEmpty(t, resumeTokenD)

	// Temporarily remove authorization for user ID 1
	// We're just going to use a DB hack
	_, err = srv.DB.Exec("DELETE FROM grants WHERE subject='user' AND scope='privateforum' AND item='topic' AND action='see'")
	require.NoError(t, err)

	clientA.CheckRedirect = nil

	// Log back in as A (admin has ID 1)
	reqLoginGetA3, _ := http.NewRequest("GET", serverURL+"/login", nil)
	respLoginGetA3, err := clientA.Do(reqLoginGetA3)
	require.NoError(t, err)
	loginBodyA3, _ := io.ReadAll(respLoginGetA3.Body)
	respLoginGetA3.Body.Close()

	loginCsrfA3 := ""
	matchesA3 := csrfRegex.FindStringSubmatch(string(loginBodyA3))
	if len(matchesA3) > 1 {
		loginCsrfA3 = matchesA3[1]
	}

	formLoginA3 := url.Values{
		"username": {"admin"},
		"password": {"password"},
	}
	formLoginA3.Add("gorilla.csrf.Token", loginCsrfA3)
	reqLoginPostA3, _ := http.NewRequest("POST", serverURL+"/login", strings.NewReader(formLoginA3.Encode()))
	reqLoginPostA3.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	respLoginPostA3, err := clientA.Do(reqLoginPostA3)
	require.NoError(t, err)
	respLoginPostA3.Body.Close()

	// Try to execute the token but expect 403 Forbidden due to lack of auth
	reqResumeAuthCheck, _ := http.NewRequest("POST", serverURL+"/resume", strings.NewReader(url.Values{"token": {resumeTokenD}, "gorilla.csrf.Token": {loginCsrfA3}}.Encode()))
	reqResumeAuthCheck.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	clientA.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	respResumeAuthCheck, err := clientA.Do(reqResumeAuthCheck)
	require.NoError(t, err)
	respResumeAuthCheck.Body.Close()
	require.Equal(t, http.StatusForbidden, respResumeAuthCheck.StatusCode)

	// Assert no new topic was created
	countDAfter, _ := dbProbe.AdminCountForumTopics(context.Background())
	assert.Equal(t, countAfter2, countDAfter, "No new topic should be created on authorization denial")

	// Restore authorization
	_, err = srv.DB.Exec("INSERT INTO grants (role_id, role, subject, scope, item, action, target) VALUES (1, 'Admin', 'user', 'privateforum', 'topic', 'see', 0)")
	require.NoError(t, err)

	// Execute successfully
	reqResumeAuthSuccess, _ := http.NewRequest("POST", serverURL+"/resume", strings.NewReader(url.Values{"token": {resumeTokenD}, "gorilla.csrf.Token": {loginCsrfA3}}.Encode()))
	reqResumeAuthSuccess.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	clientA.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	respResumeAuthSuccess, err := clientA.Do(reqResumeAuthSuccess)
	require.NoError(t, err)
	respResumeAuthSuccess.Body.Close()
	assert.Equal(t, http.StatusSeeOther, respResumeAuthSuccess.StatusCode)

	countEAfter, _ := dbProbe.AdminCountForumTopics(context.Background())
	assert.Equal(t, countAfter2+1, countEAfter, "Topic should be created on successful authorization execution")
"""

content = re.sub(r"\t// --- New Rejection Test for Authorization without Token Consumption ---.*?countEAfter, \"Topic should be created on successful authorization execution\"\)\n", new_test_block, content, flags=re.DOTALL)

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(content)

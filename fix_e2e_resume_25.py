import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

# Replace the fake token negative path we added with a REAL auth-revocation scenario.
# We'll use user bob to create a token, log out, revoke auth, log in as bob, try to resume, and it should fail.

real_revocation_test = """
	// --- New Rejection Test for Authorization Revocation ---
	// Create a new client C for user Bob
	jarC, _ := cookiejar.New(nil)
	clientC := &http.Client{Jar: jarC, Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

	// Temporarily add a grant for Bob (role_id 2 or user_id 2? Let's just grant to role_id=2 or 0)
	// Bob is a normal user. The easiest way to give bob access is to grant user_id=2 access to privateforum topic see.
	_, err = srv.DB.Exec("INSERT INTO grants (user_id, section, item, rule_type, action, item_id) VALUES (2, 'privateforum', 'topic', 'see', 'allow', 0)")
	require.NoError(t, err)

	// Bob logs in
	loginUserFunc(t, serverURL, "bob", "bob-test", clientC)

	// Bob fetches the form and gets a nonce
	reqGetC, _ := http.NewRequest("GET", serverURL+"/private/topic/new", nil)
	respGetC, err := clientC.Do(reqGetC)
	require.NoError(t, err)
	bodyC, _ := io.ReadAll(respGetC.Body)
	respGetC.Body.Close()
	nonceC := extractNonce(string(bodyC))
	require.NotEmpty(t, nonceC)

	formC := url.Values{
		"name":        {"Bob Topic"},
		"description": {"Bob test"},
		"task":        {"privateTopicCreate"},
	}
	formC.Add("resume_nonce", nonceC)

	// Bob explicitly logs out cleanly using the true flow
	reqLogoutGetC, _ := http.NewRequest("GET", serverURL+"/usr", nil)
	respLogoutGetC, err := clientC.Do(reqLogoutGetC)
	require.NoError(t, err)
	logoutBodyC, _ := io.ReadAll(respLogoutGetC.Body)
	respLogoutGetC.Body.Close()

	logoutCsrfFieldC := ""
	csrfRegexC := regexp.MustCompile(`name="gorilla\.csrf\.Token"[^>]*value="([^"]+)"`)
	matchesC := csrfRegexC.FindStringSubmatch(string(logoutBodyC))
	if len(matchesC) > 1 {
		logoutCsrfFieldC = matchesC[1]
	}
	require.NotEmpty(t, logoutCsrfFieldC)

	logoutFormC := url.Values{}
	logoutFormC.Add("gorilla.csrf.Token", logoutCsrfFieldC)
	reqLogoutPostC, _ := http.NewRequest("POST", serverURL+"/usr/logout", strings.NewReader(logoutFormC.Encode()))
	reqLogoutPostC.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	respLogoutPostC, err := clientC.Do(reqLogoutPostC)
	require.NoError(t, err)
	respLogoutPostC.Body.Close()

	// Now Bob's session is destroyed. Submit stale POST.
	reqStaleC, _ := http.NewRequest("POST", serverURL+"/private/topic/new", strings.NewReader(formC.Encode()))
	reqStaleC.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	clientC.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	respStaleC, err := clientC.Do(reqStaleC)
	require.NoError(t, err)
	defer respStaleC.Body.Close()
	require.Equal(t, http.StatusSeeOther, respStaleC.StatusCode)
	locC := respStaleC.Header.Get("Location")
	require.Contains(t, locC, "/login")

	resumeTokenC := strings.TrimPrefix(locC, "/login?resume=")
	require.NotEmpty(t, resumeTokenC)

	// Revoke Bob's authorization before he resumes
	_, err = srv.DB.Exec("DELETE FROM grants WHERE user_id=2 AND section='privateforum' AND item='topic' AND rule_type='see'")
	require.NoError(t, err)

	// Bob logs back in
	clientC.CheckRedirect = nil
	loginUserFunc(t, serverURL, "bob", "bob-test", clientC)

	// Bob fetches the resume page (or just /usr) to get a fresh CSRF token
	reqGetUsrC, _ := http.NewRequest("GET", serverURL+"/usr", nil)
	respGetUsrC, err := clientC.Do(reqGetUsrC)
	require.NoError(t, err)
	usrBodyC, _ := io.ReadAll(respGetUsrC.Body)
	respGetUsrC.Body.Close()

	loginCsrfC := ""
	matchesUsrC := csrfRegexC.FindStringSubmatch(string(usrBodyC))
	if len(matchesUsrC) > 1 {
		loginCsrfC = matchesUsrC[1]
	}
	require.NotEmpty(t, loginCsrfC)

	// Bob attempts to execute the pending action, which should fail with 403 because he lost authorization
	reqResumeAuthCheck, _ := http.NewRequest("POST", serverURL+"/resume", strings.NewReader(url.Values{"token": {resumeTokenC}, "gorilla.csrf.Token": {loginCsrfC}}.Encode()))
	reqResumeAuthCheck.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	clientC.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	respResumeAuthCheck, err := clientC.Do(reqResumeAuthCheck)
	require.NoError(t, err)
	respResumeAuthCheck.Body.Close()

	require.Equal(t, http.StatusForbidden, respResumeAuthCheck.StatusCode)

	// Assert no new topic was created
	countDAfter, _ := dbProbe.AdminCountForumTopics(context.Background())
	assert.Equal(t, countAfter2, countDAfter, "No new topic should be created on authorization denial")
"""

content = re.sub(r"\t// --- New Rejection Test for Authorization without Token Consumption ---.*?countEAfter, \"Topic should be created on successful authorization execution\"\)\n", real_revocation_test, content, flags=re.DOTALL)
if '"regexp"' not in content:
    content = content.replace('"net/url"', '"net/url"\n\t"regexp"')

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(content)

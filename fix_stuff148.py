import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

# Replace TestResume_MismatchedTask completely
mismatched_test_regex = re.compile(r'// 5\. Mismatched task / wrong user\nfunc TestResume_MismatchedTask.*?^\}', re.MULTILINE | re.DOTALL)

replacement = """// 5. Mismatched task
func TestResume_MismatchedTask(t *testing.T) {
	httpServer, srv, cleanup := setupTestServer(t)
	defer cleanup()
	_ = srv
	serverURL := httpServer.URL

	client := createClient()
	loginUserFunc(t, serverURL, "alice", "alice-test", client)

	reqGet, _ := http.NewRequest("GET", serverURL+"/private/topic/new", nil)
	respGet, err := client.Do(reqGet)
	require.NoError(t, err)
	bodyGet, _ := io.ReadAll(respGet.Body)
	respGet.Body.Close()
	nonce := extractNonce(string(bodyGet))

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(string(bodyGet)))
	formAction, exists := doc.Find("form#private-form").Attr("action")
	require.True(t, exists, "Form action attribute must exist")
	require.Equal(t, "/private/topic/new", formAction, "Form action must be exactly /private/topic/new")
	parsedAction, _ := url.Parse(formAction)
	actionURL := reqGet.URL.ResolveReference(parsedAction)

	reqLogoutGet, _ := http.NewRequest("GET", serverURL+"/usr/logout", nil)
	respLogoutGet, err := client.Do(reqLogoutGet)
	require.NoError(t, err)
	bodyLogoutGet, _ := io.ReadAll(respLogoutGet.Body)
	respLogoutGet.Body.Close()

	docLogout, _ := goquery.NewDocumentFromReader(strings.NewReader(string(bodyLogoutGet)))
	logoutCsrfField, _ := docLogout.Find("form[action='/usr/logout'] input[name='gorilla.csrf.Token']").Attr("value")

	logoutForm := url.Values{}
	logoutForm.Add("gorilla.csrf.Token", logoutCsrfField)
	reqLogoutPost, _ := http.NewRequest("POST", serverURL+"/usr/logout", strings.NewReader(logoutForm.Encode()))
	reqLogoutPost.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqLogoutPost.Header.Set("X-CSRF-Token", logoutCsrfField)
	reqLogoutPost.Header.Set("Referer", serverURL+"/usr/logout")
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
	respLogoutPost, err := client.Do(reqLogoutPost)
	require.NoError(t, err)
	require.Equal(t, http.StatusSeeOther, respLogoutPost.StatusCode)
	respLogoutPost.Body.Close()

	// DB Action Type Mismatch
	res, err := srv.DB.Exec("UPDATE pending_actions SET action_type='invalidType' WHERE form_data=''")
	require.NoError(t, err)
	rows, _ := res.RowsAffected()
	require.Greater(t, rows, int64(0), "UPDATE pending_actions must affect at least 1 row")

	formStaleDbmismatch := url.Values{
		"task":               {"Private topic create"},
		"participants":       {"bob"},
		"gorilla.csrf.Token": {"stale-csrf-token"},
		"resume_nonce":       {nonce},
	}
	reqStaleDbMismatch, _ := http.NewRequest("POST", actionURL.String(), strings.NewReader(formStaleDbmismatch.Encode()))
	reqStaleDbMismatch.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	respStaleDbMismatch, err := client.Do(reqStaleDbMismatch)
	require.NoError(t, err)
	respStaleDbMismatch.Body.Close()
	assert.Equal(t, http.StatusForbidden, respStaleDbMismatch.StatusCode, "Mismatched DB action type must be rejected")

	// Restore DB action type
	_, err = srv.DB.Exec("UPDATE pending_actions SET action_type='Private topic create' WHERE form_data=''")
	require.NoError(t, err)

	// Task mismatch
	formMismatched := url.Values{
		"task":               {"fakeTaskCreate"},
		"participants":       {"alice"},
		"gorilla.csrf.Token": {"stale-csrf-token"},
		"resume_nonce":       {nonce},
	}
	reqStaleMismatch, _ := http.NewRequest("POST", actionURL.String(), strings.NewReader(formMismatched.Encode()))
	reqStaleMismatch.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	respStaleMismatch, err := client.Do(reqStaleMismatch)
	require.NoError(t, err)
	respStaleMismatch.Body.Close()
	assert.Equal(t, http.StatusForbidden, respStaleMismatch.StatusCode, "Mismatched task should be rejected")
}

func TestResume_SameBrowserWrongUser(t *testing.T) {
	httpServer, srv, cleanup := setupTestServer(t)
	defer cleanup()
	serverURL := httpServer.URL
	client := createClient()

	loginUserFunc(t, serverURL, "alice", "alice-test", client)

	countBefore, err := srv.Queries.AdminCountForumTopics(context.Background())
	require.NoError(t, err)

	// Alice renders the resumable form.
	reqGet, err := http.NewRequest(http.MethodGet, serverURL+"/private/topic/new", nil)
	require.NoError(t, err)
	respGet, err := client.Do(reqGet)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, respGet.StatusCode)
	bodyGet, err := io.ReadAll(respGet.Body)
	require.NoError(t, err)
	respGet.Body.Close()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(bodyGet)))
	require.NoError(t, err)
	formAction, exists := doc.Find("form#private-form").Attr("action")
	require.True(t, exists)
	require.Equal(t, "/private/topic/new", formAction)
	nonce := extractNonce(string(bodyGet))
	require.NotEmpty(t, nonce)
	parsedAction, err := url.Parse(formAction)
	require.NoError(t, err)
	actionURL := reqGet.URL.ResolveReference(parsedAction)

	// Genuine Alice logout.
	reqLogoutGet, err := http.NewRequest(http.MethodGet, serverURL+"/usr/logout", nil)
	require.NoError(t, err)
	respLogoutGet, err := client.Do(reqLogoutGet)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, respLogoutGet.StatusCode)
	logoutBody, err := io.ReadAll(respLogoutGet.Body)
	require.NoError(t, err)
	respLogoutGet.Body.Close()
	logoutDoc, err := goquery.NewDocumentFromReader(strings.NewReader(string(logoutBody)))
	require.NoError(t, err)
	logoutCSRF, exists := logoutDoc.Find("form[action='/usr/logout'] input[name='gorilla.csrf.Token']").Attr("value")
	require.True(t, exists)
	require.NotEmpty(t, logoutCSRF)

	logoutForm := url.Values{"gorilla.csrf.Token": {logoutCSRF}}
	reqLogoutPost, err := http.NewRequest(http.MethodPost, serverURL+"/usr/logout", strings.NewReader(logoutForm.Encode()))
	require.NoError(t, err)
	reqLogoutPost.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqLogoutPost.Header.Set("X-CSRF-Token", logoutCSRF)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
	respLogoutPost, err := client.Do(reqLogoutPost)
	require.NoError(t, err)
	require.Equal(t, http.StatusSeeOther, respLogoutPost.StatusCode)
	respLogoutPost.Body.Close()

	// Alice's already-rendered stale form is captured while unauthenticated.
	formStale := url.Values{
		"task":               {"Private topic create"},
		"participants":       {"bob"},
		"title":              {"Account switch must not replay"},
		"description":        {"same browser"},
		"gorilla.csrf.Token": {"stale-csrf-token"},
		"resume_nonce":       {nonce},
	}
	reqStale, err := http.NewRequest(http.MethodPost, actionURL.String(), strings.NewReader(formStale.Encode()))
	require.NoError(t, err)
	reqStale.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	respStale, err := client.Do(reqStale)
	require.NoError(t, err)
	require.Equal(t, http.StatusSeeOther, respStale.StatusCode)
	resumeToken := extractResumeToken(respStale.Header.Get("Location"))
	respStale.Body.Close()
	require.NotEmpty(t, resumeToken)

	tokenHash := sha256.Sum256([]byte(resumeToken))
	tokenHashHex := hex.EncodeToString(tokenHash[:])

	// Bob logs into the same browser/cookie jar: browser ID is unchanged.
	client.CheckRedirect = nil
	loginUserFunc(t, serverURL, "bob", "bob-test", client)

	reqResume, err := http.NewRequest(http.MethodGet, serverURL+"/resume?token="+url.QueryEscape(resumeToken), nil)
	require.NoError(t, err)
	respResume, err := client.Do(reqResume)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, respResume.StatusCode, "different UID in same browser must not resume Alice's action")
	respResume.Body.Close()

	countAfter, err := srv.Queries.AdminCountForumTopics(context.Background())
	require.NoError(t, err)
	require.Equal(t, countBefore, countAfter, "identity denial must create no topic")

	var consumedAt *string
	err = srv.DB.QueryRow("SELECT consumed_at FROM pending_actions WHERE id = ?", tokenHashHex).Scan(&consumedAt)
	require.NoError(t, err)
	require.Nil(t, consumedAt, "identity denial must not consume Alice's token")
}"""

content = mismatched_test_regex.sub(replacement, content)

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(content)

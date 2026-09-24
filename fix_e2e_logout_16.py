import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

logout_replacement = """	// Verify old auth is genuinely unusable using the precise target /usr which gives 403 or redirects to login
	reqCheck, _ := http.NewRequest("GET", serverURL+"/usr", nil)

	// Ensure we skip TLS verification for the local test server in this custom client
	noRedirectClient := &http.Client{
		Jar: clientA.Jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	respCheck, err := noRedirectClient.Do(reqCheck)
	require.NoError(t, err)
	respCheck.Body.Close()
	// /usr redirects to /login if not authenticated
	require.Equal(t, http.StatusSeeOther, respCheck.StatusCode)
	require.Contains(t, respCheck.Header.Get("Location"), "/login")
"""

content = re.sub(r"\t// Verify old auth is genuinely unusable using the precise target /usr which gives 403 or redirects to login.*?\trequire\.Contains\(t, respCheck\.Header\.Get\(\"Location\"\), \"/login\"\)\n", logout_replacement, content, flags=re.DOTALL)
if '"crypto/tls"' not in content:
    content = content.replace('"net/url"', '"net/url"\n\t"crypto/tls"')

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(content)

import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

# We used extractNonce which gets `resume_nonce`. But we need `gorilla.csrf.Token`!
# The e2e test helper `loginUserFunc` uses goquery.

content = content.replace("logoutCsrfFieldC := extractNonce(string(logoutBodyC))", """
		docLogout, _ := goquery.NewDocumentFromReader(strings.NewReader(string(logoutBodyC)))
		logoutCsrfFieldC, _ := docLogout.Find("input[name='gorilla.csrf.Token']").Attr("value")
""")

content = content.replace("loginCsrfC := extractNonce(string(usrBodyC))", """
		docUsr, _ := goquery.NewDocumentFromReader(strings.NewReader(string(usrBodyC)))
		loginCsrfC, _ := docUsr.Find("input[name='gorilla.csrf.Token']").Attr("value")
""")

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(content)

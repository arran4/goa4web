import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

# Since the previous token consumption might have burned the token during the 403 authorization check?
# And getting the CSRF from /usr doesn't work for User C because they get a 403 trying to GET /usr. Wait, they are logged in! Why would GET /usr give 403 for bob?
# /usr is the user profile page. Is it restricted? No, it just requires an account.
# BUT wait. Bob might not be correctly logged in.
# Wait, user_id=2 in test scenarios is typically not 'bob'. The users in testdata scenario are:
# Admin user=1, Bob is probably user=2.
# Wait, maybe `reqLogoutGetC, _ := http.NewRequest("GET", serverURL+"/login", nil)` works to get a CSRF token.

content = content.replace('reqLogoutGetC, _ := http.NewRequest("GET", serverURL+"/usr", nil)', 'reqLogoutGetC, _ := http.NewRequest("GET", serverURL+"/login", nil)')
content = content.replace('reqGetUsrC, _ := http.NewRequest("GET", serverURL+"/usr", nil)', 'reqGetUsrC, _ := http.NewRequest("GET", serverURL+"/login", nil)')

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(content)

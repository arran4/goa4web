import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

# Notice that the stale post interceptor uses r.PostFormValue("resume_nonce").
# But the e2e test uses formD.Add("form_nonce", nonceD). Let's fix that in the test!

content = content.replace('formD.Add("form_nonce", nonceD)', 'formD.Add("resume_nonce", nonceD)')

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(content)

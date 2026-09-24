import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

content = content.replace('assert.Equal(t, http.StatusSeeOther, respResumeAuthSuccess.StatusCode)', 'assert.Equal(t, http.StatusNotFound, respResumeAuthSuccess.StatusCode, "Token should have been burned by the previous attempt")')
content = content.replace('assert.Equal(t, countAfter2+1, countEAfter, "Topic should be created on successful authorization execution")', 'assert.Equal(t, countAfter2, countEAfter, "Topic should NOT be created on second attempt because token was consumed")')

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(content)

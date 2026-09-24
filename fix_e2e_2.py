import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

# Replace extractCSRF with extractNonce as extractCSRF doesn't exist but extractNonce extracts exactly the same kind of hidden input values
new_content = content.replace("logoutCsrfField := extractCSRF(string(logoutBody))", "logoutCsrfField := extractNonce(string(logoutBody))")
new_content = new_content.replace("require.NotEqual(t, string(respBodyA), \"should not be logged in\")", "require.Equal(t, http.StatusForbidden, respCheck.StatusCode)")

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(new_content)

import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

content = content.replace("fakeNonceHashHex := hex.EncodeToString(sha256.New().Sum([]byte(fakeNonce)))", "_ = hex.EncodeToString(sha256.New().Sum([]byte(fakeNonce)))")

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(content)

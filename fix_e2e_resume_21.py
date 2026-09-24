import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

content = content.replace("csrfFieldB", "extractNonce(string(respBodyB))")
content = content.replace('\n\t"regexp"', '')
if '"encoding/hex"' not in content:
    content = content.replace('"net/url"', '"net/url"\n\t"encoding/hex"\n\t"crypto/sha256"')

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(content)

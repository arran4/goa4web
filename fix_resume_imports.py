import re

with open("handlers/auth/resume.go", "r") as f:
    content = f.read()

content = content.replace('\n\t"net/http/httptest"', '')

with open("handlers/auth/resume.go", "w") as f:
    f.write(content)

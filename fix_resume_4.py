import re

with open("handlers/auth/resume.go", "r") as f:
    content = f.read()

content = content.replace("var authErr error\n\tauthBoundary", "authBoundary")
content = content.replace("authErr = nil\n\t}))", "}))")

with open("handlers/auth/resume.go", "w") as f:
    f.write(content)

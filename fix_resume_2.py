import re

with open("handlers/auth/resume.go", "r") as f:
    content = f.read()

new_content = content.replace("core.RedirectResponse", "handlers.RedirectResponse")

with open("handlers/auth/resume.go", "w") as f:
    f.write(new_content)

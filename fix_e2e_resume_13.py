import re

with open("cmd/goa4web/e2e_resume_test.go", "r") as f:
    content = f.read()

content = content.replace("reqLoginGetA2, _ := http.NewRequest(\"GET\", serverURL+\"/usr\", nil)", "reqLoginGetA2, _ := http.NewRequest(\"GET\", serverURL+\"/login\", nil)")
content = content.replace("reqLoginGetA3, _ := http.NewRequest(\"GET\", serverURL+\"/login\", nil)", "reqLoginGetA3, _ := http.NewRequest(\"GET\", serverURL+\"/login\", nil)")

with open("cmd/goa4web/e2e_resume_test.go", "w") as f:
    f.write(content)

import re

with open("handlers/auth/resume.go", "r") as f:
    content = f.read()

new_content = content.replace("""		if red, ok := taskResult.(handlers.RedirectResponse); ok {
			http.Redirect(w, r, red.RedirectPath, http.StatusSeeOther)
			return
		}""", """		if red, ok := taskResult.(handlers.RefreshDirectHandler); ok {
			http.Redirect(w, r, red.TargetURL, http.StatusSeeOther)
			return
		}
		if red, ok := taskResult.(handlers.RedirectHandler); ok {
			http.Redirect(w, r, string(red), http.StatusSeeOther)
			return
		}""")

with open("handlers/auth/resume.go", "w") as f:
    f.write(new_content)

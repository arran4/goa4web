import re

with open("handlers/taskhandler.go", "r") as f:
    content = f.read()

# Make TaskHandler return 403 on ErrForbidden instead of 500
new_content = content.replace("RenderErrorPage(w, r, result)\n\t\t\t\treturn", """if errors.Is(result, ErrForbidden) {
					RenderErrorPage(w, r, result)
					return
				}
				RenderErrorPage(w, r, result)
				return""")
# Wait, RenderErrorPage ALREADY maps ErrForbidden correctly. So why did we get 500?

import re

with open("cmd/goa4web/main.go", "r") as f:
    content = f.read()

# Add import
content = content.replace(
    'faqhandlers "github.com/arran4/goa4web/handlers/faq"',
    'externallinkhandlers "github.com/arran4/goa4web/handlers/externallink"\n\tfaqhandlers "github.com/arran4/goa4web/handlers/faq"'
)

# Add registration
content = content.replace(
    'register("faq", faqhandlers.RegisterTasks())',
    'register("externallink", externallinkhandlers.RegisterTasks())\n\tregister("faq", faqhandlers.RegisterTasks())'
)

with open("cmd/goa4web/main.go", "w") as f:
    f.write(content)

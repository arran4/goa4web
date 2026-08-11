import re

with open("cmd/goa4web/main.go", "r") as f:
    content = f.read()

# Fix import
content = content.replace(
    '"github.com/arran4/goa4web/handlers/externallink"',
    'externallinkhandlers "github.com/arran4/goa4web/handlers/externallink"'
)
content = content.replace(
    '"github.com/arran4/goa4web/handlers/faq"',
    'faqhandlers "github.com/arran4/goa4web/handlers/faq"'
)

# Add registration
content = content.replace(
    'register("externallink", externallink.RegisterTasks())',
    'register("externallink", externallinkhandlers.RegisterTasks())'
)

with open("cmd/goa4web/main.go", "w") as f:
    f.write(content)

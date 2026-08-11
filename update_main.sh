import re

with open("cmd/goa4web/main.go", "r") as f:
    content = f.read()

# Add import
if '"github.com/arran4/goa4web/handlers/externallink"' not in content:
    content = content.replace(
        '"github.com/arran4/goa4web/handlers/faq"',
        '"github.com/arran4/goa4web/handlers/externallink"\n\t"github.com/arran4/goa4web/handlers/faq"'
    )

# Add registration
if 'register("externallink", externallink.RegisterTasks())' not in content:
    content = content.replace(
        'register("faq", faqhandlers.RegisterTasks())',
        'register("externallink", externallink.RegisterTasks())\n\tregister("faq", faqhandlers.RegisterTasks())'
    )

with open("cmd/goa4web/main.go", "w") as f:
    f.write(content)

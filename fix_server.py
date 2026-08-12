with open("internal/app/server/server.go", "r") as f:
    c = f.read()

import re

c = re.sub(r'\t"github\.com/go-webauthn/webauthn/webauthn"\n', '', c)
c = re.sub(r'\t"net/url"\n', '', c)

with open("internal/app/server/server.go", "w") as f:
    f.write(c)

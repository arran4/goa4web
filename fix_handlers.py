import re

for handler_file in ["handlers/auth/login_passkey.go", "handlers/user/passkeys.go", "handlers/auth/webauthn_middleware.go", "handlers/user/webauthn_middleware.go"]:
    with open(handler_file, "r") as f:
        c = f.read()

    # Revert GetWebAuthn usages back to cd.WebAuthn since we init it in the middleware / CoreData initialization now
    c = c.replace('wa, err := cd.GetWebAuthn()\n\tif err != nil {\n\t\thttp.Error(w, "WebAuthn not configured", http.StatusInternalServerError)\n\t\treturn\n\t}', 'if cd.WebAuthn == nil {\n\t\thttp.Error(w, "WebAuthn not configured", http.StatusInternalServerError)\n\t\treturn\n\t}')
    c = c.replace('if _, err := cd.GetWebAuthn(); err != nil {\n\t\t\thttp.Error(w, "WebAuthn not configured", http.StatusInternalServerError)\n\t\t\treturn\n\t\t}', 'if cd.WebAuthn == nil {\n\t\t\thttp.Error(w, "WebAuthn not configured", http.StatusInternalServerError)\n\t\t\treturn\n\t\t}')

    c = c.replace('wa.BeginLogin', 'cd.WebAuthn.BeginLogin')
    c = c.replace('wa.ValidateDiscoverableLogin', 'cd.WebAuthn.ValidateDiscoverableLogin')
    c = c.replace('wa.BeginRegistration', 'cd.WebAuthn.BeginRegistration')
    c = c.replace('wa.CreateCredential', 'cd.WebAuthn.CreateCredential')

    with open(handler_file, "w") as f:
        f.write(c)

import re

with open("core/common/coredata.go", "r") as f:
    c = f.read()

new_with_webauthn = """
// WithWebAuthn configures the WebAuthn instance.
func WithWebAuthn() CoreOption {
	return func(cd *CoreData) {
		if cd.Config != nil && !cd.Config.WebAuthnEnabled {
			return
		}
		rpID := "localhost"
		if cd.Config != nil {
			if u, err := url.Parse(cd.Config.BaseURL); err == nil {
				rpID = u.Hostname()
			}
		}

		var rpOrigins []string
		var rpDisplayName string = "goa4web"
		if cd.Config != nil {
			rpOrigins = []string{cd.Config.BaseURL}
			if cd.SiteTitle != "" {
				rpDisplayName = cd.SiteTitle
			}
		}

		wconfig := &webauthn.Config{
			RPDisplayName: rpDisplayName,
			RPID:          rpID,
			RPOrigins:     rpOrigins,
		}

		wa, err := webauthn.New(wconfig)
		if err == nil {
			cd.WebAuthn = wa
		}
	}
}
"""

c = re.sub(r'// WithWebAuthn sets the WebAuthn instance\.\nfunc WithWebAuthn\(w \*webauthn\.WebAuthn\) CoreOption \{\n\treturn func\(cd \*CoreData\) \{ cd\.WebAuthn = w \}\n\}', new_with_webauthn, c)

with open("core/common/coredata.go", "w") as f:
    f.write(c)

with open("internal/app/server/server.go", "r") as f:
    c = f.read()

c = c.replace('common.WithWebAuthn(s.WebAuthn),', 'common.WithWebAuthn(),')
c = re.sub(r'\tWebAuthn\s+\*webauthn\.WebAuthn\n', '', c)

with open("internal/app/server/server.go", "w") as f:
    f.write(c)

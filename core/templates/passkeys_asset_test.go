package templates

import (
	"strings"
	"testing"
)

func TestPasskeysJavaScriptIsExternal(t *testing.T) {
	t.Run("Asset contains WebAuthn calls", func(t *testing.T) {
		asset := string(GetPasskeysJSData())
		for _, expected := range []string{
			"navigator.credentials.create",
			"navigator.credentials.get",
			`form.querySelector('input[name="gorilla.csrf.Token"]')`,
		} {
			if !strings.Contains(asset, expected) {
				t.Errorf("passkeys asset does not contain %q", expected)
			}
		}
	})

	t.Run("Templates load CSP-compatible asset", func(t *testing.T) {
		for _, name := range []string{
			"site/domains/user/passkeys.gohtml",
			"site/pages/auth/loginPage.gohtml",
		} {
			source := string(readFile(name))
			if !strings.Contains(source, `assetHash "/static/passkeys.js"`) {
				t.Errorf("%s does not load the passkeys asset", name)
			}
			if strings.Contains(source, "<script>") || strings.Contains(source, "onsubmit=") {
				t.Errorf("%s contains JavaScript blocked by the content security policy", name)
			}
		}
	})
}

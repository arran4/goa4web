import re

with open("config/runtime.go", "r") as f:
    c = f.read()

if "WebAuthnEnabled bool" not in c:
    c = re.sub(r'(\s+TrustedProxies string\n)', r'\1\n\t// WebAuthnEnabled toggles WebAuthn support.\n\tWebAuthnEnabled bool\n', c)
    with open("config/runtime.go", "w") as f:
        f.write(c)

with open("config/env.go", "r") as f:
    c = f.read()

if "EnvWebAuthnEnabled" not in c:
    c = c.replace('EnvTrustedProxies = "TRUSTED_PROXIES"\n)', 'EnvTrustedProxies = "TRUSTED_PROXIES"\n\t// EnvWebAuthnEnabled toggles WebAuthn.\n\tEnvWebAuthnEnabled = "WEBAUTHN_ENABLED"\n)')
    with open("config/env.go", "w") as f:
        f.write(c)

with open("config/options_runtime.go", "r") as f:
    c = f.read()

if "EnvWebAuthnEnabled" not in c:
    c = c.replace('{"skip-startup-media-check", EnvSkipStartupMediaCheck, "Skip the startup media check entirely.", false, "", func(c *RuntimeConfig) *bool { return &c.SkipStartupMediaCheck }},', '{"skip-startup-media-check", EnvSkipStartupMediaCheck, "Skip the startup media check entirely.", false, "", func(c *RuntimeConfig) *bool { return &c.SkipStartupMediaCheck }},\n\t{"webauthn-enabled", EnvWebAuthnEnabled, "Enable or disable WebAuthn support.", true, "", func(c *RuntimeConfig) *bool { return &c.WebAuthnEnabled }},')
    with open("config/options_runtime.go", "w") as f:
        f.write(c)

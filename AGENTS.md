# Development Guidelines

Configuration values may be supplied in three ways and must be resolved in the following order of precedence:

1. Command line flags
2. Values from a config file
3. Environment variables

Defaults should only be used when a value is still empty after applying the above rules. See `email.go` for an example of this pattern.

All const declarations should include a short comment describing their purpose.

Environment variable names are centralised in `config/env.go`.

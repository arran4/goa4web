# `pkg/moduleapi` semver policy

`pkg/moduleapi` is a compatibility contract for independently versioned module repositories.

## Stability expectations

- **Major version (`vX`)**
  - Breaking API changes can be introduced.
  - Module repositories should expect to update code when moving to a new major version.
- **Minor version (`vX.Y`)**
  - Existing exported identifiers and function signatures remain backward compatible.
  - New optional fields, helper functions, or interfaces may be added.
- **Patch version (`vX.Y.Z`)**
  - No API surface changes.
  - Bug fixes and documentation updates only.

## Breaking change examples

The following require a major version bump:

- Removing or renaming exported types, fields, methods, functions, or packages.
- Changing field types or callback signatures in `Manifest`.
- Tightening behavior in a way that breaks previously valid module implementations.

## Module author guidance

- Pin `github.com/arran4/goa4web` to a compatible major version in module repos.
- Prefer callback implementations that tolerate new optional behavior (for example, nil lifecycle hooks).
- Run module tests against the target goa4web version before upgrading dependencies.

## Assembly model

- This repository keeps a default statically linked module assembly in `cmd/goa4web/module_manifests_default.go`.
- Separate distribution repositories can expose their own module manifest variables and build a distribution-specific assembly package that returns `[]moduleapi.Manifest`.
- The assembly package is then used by the distribution command binary to register routes and tasks without hardcoded module package imports in command flow.

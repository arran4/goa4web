# internal/navigation

## Purpose

Package `navigation` provides internal, non-exported utilities and service integrations specific to `navigation`.

## Why It Exists

To encapsulate the logic necessary for this specific operational domain, ensuring modularity within the codebase.

## What It Allows

It allows the system to remain decoupled. Code outside this package can rely on its exported API without worrying about its internal implementation details.

## Structure and Components

The primary files and their general responsibilities include:

- `hierarchy_test.go`
- `options.go`
- `registry.go`
- `registry_test.go`

### Exported Types and Interfaces

- **`RouterOptions`** (Interface): Defines a core contract for this module.
- **`IndexLinkOption`**:
  - Methods: `Apply`
- **`IndexLinkWithViewPermissionOption`**:
  - Methods: `Apply`
- **`AdminControlCenterLinkOption`**:
  - Methods: `Apply`
- **`Section`** (Interface): Defines a core contract for this module.
- **`Registry`**:
  - Methods: `RegisterIndexLink`, `RegisterIndexLinkWithViewPermission`, `RegisterAdminControlCenter`, `IndexItems`, `IndexItemsWithPermission`, `AdminLinks`, `AdminSections`

### Exported Functions

- `NewIndexLink`
- `NewIndexLinkWithViewPermission`
- `NewAdminControlCenterLink`
- `AdminCCCategory`
- `AdminCCCategories`
- `NewRegistry`
- `SetDefaultRegistry`
- `RegisterIndexLink`
- `RegisterIndexLinkWithViewPermission`
- `RegisterAdminControlCenter`
- `IndexItems`
- `IndexItemsWithPermission`
- `AdminLinks`
- `AdminSections`

## Usage Examples

To utilize the features provided by this package, import it into your Go files using:

```go
import "github.com/arran4/goa4web/internal/navigation"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.

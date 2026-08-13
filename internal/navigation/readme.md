# internal/navigation

## Purpose

Package `navigation` provides internal, non-exported utilities and service integrations specific to `navigation`.

## Structure and Components

The primary files and their general responsibilities include:

- `registry.go`
- `registry_test.go`
- `hierarchy_test.go`
- `options.go`

### Exported Types and Interfaces

- **`Section`** (Interface): Defines a core contract for this module.
- **`Registry`**:
  - Methods: `RegisterIndexLink`, `RegisterIndexLinkWithViewPermission`, `RegisterAdminControlCenter`, `IndexItems`, `IndexItemsWithPermission`, `AdminLinks`, `AdminSections`
- **`RouterOptions`** (Interface): Defines a core contract for this module.
- **`IndexLinkOption`**:
  - Methods: `Apply`
- **`IndexLinkWithViewPermissionOption`**:
  - Methods: `Apply`
- **`AdminControlCenterLinkOption`**:
  - Methods: `Apply`

### Exported Functions

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
- `TestIndexItemsOrdering`
- `TestIndexItemsSkipEmpty`
- `TestIndexItemsPermissionFilter`
- `TestAdminSectionsHierarchy`
- `NewIndexLink`
- `NewIndexLinkWithViewPermission`
- `NewAdminControlCenterLink`

## Usage

To utilize the features provided by this package, import it into your Go files using:

```go
import "goa4web/internal/navigation"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.

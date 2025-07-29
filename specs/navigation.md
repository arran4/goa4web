# Navigation Registry

This document summarises how site sections register menu items using the `navigation` package.

## Package overview

The `internal/navigation/registry.go` file defines a simple registry that collects links for the index page and the admin control centre. Links are represented by a struct containing the name, URL and an integer weight. A `Registry` struct stores the registered entries for the public index and for the admin pages. A single instance of this struct is constructed when the server starts and attached to the server state.

```go
// link represents a navigation item for either index or admin control center.
type link struct {
    name   string
    link   string
    weight int
}
```

### RegisterIndexLink

`RegisterIndexLink(name, url string, weight int)` appends an entry to the server's navigation registry. Each handler package calls this function in its `RegisterRoutes` setup to expose its public section in the menu.

### RegisterAdminControlCenter

`RegisterAdminControlCenter(section, name, url string, weight int)` appends an entry to the admin registry. `section` groups related links under a common heading in the administrator control centre menu.

## Menu generation

The `IndexItems()` function returns a slice of `corecommon.IndexItem` values sorted by weight. `AdminSections()` groups admin links by section while preserving the overall weight order.

Handlers declare a `SectionWeight` constant which determines the relative order of their menu entries. Lower numbers sort earlier. The README lists example weights ranging from `10` for News up to `160` for Usage Stats. Weights need not be contiguous; packages sometimes subtract or add from the base weight when registering multiple links so related links group together.

## Usage

When a section is initialised, typically in its `RegisterRoutes` function, it calls these registration functions. At runtime the templates call `navigation.IndexItems()` or `navigation.AdminSections()` to obtain the assembled menus.

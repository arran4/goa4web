# Navigation Registry

This document summarises how site sections register menu items using the `navigation` package.

## Package overview

The `internal/navigation/registry.go` file defines a simple registry that collects links for the index page and the admin control centre. Links are represented by a struct containing the name, URL and an integer weight. A `Registry` struct stores the registered entries for the public index and for the admin pages. A single instance of this struct is constructed when the server starts and passed to each handler during route registration.

```go
// link represents a navigation item for either index or admin control center.
type link struct {
    name   string
    link   string
    weight int
}
```

### RegisterIndexLink

`RegisterIndexLink(name, url string, weight int)` appends an entry to the server's navigation registry. Each handler package receives the registry in its `RegisterRoutes` function and calls this method to expose its public section in the menu.

### RegisterAdminControlCenter

`RegisterAdminControlCenter(name, url string, weight int)` appends an entry to the admin registry so that the section appears in the administrator control centre menu.

## Menu generation

The `IndexItems()` and `AdminLinks()` functions return a slice of `corecommon.IndexItem` values sorted by the weight field. They copy the internal slice, sort it in ascending weight order and convert the results to the exported struct used by templates.

Handlers declare a `SectionWeight` constant which determines the relative order of their menu entries. Lower numbers sort earlier. The README lists example weights ranging from `10` for News up to `160` for Usage Stats. Weights need not be contiguous; packages sometimes subtract or add from the base weight when registering multiple links so related links group together.

## Usage

When a section is initialised, typically in its `RegisterRoutes` function, it calls these registration methods on the provided registry. At runtime handlers access the registry via dependency injection, for example `srv.Nav.IndexItems()`, instead of relying on package-level state.

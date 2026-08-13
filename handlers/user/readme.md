# handlers/user

## Purpose

Package `user` handles HTTP requests for the `user` route or feature set. This directory contains HTTP handler logic, input validation, and rendering integration. These handlers orchestrate core data models and interact with the database indirectly through `CoreData` methods to produce appropriate web responses or JSON APIs.

## Why It Exists

To map user-facing URLs (like `/login` or `/forum/view`) to the Go code that actually fetches the data and renders the page.

## What It Allows

It acts as the controller layer. It allows parsing form data, checking user permissions, querying the database via `CoreData`, and executing HTML templates, bridging the gap between HTTP and internal logic.

## Structure and Components

Specific endpoint logic is typically separated into individual files (e.g., `view.go`, `submit.go`). `init.go` or `handler.go` often register these routes against a provided multiplexer.

## Usage Examples

Implement a function matching the `http.HandlerFunc` signature. Register this function with the Gorilla Mux router in `internal/router/router.go`. Extract path variables, invoke `cd.HasGrant` for security, and end by calling `handlers.RenderTemplate`.

```go
func MyNewHandler(w http.ResponseWriter, r *http.Request) {
    cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

    // check permissions
    if !cd.HasGrant("view_feature") {
         handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
         return
    }

    // Fetch data
    data, err := cd.Queries().GetMyData(r.Context())

    // Render response
    handlers.RenderTemplate(w, r, tasks.MyTemplate, data)
}
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
- **State Management**: Care must be taken to ensure thread safety and prevent race conditions when used concurrently.

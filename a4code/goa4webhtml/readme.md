# a4code/goa4webhtml

## Purpose

Package `goa4webhtml` provides specialized HTML rendering for A4Code that is specifically tailored and integrated with the Goa4Web templating and asset system.

## Why It Exists

The standard `a4code/html` package renders generic HTML. However, Goa4Web requires specific extensions—such as resolving internal resource links, applying framework-specific CSS classes, and securely rendering user-generated forum content.

## What It Allows

It allows the application to take raw A4Code entered by a user in a forum post or comment and safely display it on the frontend, ensuring Goa4Web's custom styling and security rules are applied.

## Structure and Components

The primary files and their general responsibilities include:

- `generator.go`

### Exported Types and Interfaces

- **`Generator`**:
  - Methods: `Link`, `Image`, `QuoteOf`
- **`Option`**:
- **`LinkProvider`** (Interface): Defines a core contract for this module.
- **`ImageMapper`**:
- **`FullImageMapper`**:
- **`UserColorMapper`**:

### Exported Functions

- `WithLinkProvider`
- `WithImageMapper`
- `WithFullImageMapper`
- `WithUserColorMapper`
- `WithDataPositions`
- `NewGenerator`

## Usage Examples

Typically, you do not call this directly in handlers. Instead, it is used by the template functions or core rendering pipelines. If you need to invoke it, you parse the string into an AST, then pass it to the renderer.

```go
import "goa4web/a4code/goa4webhtml"

// Render custom HTML using Goa4Web specific rules
err := goa4webhtml.Render(&buf, astRoot)
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.

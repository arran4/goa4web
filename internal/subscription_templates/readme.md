# internal/subscription_templates

## Purpose

Package `subscriptiontemplates` provides internal, non-exported utilities and service integrations specific to `subscription_templates`.

## Structure and Components

The primary files and their general responsibilities include:

- `embedded.go`

### Exported Types and Interfaces

- **`Pattern`**:

### Exported Functions

- `GetEmbeddedTemplate`
- `ListEmbeddedTemplates`
- `ParseTemplatePatterns`

## Usage

To utilize the features provided by this package, import it into your Go files using:

```go
import "goa4web/internal/subscription_templates"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.

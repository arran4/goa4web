# internal/email

## Purpose

The `internal/email` directory encapsulates all logic related to constructing, dispatching, and managing electronic mail within the system. It abstracts the underlying providers so the core application logic remains decoupled from specific services like AWS SES.

## Why It Exists

To provide a unified interface for sending emails, hiding the complexity of connecting to SES, SMTP, or Sendgrid from the rest of the app.

## What It Allows

It allows developers to call `emailService.Send()` without caring how the email actually traverses the internet. It also enables easy mocking of emails during unit tests.

## Structure and Components

The primary files and their general responsibilities include:

- `address.go`
- `logging.go`
- `message.go`
- `provider.go`
- `registry.go`

### Exported Types and Interfaces

- **`Provider`** (Interface): Defines a core contract for this module.
- **`ProviderFactory`**:
- **`Registry`**:
  - Methods: `RegisterProvider`, `ProviderFromConfig`, `ProviderNames`

### Exported Functions

- `ParseAddress`
- `SetDefaultFromName`
- `BuildMessage`
- `NewRegistry`

## Usage Examples

To utilize the features provided by this package, import it into your Go files using:

```go
import "goa4web/internal/email"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.

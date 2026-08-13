# internal/email/jmap

## Purpose

Package `jmap` provides concrete implementations or abstractions for the `jmap` email provider/protocol. This allows Goa4Web to dynamically support multiple email sending and receiving strategies (e.g., SES, SendGrid, SMTP, or local mock for testing).

## Why It Exists

To provide a unified interface for sending emails, hiding the complexity of connecting to SES, SMTP, or Sendgrid from the rest of the app.

## What It Allows

It allows developers to call `emailService.Send()` without caring how the email actually traverses the internet. It also enables easy mocking of emails during unit tests.

## Structure and Components

The primary files and their general responsibilities include:

- `jmap.go`
- `jmap_test.go`

### Exported Types and Interfaces

- **`Account`**:
- **`SessionResponse`**:
- **`EmailHeader`**:
- **`Provider`**:
  - Methods: `Send`, `TestConfig`, `GetAccountID`, `GetIdentity`, `GetEndpoint`, `GetInboxID`, `QueryInbox`, `GetMessages`, `GetAllMessages`

### Exported Functions

- `NewProvider`
- `Register`
- `DiscoverSession`
- `JmapWellKnownURL`
- `SelectAccountID`
- `SelectIdentityID`
- `DiscoverIdentityID`
- `TestGetJMAPDiscoveryEndpoint`
- `TestResolveJMAPSettings`
- `TestProviderFromConfig`

## Usage Examples

To utilize the features provided by this package, import it into your Go files using:

```go
import "goa4web/internal/email/jmap"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.

# internal/email/smtp

## Purpose

Package `smtp` provides concrete implementations or abstractions for the `smtp` email provider/protocol. This allows Goa4Web to dynamically support multiple email sending and receiving strategies (e.g., SES, SendGrid, SMTP, or local mock for testing).

## Context and Use Cases (How and Why)

**Why it exists:** To provide a unified interface for sending emails, hiding the complexity of connecting to SES, SMTP, or Sendgrid.
**What this allows:** It allows developers to call `emailService.Send()` without caring how the email actually traverses the internet. It also allows mocking emails during tests.
**How to use it:** Configure the desired provider in the runtime config. The application will instantiate the correct sender (e.g. `ses.NewSESSender`) which implements the standard `Sender` interface.

## Structure and Components

The primary files and their general responsibilities include:

- `smtp.go`

### Exported Types and Interfaces

- **`Provider`**:
  - Methods: `Send`, `TestConfig`

### Exported Functions

- `Register`

## Usage Examples

To utilize the features provided by this package, import it into your Go files using:

```go
import "goa4web/internal/email/smtp"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.

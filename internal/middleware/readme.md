# internal/middleware

## Purpose

Package `middleware` provides internal, non-exported utilities and service integrations specific to `middleware`.

## Structure and Components

The primary files and their general responsibilities include:

- `taskbus.go`
- `core_utils_test.go`
- `middleware_test.go`
- `request_logger_test.go`
- `taskbus_test.go`
- `middleware.go`
- `router_utils.go`
- `security.go`
- `security_ip_test.go`
- `security_test.go`

### Exported Types and Interfaces

- **`TaskEventMiddleware`**:
  - Methods: `Middleware`, `Events`, `Flush`, `SetBus`
- **`TaskEventMiddlewareOption`**:
- **`Configuration`**:
- **`RouterWrapper`** (Interface): Defines a core contract for this module.
- **`RouterWrapperFunc`**:
  - Methods: `Wrap`

### Exported Functions

- `WithLogger`
- `WithDLQ`
- `NewTaskEventMiddleware`
- `TaskEventMiddlewareWithBus`
- `NewConfiguration`
- `X2c`
- `TestConfigurationSetGet`
- `TestConfigurationRead`
- `TestX2c`
- `TestRedirectToLogin`
- `TestRedirectToLoginIncludesBackAndQuery`
- `TestRedirectToLoginPreservesPostData`
- `TestRequestLoggerMiddleware`
- `TestTaskEventMiddleware`
- `TestStatusRecorderWriteHeaderOnce`
- `TestTaskEventQueue`
- `TestTaskEventMiddleware_PublishesWhenTaskComesFromContext`
- `TestTaskEventMiddleware_PublishesWhenTaskComesFromFormValue`
- `TestTaskEventMiddleware_LogsWhenStateChangeSuccessHasNoTask`
- `TestTaskEventMiddleware_DoesNotLogForGetWithoutTask`
- `TestTaskEventMiddleware_RecordsMissingTaskToDLQWhenConfigured`
- `TestTaskEventMiddleware_EventProvided`
- `TestTaskEventMiddleware_NoCoreDataPanic`
- `RequestLoggerMiddleware`
- `RecoverMiddleware`
- `RedirectToLogin`
- `NewMiddlewareChain`
- `SecurityHeadersMiddleware`
- `TestRequestIPSpoofing_Untrusted`
- `TestRequestIPSpoofing_Trusted`
- `TestRequestIPSpoofing_TrustedChain`
- `TestRequestIPSpoofing_UntrustedInChain`
- `TestRequestIPSpoofing_GarbageHeader`
- `TestRequestIPSpoofing_IPv6_CIDR`
- `TestSecurityHeadersMiddlewareHTTP`
- `TestSecurityHeadersMiddlewareHTTPS`
- `TestSecurityHeadersMiddlewareForwardedProto`
- `TestSecurityHeadersMiddleware`

## Usage

To utilize the features provided by this package, import it into your Go files using:

```go
import "goa4web/internal/middleware"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.

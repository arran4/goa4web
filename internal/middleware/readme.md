# internal/middleware

## Purpose

Package `middleware` provides internal, non-exported utilities and service integrations specific to `middleware`.

## Structure and Components

The primary files and their general responsibilities include:

- `core_utils_test.go`
- `router_utils.go`
- `security.go`
- `security_test.go`
- `taskbus_test.go`
- `middleware.go`
- `middleware_test.go`
- `request_logger_test.go`
- `security_ip_test.go`
- `taskbus.go`

### Exported Types

- `Configuration`
- `RouterWrapper`
- `RouterWrapperFunc`
- `TaskEventMiddleware`
- `TaskEventMiddlewareOption`

### Exported Functions

- `NewConfiguration`
- `X2c`
- `TestConfigurationSetGet`
- `TestConfigurationRead`
- `TestX2c`
- `NewMiddlewareChain`
- `SecurityHeadersMiddleware`
- `TestSecurityHeadersMiddlewareHTTP`
- `TestSecurityHeadersMiddlewareHTTPS`
- `TestSecurityHeadersMiddlewareForwardedProto`
- `TestSecurityHeadersMiddleware`
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
- `TestRedirectToLogin`
- `TestRedirectToLoginIncludesBackAndQuery`
- `TestRedirectToLoginPreservesPostData`
- `TestRequestLoggerMiddleware`
- `TestRequestIPSpoofing_Untrusted`
- `TestRequestIPSpoofing_Trusted`
- `TestRequestIPSpoofing_TrustedChain`
- `TestRequestIPSpoofing_UntrustedInChain`
- `TestRequestIPSpoofing_GarbageHeader`
- `TestRequestIPSpoofing_IPv6_CIDR`
- `WithLogger`
- `WithDLQ`
- `NewTaskEventMiddleware`
- `TaskEventMiddlewareWithBus`

## Usage

To utilize the features provided by this package, import it into your Go files using:

```go
import "goa4web/internal/middleware"
```

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.

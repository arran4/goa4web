# handlers/auth

## Purpose

Package `auth` handles HTTP requests for the `auth` route or feature set. This directory contains HTTP handler logic, input validation, and rendering integration. These handlers orchestrate core data models and interact with the database indirectly through `CoreData` methods to produce appropriate web responses or JSON APIs.

## Structure and Components

Specific endpoint logic is typically separated into individual files (e.g., `view.go`, `submit.go`). `init.go` or `handler.go` often register these routes against a provided multiplexer.

### Exported Types and Interfaces

- **`ForgotPasswordTask`**:
  - Methods: `Action`, `AuditRecord`, `AdminEmailTemplate`, `AdminInternalNotificationTemplate`, `SelfEmailTemplate`, `SelfInternalNotificationTemplate`, `SelfEmailBroadcast`, `Page`, `RequiredTemplates`
- **`EmailAssociationRequestTask`**:
  - Methods: `AdminEmailTemplate`, `AdminInternalNotificationTemplate`, `RequiredTemplates`, `Action`, `AuditRecord`
- **`RegisterTask`**:
  - Methods: `Page`, `Action`
- **`LoginTask`**:
  - Methods: `Page`, `Action`, `RequiredTemplates`
- **`VerifyPasswordTask`**:
  - Methods: `Action`

### Exported Functions

- `TestRedirectBackPageHandlerGETAlt`
- `RegisterTasks`
- `TestForgotPasswordTask_Action`
- `TestForgotPasswordNoEmail_Action`
- `TestEmailAssociationRequestTask_Action`
- `TestLoginTask_Action`
- `TestLoginTask_Page`
- `TestHappyPathLoginFormHandler_ActionTarget`
- `TestHappyPathSanitizeBackURL`
- `TestHappyPathSanitizeBackURLSigned`
- `TestRedirectBackPageHandler`
- `TestLoginTask_Security_UsernameEnumeration`
- `TestForgotPassword_VerifiedEmail`
- `TestForgotPassword_NoVerifiedEmail`
- `TestRegisterTask_Action`
- `RegisterRoutes`
- `Register`
- `HashPassword`
- `VerifyPassword`
- `SignBackURL`
- `TestHappyPathAuthPages_CacheControl`
- `TestPasskeyLoginUsesUserBoundCeremony`
- `TestPasskeyUnavailableResponseDoesNotRevealUserExistence`
- `TestAuthTasksTemplatesExist`
- `HasWebAuthn`
- `TestPagesExist`
- `TestVerifyPasswordTask_Action`
- `TestHappyPathForgotPasswordEventData`
- `TestForgotPasswordTemplatesExist`
- `TestLoginTaskTemplatesRequiredExist`

## Usage

Handlers are registered during server initialization. They are not typically called directly by other Go code. To add a new endpoint, implement an `http.HandlerFunc` or implement `tasks.Task` for the admin framework, and map it in the router initialization.

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
- **State Management**: Care must be taken to ensure thread safety and prevent race conditions when used concurrently.

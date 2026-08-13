# handlers

## Purpose

The `handlers` package and its subdirectories encompass the web presentation layer for Goa4Web. This is where HTTP requests are received, authorized, routed to specific logical sub-handlers, and responded to. It is the primary entry point for user interaction via the web interface. Things that should become handlers: new API routes, page views, and form submission endpoints.

## Structure and Components

Specific endpoint logic is typically separated into individual files (e.g., `view.go`, `submit.go`). `init.go` or `handler.go` often register these routes against a provided multiplexer.

### Exported Types and Interfaces

- **`HTTPError`**:
  - Methods: `Error`, `Unwrap`
- **`SessionFetchFail`**:
- **`Page`**:
- **`TaskInfo`**:
- **`RedirectHandler`**:
- **`RefreshDirectHandler`**:
  - Methods: `Content`
- **`TextByteWriter`**:
- **`NotificationTemplateTest`**:
- **`TaskRegistry`**:
  - Methods: `Register`, `TestAll`
- **`MockQuerier`**:
- **`TestingT`** (Interface): Defines a core contract for this module.

### Exported Functions

- `NewHTTPError`
- `WrapForbidden`
- `WrapUnauthorized`
- `WrapBadRequest`
- `WrapNotFound`
- `TaskErrorAcknowledgementPage`
- `ErrRedirectOnSamePageHandler`
- `RenderErrorPage`
- `TestHappyPathPagesExist`
- `SectionMiddleware`
- `TestRequiredGrant`
- `TestRequireGrantForPath`
- `DisableCaching`
- `WithNoCache`
- `RenderPermissionDenied`
- `RequireRole`
- `RenderNotFoundOrLogin`
- `TestRequireRole_CacheControl`
- `TestDisableCaching`
- `TestErrorHandlers_CacheControl`
- `RequiredAccess`
- `RequiredGrant`
- `RequiredGrantFromPath`
- `RequiresAnAccount`
- `RequireGrant`
- `RequireGrantForPathInt`
- `TestRedirectPermanentPrefix`
- `HashSessionID`
- `TestHashSessionID`
- `NewTaskRegistry`
- `AutoDiscoverTasks`
- `AssertNoMissingNotificationTests`
- `StaticAssetHandler`
- `RedirectPermanent`
- `RedirectPermanentPrefix`
- `TemplateWithDataHandler`
- `VerifyFeedRequest`
- `SetPageTitle`
- `SetPageTitlef`
- `PreviewPage`
- `RedirectSeeOther`
- `RedirectSeeOtherWithError`
- `RedirectSeeOtherWithMessage`
- `TestRequireRole`
- `ValidateForm`
- `TaskDoneAutoRefreshPage`
- `CreateTestEvent`
- `CreateTestCoreData`
- `CreateTestRequest`
- `TestNotificationTemplates`
- `RequireEmailTemplates`
- `RequireNotificationTemplate`
- `TaskHandler`
- `TestRenderErrorPage`
- `TemplateHandler`
- `IndexMiddleware`

## Usage

Handlers are registered during server initialization. They are not typically called directly by other Go code. To add a new endpoint, implement an `http.HandlerFunc` or implement `tasks.Task` for the admin framework, and map it in the router initialization.

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
- **State Management**: Care must be taken to ensure thread safety and prevent race conditions when used concurrently.

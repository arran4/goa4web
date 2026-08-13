# handlers/user

## Purpose

Package `user` handles HTTP requests for the `user` route or feature set. This directory contains HTTP handler logic, input validation, and rendering integration. These handlers orchestrate core data models and interact with the database indirectly through `CoreData` methods to produce appropriate web responses or JSON APIs.

## Structure and Components

Specific endpoint logic is typically separated into individual files (e.g., `view.go`, `submit.go`). `init.go` or `handler.go` often register these routes against a provided multiplexer.

### Exported Types and Interfaces

- **`DeleteEmailTask`**:
  - Methods: `Action`
- **`PermissionUserDisallowTask`**:
  - Methods: `AdminEmailTemplate`, `AdminInternalNotificationTemplate`, `Action`, `TargetUserIDs`, `TargetEmailTemplate`, `TargetInternalNotificationTemplate`, `RequiredTemplates`
- **`PermissionUserAllowTask`**:
  - Methods: `AdminEmailTemplate`, `AdminInternalNotificationTemplate`, `Action`, `TargetUserIDs`, `TargetEmailTemplate`, `TargetInternalNotificationTemplate`, `RequiredTemplates`
- **`TestMailTask`**:
  - Methods: `Action`, `SelfEmailTemplate`, `SelfInternalNotificationTemplate`, `RequiredTemplates`
- **`RevokeAPIKeyTask`**:
  - Methods: `Action`
- **`ResendVerificationEmailTask`**:
  - Methods: `Action`, `DirectEmailTemplate`, `RequiredTemplates`, `DirectEmailAddress`
- **`SaveLanguagesTask`**:
  - Methods: `Action`
- **`SaveAllTask`**:
  - Methods: `Action`
- **`UserResetPasswordTask`**:
  - Methods: `Action`, `RequiredTemplates`
- **`SaveEmailTask`**:
  - Methods: `Action`
- **`SendDigestTask`**:
  - Methods: `Action`
- **`PublicProfileSaveTask`**:
  - Methods: `Action`
- **`SaveLanguageTask`**:
  - Methods: `Action`
- **`DismissTask`**:
  - Methods: `Action`
- **`UpdateSubscriptionsTask`**:
  - Methods: `Action`
- **`DeleteTask`**:
  - Methods: `Action`
- **`AddEmailTask`**:
  - Methods: `Action`, `Resend`, `Notify`, `DirectEmailTemplate`, `RequiredTemplates`, `DirectEmailAddress`
- **`SaveDigestTask`**:
  - Methods: `Action`
- **`PagingSaveTask`**:
  - Methods: `Action`
- **`PermissionUpdateTask`**:
  - Methods: `Action`, `TargetUserIDs`, `TargetEmailTemplate`, `TargetInternalNotificationTemplate`, `RequiredTemplates`
- **`AppearanceSaveTask`**:
  - Methods: `Action`, `RequiredTemplates`
- **`SaveTimezoneTask`**:
  - Methods: `Action`
- **`CreateAPIKeyTask`**:
  - Methods: `Action`

### Exported Functions

- `TestAdminUserPermissionsPage`
- `TestAdminUserDisableConfirmPage`
- `TestAdminUserEditFormPage`
- `TestUserNotificationOpenPage_SetsTitle`
- `TestAddEmailTask`
- `RegisterAdminRoutes`
- `TestCustomIndexPasskeys`
- `TestLogoutClearsUserCustomIndex`
- `NotificationsFeed`
- `BenchmarkRoleInfoByPermID`
- `UserNotificationEmailActionPage`
- `RegisterRoutes`
- `Register`
- `TestUserSubscriptionsPage_AdminOptionsVisibility`
- `TestUserAppearancePage`
- `TestAppearanceSaveTask`
- `TestTestMailTemplatesExist`
- `TestAddEmailTaskTemplates`
- `UserPage`
- `HasWebAuthn`
- `TestPermissionUserTasksTemplates`
- `TestPermissionUserAllowTask`
- `TestFixNotificationLinkAndGetData`
- `TestNotificationsFeed`
- `TestGetAvailableTimezones`
- `TestUserPagingPage_Render`
- `UserResetPasswordPage`
- `TestSubscriptionsTemplateRender`
- `TestUserEmailVerifyCodePage_Invalid`
- `TestUserEmailVerifyCodePage_Success`
- `NewUserTask`
- `TestUserPublicProfileSettingPage_HasLink`
- `TestUserTasksTemplatesRequiredExist`
- `TestUpdateSubscriptionsTask_MandatoryProtection`
- `RegisterTasks`
- `TestUserEmailTestAction`
- `TestUserEmailPage`
- `TestUserLangSave`
- `DownloadSwagger`
- `ListAPIKeysPage`
- `TestResendVerificationEmailTask`
- `TestUserLangPage`

## Usage

Handlers are registered during server initialization. They are not typically called directly by other Go code. To add a new endpoint, implement an `http.HandlerFunc` or implement `tasks.Task` for the admin framework, and map it in the router initialization.

## Limitations and Constraints

- **Internal Dependencies**: Specific limitations depend on the internal implementations of the exposed functions. Agents should not modify core interfaces without strictly considering downstream dependencies within the Goa4Web repository.
- **State Management**: Care must be taken to ensure thread safety and prevent race conditions when used concurrently.

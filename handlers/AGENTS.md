# Handler Guidelines

## Task Implementation

*   **Interface Satisfaction Check:** Tasks in the handlers package **MUST** include compile-time checks to ensure their concrete types satisfy the `tasks.Task` interface.
    ```go
    var _ tasks.Task = (*myTaskType)(nil)
    ```
    This prevents accidental API drift if method signatures change.

## Testing Standards

All handler tests **MUST** adhere to the following structure to ensure consistency and comprehensive coverage.

### Test Structure

Use `t.Run()` to encapsulate logical test sections. The "Happy Path" **MUST** be explicitly defined.

If a test function covers **only** the happy path (e.g., verifying a complex flow without error branches), the function itself **SHOULD** be named `TestHappyPath<Feature>`. In this case, the test function itself implicitly represents the happy path scope, so a top-level `t.Run("Happy Path", ...)` block is optional. However, side effects (emails, notifications, etc.) **MUST** be verified in distinct, nested `t.Run` blocks.

```go
func TestHandlerName(t *testing.T) {
    t.Run("Happy Path", func(t *testing.T) {
        // 1. Setup
        // Initialize mocks (QuerierStub), CoreData, and Context.
        // Use testhelpers.NewQuerierStub() and common.NewCoreData().

        // 2. Page Execution
        // Create the request and recorder, then call the handler.

        // 3. Head Data Verification
        // Assert expectations on cd.PageTitle, cd.OpenGraph, etc.

        // 4. Page Content Verification
        // Assert expectations on the response body (breadcrumbs, specific HTML elements).

        // 5. Data Consequences (Side Effects)
        // Verify DB calls (e.g., mock call history) and state changes.

        // 6. Event Bus Verification
        // Verify that expected events were published.

        // 7. Subscription Notification
        // Verify that subscriptions were triggered correctly.

        // 8. Background Workers
        // Verify indexer, post count, and external link worker behaviors if applicable.

        // 9. Notifications
        // Verify notification generation and matching.

        // 10. Email & Internal Notifications
        // Verify the content of generated emails or internal notifications.
        t.Run("Email Notifications", func(t *testing.T) {
             // ...
        })

        // 11. RSS/Atom Feeds
        // Verify feed presence and content if applicable.
    })

    t.Run("Unhappy Path - [Scenario Name]", func(t *testing.T) {
        // ...
    })
}
```

### Conventions

*   **Shared Helpers:** Shared test helpers **SHOULD** be placed in `/handlers/testframework` or `/handlers/testutil` (or `internal/testhelpers` if broadly applicable). Do not duplicate setup logic across multiple test files if it can be shared.
*   **Context Injection:** Tests relying on session data **MUST** inject a valid session into the context using `core.Store.New` and `core.ContextValues("session")`.
    ```go
    sess, _ := core.Store.New(req, core.SessionName)
    ctx := context.WithValue(req.Context(), core.ContextValues("session"), sess)
    ```
*   **Mocking:** Use `testhelpers.NewQuerierStub()` for database interaction. Manually update the stub if SQLC generates new methods.
*   **Sub-tests:** Use `t.Run()` for all distinct test cases. This mimics a `@Before` / `@After` structure where the outer function handles common setup if designed correctly, though explicit setup per `t.Run` is preferred for isolation.

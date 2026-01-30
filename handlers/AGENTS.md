# Handler Guidelines

Tasks in the handlers package must include compile time checks to ensure their concrete types satisfy the `tasks.Task` interface. Declare these using:

```go
var _ tasks.Task = (*myTaskType)(nil)
```

This pattern prevents accidental API drift if method signatures change.

## Standard Page and Task Testing System

When adding or modifying handlers, ensure that tests follow this structure to ensure comprehensive coverage. Use `t.Run()` or explicit sections (comments) to separate these concerns.

```go
func TestHandlerName(t *testing.T) {
    t.Run("Happy Path", func(t *testing.T) {
        // Setup
        // ... (Mocks, Data, Context)

        // Page execution(s -- there might be multiple)
        // ... (Call the handler)

        // Head data test (card wall, title, etc)
        // Verify cd.PageTitle, cd.OpenGraph, etc.

        // page content test (bread crumbs)
        // Verify response body content, breadcrumbs, etc.

        // data consequences test (if applicable.)
        // Verify DB side effects, calls to mocks

        // Event bus test (if applicable.)
        // Verify events published

        // Subscription notification test  (if applicable.)
        // Verify subscriptions triggered

        // Indexer & post count worker tests (if applicable)

        // External link worker tests (if applicable)

        // Background worker tests (if applicable)

        // Notification generation (matching) test  (if applicable.)

        // Email & Internal notification content test  (if applicable.)

        // RSS/Atom presence  (if applicable.)
    })

    // Unhappy path tests
    // ...
}
```

Ideally functions shared only between the test and the unhappy tests are not ideal. However something funky with `t.Run()` where we effectively have some of this in a lambda inside the function emulating a kind of `@Before`/`@after` like structure is acceptable. But it has to be go-like.

Overall shared functions between all the tests in handlers is acceptable, `/handlers/testframework` or `/handlers/testutil` or something is an acceptable location if there is no better one.

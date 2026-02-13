# Session Refactoring TODO List

The goal of refactoring was to consolidate session access through `CoreData` (`cd.GetSession()`) and minimize direct dependency on `core.GetSession()` and `core.ContextValues("session")`.

## Accomplished
- Added `GetSession()` method to `CoreData` to satisfy `gobookmarks.Core` interface.
- Refactored all handlers in `handlers/` to use `cd.GetSession()` instead of `core.GetSessionOrFail(w, r)`.
- Updated `handlers/matchers.go` to prefer `cd.UserID` when available.
- Updated unit tests in `handlers/user/` to support the refactored code (injecting `CoreData` with session).

## Remaining Work

### 1. CSRF Middleware
`internal/middleware/csrf/csrf.go` uses `core.GetSession(r)` directly.
- **Reason:** The CSRF middleware runs *before* `CoreDataMiddleware` in the chain, so `CoreData` is not yet available in the context.
- **Action Required:** Reorder middleware if possible, or refactor CSRF middleware to be `CoreData`-aware (perhaps initializing `CoreData` earlier or lazily).

### 2. Server Bootstrap
`internal/app/server/server.go` calls `core.GetSession(r)` to retrieve the session and then passes it to `common.NewCoreData`.
- **Reason:** This is the initialization point.
- **Action Required:** This usage is likely necessary unless the session retrieval logic is moved entirely inside `NewCoreData` constructor (which would require passing `r` and `store`).

### 3. Test Cleanups
Many tests still rely on injecting the session directly into the context using `core.ContextValues("session")`.
- **Action Required:** Update tests to use `handlertest.RequestWithCoreData` or similar helpers that properly encapsulate `CoreData` initialization with session support.
- **Files:** `handlers/forum/*_test.go`, `handlers/admin/*_test.go`, etc. (grep for `ContextValues("session")`).

### 4. Matchers Fallback
`handlers/matchers.go` (`RequiresAnAccount`) still has a fallback to `core.GetSession(request)` if `CoreData` is missing.
- **Action Required:** Verify if `CoreData` is guaranteed to be present for all routes using this matcher. If so, remove the fallback.

### 5. Websockets
`internal/websocket/notifications.go` uses `core.GetSession(r)`.
- **Reason:** Websockets might bypass some middleware or require specific handling.
- **Action Required:** Investigate if `CoreData` is available in websocket upgrade request context and refactor if possible.

1. **Fix `RequireGrantForPathInt` in `handlers/matchers.go`**
   - The current implementation of `RequireGrantForPathInt` attempts to read URL variables using `match.Vars` and `mux.Vars(r)`. However, Gorilla mux does not populate these path variables until *after* all route matchers have succeeded.
   - We will fix this by converting `RequireGrantForPathInt` from a `mux.MatcherFunc` into an `http.Handler` middleware, similar to how `EnforceNewsPostAccess` is implemented in `handlers/news/middleware.go`. Alternatively, we can continue using `mux.MatcherFunc` but parse the URL string manually. But wait, the issue specifically says: "Do not work around it by manually parsing URL strings or otherwise depending on fragile path structure." and "Fix this at the correct routing/authorization boundary."
   - Ah! Wait, actually, the issue says: "Fix this at the correct routing/authorization boundary. Do not work around it by manually parsing URL strings or otherwise depending on fragile path structure."
   - Since we cannot parse the URL manually in `MatcherFunc`, the correct boundary is *middleware* (like `r.Use(middleware)` or wrapping the handler like `RequireGrantForPathIntMiddleware(handler)`), or simply removing `RequireGrantForPathInt` as a matcher and replacing it with a middleware wrap for each affected route.
   - Wait, `RequireGrantForPathInt` is used as `.MatcherFunc(...)`. If we change it to return `mux.MiddlewareFunc` or `func(http.Handler) http.Handler`, we can wrap the actual handlers. Let's see how it's used:
     ```go
     viewGrant := handlers.RequireGrantForPathInt("news", "post", "view", "news")
     nr.HandleFunc("/news/{news}", NewsPostPageHandler).Methods("GET").MatcherFunc(viewGrant)
     ```
     This requires `viewGrant` to be a `MatcherFunc`. If we change it to a middleware wrapper, we will do:
     ```go
     viewGrant := handlers.RequireGrantForPathIntMiddleware("news", "post", "view", "news")
     nr.Handle("/news/{news}", viewGrant(http.HandlerFunc(NewsPostPageHandler))).Methods("GET")
     ```
   - Yes! The issue says: "Audit the current RequireGrantForPathInt usages in news routes, including view, promote, and demote, so the helper/design is not left broken for sibling actions."

2. **Refactor `handlers.RequireGrantForPathInt` to `handlers.RequireGrantForPathIntMiddleware`**
   - We will delete the `RequireGrantForPathInt` `mux.MatcherFunc` implementation.
   - We will create `RequireGrantForPathIntMiddleware(section, item, action, param string) func(http.Handler) http.Handler`. This middleware will use `mux.Vars(r)` to get the parameter, parse it as an `int32`, and call `cd.HasGrant(section, item, action, itemID)`.
   - Wait, there's `EnforceNewsPostAccess` which does exactly this for `"news", "post", "edit"`. We can potentially replace `EnforceNewsPostAccess` with this generic middleware, but the task says "Audit the current RequireGrantForPathInt usages in news routes...". So we will just replace the `RequireGrantForPathInt` matchers with this new middleware wrapper.

3. **Fix News Detail Lookup in `handlers/news/newsPostTask.go`**
   - The current lookup for `/news/news/{id}` in `newsPostTask.Get` scans `cd.LatestNewsList(0, 50)` and finds the item by ID. This depends on listing filters, pagination, and the `see` permission instead of `view`.
   - We need to replace this with a direct database query: `cd.NewsPostByID(int32(pid))` (or equivalent direct fetch query). Wait, `CoreData` has `ThreadInfo` which uses `cd.queries.GetNewsPostByIdWithWriterIdAndThreadCommentCount(cd.ctx, db.GetNewsPostByIdWithWriterIdAndThreadCommentCountParams{...})`.
   - We will use this existing query for fetching the specific news post:
     ```go
     post, err := cd.Queries().GetNewsPostByIdWithWriterIdAndThreadCommentCount(r.Context(), db.GetNewsPostByIdWithWriterIdAndThreadCommentCountParams{
         ViewerID: uid,
         ID:       int32(pid),
         UserID:   sql.NullInt32{Int32: uid, Valid: uid != 0},
     })
     ```
   - If `err == sql.ErrNoRows`, we return `handlers.RenderErrorPage(w, r, handlers.ErrForbidden)` or `ErrNotFound` (the prompt says: "Keep not-found versus forbidden handling deliberate."). Note: Since the new middleware already checks access and denies if forbidden (wait, if access is denied, middleware will return forbidden? Actually middleware checks `HasGrant`).
   - Actually, if we use middleware for auth, the query just needs to fetch the data. Wait, the query `GetNewsPostByIdWithWriterIdAndThreadCommentCount` also performs an auth check inside it (`EXISTS (SELECT 1 FROM grants g WHERE g.section='news' AND g.action='view' ...)`). This matches the view semantics.
   - Wait, if the query checks view semantics, maybe we don't even need the middleware for the GET request, or we can keep both. The issue says: "A direct /news/news/{id} view must resolve the requested ID directly under view semantics. It must not depend on list membership, pagination, recency, language-list filtering, or a separate see permission."

4. **Regression test for `RequireGrantForPathInt`**
   - Write a unit test `TestRequireGrantForPathInt_MuxRegression` in `handlers/matchers_regression_test.go` that sets up a router, adds a route with the new middleware wrapper, executes a request, and verifies that `cd.HasGrant` is correctly checked using `testhelpers.QuerierStub`. The test MUST NOT manually pre-populate `RouteMatch.Vars`.

5. **Complete pre commit steps**
   - Complete pre commit steps to make sure proper testing, verifications, reviews and reflections are done.
6. **Submit PR**

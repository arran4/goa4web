# Bug Report: `common.UserError` does not implement `Unwrap() error`

## Issue Description
When passing `common.UserError` to `handlers.RenderErrorPage(w, r, err)`, the intended HTTP status code (e.g., 403 Forbidden or 404 Not Found) is lost, resulting in a generic 500 Internal Server Error.

This happens because `RenderErrorPage` relies on `errors.As(err, &he)` to extract the embedded `*handlers.HTTPError` to determine the correct HTTP status code. However, `common.UserError` does not currently implement the `Unwrap() error` interface method. As a result, the `errors` package cannot inspect the inner error, the type assertion fails, and it falls back to the default 500 status.

## Steps to Reproduce
1. In an HTTP handler, execute:
   ```go
   handlers.RenderErrorPage(w, r, common.UserError{
       ErrorMessage: "you lack permissions",
       Err:          handlers.ErrForbidden,
   })
   ```
2. Observe the HTTP response. The rendered page will display the correct user-friendly text, but the HTTP status code returned to the client will be `500` instead of `403`.

## Suggested Fix
Implement the `Unwrap` interface method on the `UserError` struct in `core/common/usererror.go`:

```go
func (e UserError) Unwrap() error {
	return e.Err
}
```
This allows `errors.As` to properly unwrap the inner error and extract status codes when used with `handlers.RenderErrorPage`.

# Image URL Signing

Uploaded images and cached thumbnails are served through the `/images` endpoints. Access is restricted by short lived HMAC signatures embedded in the URL.

## Signing

The signing key comes from the `IMAGE_SIGN_SECRET` setting (or the file referenced by `IMAGE_SIGN_SECRET_FILE`). It is loaded during startup and used to create an `images.Signer` instance.

Functions in `pkg/images/sign.go` produce signed URLs:

- `SignedURL` returns a URL for an uploaded image using the default 24 hour expiration.
- `SignedCacheURL` does the same for cache entries.
- `SignedURLTTL` and `SignedCacheURLTTL` accept a custom `time.Duration` specifying how long the link should remain valid.

All helpers append `ts` and `sig` query parameters to the host configured in `BaseURL`. Signatures use HMAC‑SHA256 and expire after the supplied duration.

## Verification

Requests to `/images/image/{id}` and `/images/cache/{id}` pass through `verifyMiddleware` which extracts the query parameters and calls `Verify`.

Only URLs produced by `SignedURL` or `SignedCacheURL` (or `SignedRef`) will pass verification.

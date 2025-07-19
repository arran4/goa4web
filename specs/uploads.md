# Upload Providers

Goa4Web stores uploaded images through pluggable backends known as **providers**.
A provider implements the [`Provider`](../internal/upload/provider.go) interface
which defines methods for basic operations:

```
Check(ctx) error      // verify the backend is reachable
Write(ctx, name, data) error
Read(ctx, name) ([]byte, error)
```

Thumbnail caches use the [`CacheProvider`](../internal/upload/provider.go)
interface which extends `Provider` with a `Cleanup` method. `Cleanup` removes
old files until the cache size falls below a configured limit.

## Built‑in Implementations

Two provider implementations are included:

- **local** – stores files on the local filesystem. This backend is registered
  unconditionally.
- **s3** – uploads files to an S3 bucket. It is only available when the
  application is built with the `s3` build tag.

Both upload and cache providers share the same registry so the `s3` provider can
also be used for caching when compiled in.

The [`uploaddefaults` package](../internal/upload/uploaddefaults) registers all
built‑in providers at startup.

## Configuration

The following configuration keys control upload behaviour:

- `IMAGE_UPLOAD_PROVIDER` – provider name (`local` or `s3`).
- `IMAGE_UPLOAD_DIR` – directory used by the local provider.
- `IMAGE_UPLOAD_S3_URL` – bucket and prefix for the S3 provider.
- `IMAGE_CACHE_PROVIDER` – cache provider name.
- `IMAGE_CACHE_DIR` – cache directory when using the local provider.
- `IMAGE_CACHE_S3_URL` – cache bucket and prefix for the S3 provider.
- `IMAGE_CACHE_MAX_BYTES` – maximum size of the cache in bytes.
- `IMAGE_MAX_BYTES` – maximum allowed upload size.

Defaults are applied by `normalizeRuntimeConfig` in
[`config/runtime.go`](../config/runtime.go) when values are empty.

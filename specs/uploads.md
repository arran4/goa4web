# Upload Providers

Goa4Web uses a pluggable interface for file storage.

## Interfaces

- **`Provider`**: Basic Read/Write/Check operations.
- **`CacheProvider`**: Extends `Provider` with `Cleanup` for managing cache limits.

## Implementations

- **`local`**: Filesystem storage. Always available.
- **`s3`**: AWS S3 compatible storage. Requires `s3` build tag.

## Configuration

Select providers via `IMAGE_UPLOAD_PROVIDER` and `IMAGE_CACHE_PROVIDER`.
Configure specific backends using `IMAGE_UPLOAD_DIR` / `IMAGE_CACHE_DIR` (local) or `IMAGE_UPLOAD_S3_URL` / `IMAGE_CACHE_S3_URL` (S3).
Limits are set via `IMAGE_MAX_BYTES` and `IMAGE_CACHE_MAX_BYTES`.

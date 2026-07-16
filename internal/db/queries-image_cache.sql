-- name: GetImageCacheEntry :one
SELECT *
FROM image_cache_entries
WHERE id = ?
LIMIT 1;

-- name: UpsertImageCacheEntry :exec
INSERT INTO image_cache_entries (
  id,
  source_url,
  source_kind,
  status,
  created_at,
  last_used_at,
  fetched_at,
  expires_at,
  content_expires_at,
  content_type,
  size_bytes,
  width,
  height,
  checksum,
  thumbnail_id,
  error_message,
  retry_count,
  last_attempt_at,
  next_attempt_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  source_url = VALUES(source_url),
  source_kind = VALUES(source_kind),
  status = VALUES(status),
  last_used_at = VALUES(last_used_at),
  fetched_at = VALUES(fetched_at),
  expires_at = VALUES(expires_at),
  content_expires_at = VALUES(content_expires_at),
  content_type = VALUES(content_type),
  size_bytes = VALUES(size_bytes),
  width = VALUES(width),
  height = VALUES(height),
  checksum = VALUES(checksum),
  thumbnail_id = VALUES(thumbnail_id),
  error_message = VALUES(error_message),
  retry_count = VALUES(retry_count),
  last_attempt_at = VALUES(last_attempt_at),
  next_attempt_at = VALUES(next_attempt_at);

-- name: CreatePendingImageCacheEntry :exec
INSERT INTO image_cache_entries (
  id,
  source_url,
  source_kind,
  status,
  created_at,
  last_used_at,
  retry_count,
  next_attempt_at
) VALUES (?, ?, ?, 'pending', ?, ?, 0, ?)
ON DUPLICATE KEY UPDATE
  source_url = VALUES(source_url),
  source_kind = VALUES(source_kind),
  last_used_at = VALUES(last_used_at),
  next_attempt_at = VALUES(next_attempt_at);

-- name: RecordImageCacheFetchFailure :exec
UPDATE image_cache_entries
SET status = CASE WHEN retry_count + 1 >= ? THEN 'failed' ELSE 'pending' END,
    error_message = ?,
    retry_count = retry_count + 1,
    last_attempt_at = ?,
    next_attempt_at = CASE WHEN retry_count + 1 >= ? THEN NULL ELSE ? END
WHERE id = ?;

-- name: TouchImageCacheEntry :exec
UPDATE image_cache_entries
SET last_used_at = ?
WHERE id = ?;

-- name: DeleteImageCacheEntry :exec
DELETE FROM image_cache_entries
WHERE id = ?;

-- name: ListDuePendingImageCacheEntries :many
SELECT *
FROM image_cache_entries
WHERE source_kind = 'remote'
  AND status = 'pending'
  AND retry_count < ?
  AND (next_attempt_at IS NULL OR next_attempt_at <= ?)
ORDER BY created_at ASC
LIMIT ?;

-- name: ListExpiredExternalImageCacheEntries :many
SELECT *
FROM image_cache_entries
WHERE source_kind = 'remote'
  AND status = 'ready'
  AND expires_at IS NOT NULL
  AND expires_at <= ?
ORDER BY expires_at ASC
LIMIT ?;

-- name: ListOldestUsedImageCacheEntries :many
SELECT *
FROM image_cache_entries
WHERE status = 'ready'
ORDER BY COALESCE(last_used_at, created_at) ASC
LIMIT ?;

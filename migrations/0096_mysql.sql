-- +goose Up
-- Step 1: Drop the old 255-byte unique prefix index and modify url column to text.
-- Dropping the unique index first ensures subsequent canonicalization updates do not collide before duplicate consolidation.
ALTER TABLE external_links DROP INDEX external_links_url_idx, MODIFY COLUMN url text NOT NULL;

-- Step 2: Clean known tracking parameters from pre-existing URLs while preserving signed URLs, functional parameters, and query delimiters.
UPDATE external_links
SET url = REGEXP_REPLACE(
    REGEXP_REPLACE(
        REGEXP_REPLACE(
            url,
            '&(?i)(utm_[a-zA-Z0-9_]*|fbclid|gclid|gbraid|wbraid|mc_cid|mc_eid|igshid|msclkid|twclid|yclid|click_id|clickid|_hsenc|_hsmi|mkt_tok)(=[^&#]*)?',
            ''
        ),
        '\\?(?i)(utm_[a-zA-Z0-9_]*|fbclid|gclid|gbraid|wbraid|mc_cid|mc_eid|igshid|msclkid|twclid|yclid|click_id|clickid|_hsenc|_hsmi|mkt_tok)(=[^&#]*)?&',
        '?'
    ),
    '\\?(?i)(utm_[a-zA-Z0-9_]*|fbclid|gclid|gbraid|wbraid|mc_cid|mc_eid|igshid|msclkid|twclid|yclid|click_id|clickid|_hsenc|_hsmi|mkt_tok)(=[^&#]*)?(#.*)?$',
    '$3'
)
WHERE url REGEXP '^(?i)https?://'
  AND url REGEXP '([?&])(?i)(utm_[a-zA-Z0-9_]*|fbclid|gclid|gbraid|wbraid|mc_cid|mc_eid|igshid|msclkid|twclid|yclid|click_id|clickid|_hsenc|_hsmi|mkt_tok)(=[^&#]*)?(&|#|$)'
  AND url NOT REGEXP '([?&])(?i)(x-amz-signature|x-amz-credential|signature|sig|hash|hmac|x-goog-signature|x-ms-signature)(=|&|#|$)';

-- Step 3: Consolidate duplicate pre-existing URLs resulting from tracking cleanup
CREATE TEMPORARY TABLE IF NOT EXISTS _merged_links AS
SELECT 
    MIN(id) AS keep_id,
    url,
    SUM(clicks) AS total_clicks,
    MAX(card_title) AS card_title,
    MAX(card_description) AS card_description,
    MAX(card_image) AS card_image,
    MAX(card_image_cache) AS card_image_cache,
    MAX(favicon_cache) AS favicon_cache,
    MAX(card_duration) AS card_duration,
    MAX(card_upload_date) AS card_upload_date,
    MAX(card_author) AS card_author,
    MIN(created_at) AS created_at,
    MAX(updated_at) AS updated_at
FROM external_links
GROUP BY url
HAVING COUNT(*) > 1;

UPDATE external_links el
JOIN _merged_links ml ON el.id = ml.keep_id
SET 
    el.clicks = ml.total_clicks,
    el.card_title = COALESCE(NULLIF(el.card_title, ''), ml.card_title),
    el.card_description = COALESCE(NULLIF(el.card_description, ''), ml.card_description),
    el.card_image = COALESCE(NULLIF(el.card_image, ''), ml.card_image),
    el.card_image_cache = COALESCE(NULLIF(el.card_image_cache, ''), ml.card_image_cache),
    el.favicon_cache = COALESCE(NULLIF(el.favicon_cache, ''), ml.favicon_cache),
    el.card_duration = COALESCE(NULLIF(el.card_duration, ''), ml.card_duration),
    el.card_upload_date = COALESCE(NULLIF(el.card_upload_date, ''), ml.card_upload_date),
    el.card_author = COALESCE(NULLIF(el.card_author, ''), ml.card_author),
    el.created_at = ml.created_at,
    el.updated_at = ml.updated_at;

DELETE el FROM external_links el
JOIN _merged_links ml ON el.url = ml.url AND el.id != ml.keep_id;

DROP TEMPORARY TABLE IF EXISTS _merged_links;

-- Step 4: Add generated stored url_hash and full-URL unique index
ALTER TABLE external_links ADD COLUMN url_hash binary(32) GENERATED ALWAYS AS (unhex(sha2(url, 256))) STORED NOT NULL, ADD UNIQUE KEY external_links_url_hash_idx (url_hash);

-- Step 5: Advance schema version
UPDATE schema_version SET version = 96;

-- +goose Down
-- Guarded rollback: fail if any URL exceeds 255 bytes before attempting destructive narrowing
SELECT IF(IFNULL(MAX(LENGTH(url)), 0) > 255, (SELECT table_name FROM information_schema.tables LIMIT 2), 1) FROM external_links INTO @guard;
ALTER TABLE external_links DROP INDEX external_links_url_hash_idx;
ALTER TABLE external_links DROP COLUMN url_hash;
ALTER TABLE external_links MODIFY COLUMN url tinytext NOT NULL;
ALTER TABLE external_links ADD UNIQUE KEY external_links_url_idx (url(255));
UPDATE schema_version SET version = 95;

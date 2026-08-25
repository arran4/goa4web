import sys

# Patch MySQL queries
with open("internal/db/queries-comments.sql", "r") as f:
    content = f.read()

bad_mysql = "SET text = sqlc.arg(text), written = sqlc.arg(written)"
good_mysql = "SET text = CONCAT(c.text, '\\n\\n[hr]\\n\\n', sqlc.arg(text)), written = sqlc.arg(written)"

content = content.replace(bad_mysql, good_mysql)
with open("internal/db/queries-comments.sql", "w") as f:
    f.write(content)

# Patch SQLite queries
with open("internal/dbsqlite_queries/queries-comments.sql", "r") as f:
    content = f.read()

bad_sqlite = "SET text = sqlc.arg(text), written = sqlc.arg(written)"
good_sqlite = "SET text = comments.text || '\\n\\n[hr]\\n\\n' || sqlc.arg(text), written = sqlc.arg(written)"

content = content.replace(bad_sqlite, good_sqlite)
with open("internal/dbsqlite_queries/queries-comments.sql", "w") as f:
    f.write(content)

# Add read marker query for UI check
read_marker_query = """

-- name: HasOtherUserReadItemAtOrBeyond :one
SELECT EXISTS (
    SELECT 1 FROM content_read_markers crm
    WHERE crm.item = sqlc.arg(item)
      AND crm.item_id = sqlc.arg(item_id)
      AND crm.user_id != sqlc.arg(user_id)
      AND crm.last_comment_id >= sqlc.arg(last_comment_id)
) AS has_read;
"""

with open("internal/db/queries-read_markers.sql", "a") as f:
    f.write(read_marker_query)

with open("internal/dbsqlite_queries/queries-read_markers.sql", "a") as f:
    f.write(read_marker_query)

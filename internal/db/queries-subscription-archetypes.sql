-- name: CreateSubscriptionArchetype :exec
INSERT INTO role_subscription_archetypes (role_id, archetype_name, pattern, method)
VALUES (?, ?, ?, ?);

-- name: GetSubscriptionArchetypesByRole :many
SELECT * FROM role_subscription_archetypes
WHERE role_id = ?;

-- name: DeleteSubscriptionArchetypesByRoleAndName :exec
DELETE FROM role_subscription_archetypes
WHERE role_id = ? AND archetype_name = ?;

-- name: ListSubscriptionArchetypes :many
SELECT * FROM role_subscription_archetypes
ORDER BY role_id, archetype_name;

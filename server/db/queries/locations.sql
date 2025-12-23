-- name: ListLocations :many
SELECT id, name, COALESCE(note, '') AS note
FROM locations
ORDER BY name;

-- name: CreateLocation :one
INSERT INTO locations(name, note)
VALUES($1, $2)
RETURNING id;

-- name: UpdateLocation :exec
UPDATE locations
SET name = $1,
    note = $2
WHERE id = $3;

-- name: DeleteLocation :exec
DELETE FROM locations
WHERE id = $1;

-- name: CountLocations :one
SELECT COUNT(*)
FROM locations;

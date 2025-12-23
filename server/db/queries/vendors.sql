-- name: ListVendors :many
SELECT id, name, COALESCE(country, '') AS country
FROM vendors
ORDER BY name;

-- name: CreateVendor :one
INSERT INTO vendors(name, country)
VALUES($1, $2)
RETURNING id;

-- name: UpdateVendor :exec
UPDATE vendors
SET name = $1,
    country = $2
WHERE id = $3;

-- name: DeleteVendor :exec
DELETE FROM vendors
WHERE id = $1;

-- name: CountVendors :one
SELECT COUNT(*)
FROM vendors;

-- name: GetFirstVendorID :one
SELECT id
FROM vendors
ORDER BY id
LIMIT 1;

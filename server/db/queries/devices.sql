-- name: ListDevices :many
SELECT d.id,
       v.name AS vendor_name,
       m.name AS model_name,
       COALESCE(l.name, '') AS location_name,
       COALESCE(d.serial_number, '') AS serial_number,
       COALESCE(d.inventory_number, '') AS inventory_number,
       d.status,
       COALESCE(to_char(d.installed_at, 'YYYY-MM-DD'), '') AS installed_at
FROM devices d
JOIN models m ON m.id = d.model_id
JOIN vendors v ON v.id = m.vendor_id
LEFT JOIN locations l ON l.id = d.location_id
WHERE (
  $1::text = ''
  OR d.serial_number ILIKE '%' || $1 || '%'
  OR d.inventory_number ILIKE '%' || $1 || '%'
  OR m.name ILIKE '%' || $1 || '%'
  OR v.name ILIKE '%' || $1 || '%'
  OR d.status ILIKE '%' || $1 || '%'
)
ORDER BY d.id DESC;

-- name: CreateDevice :one
INSERT INTO devices(model_id, location_id, serial_number, inventory_number, status, installed_at, description)
VALUES($1, $2, $3, $4, $5, $6, $7)
RETURNING id;

-- name: GetDeviceByID :one
SELECT id,
       model_id,
       location_id,
       COALESCE(serial_number, '') AS serial_number,
       COALESCE(inventory_number, '') AS inventory_number,
       status,
       COALESCE(to_char(installed_at, 'YYYY-MM-DD'), '') AS installed_at,
       COALESCE(description, '') AS description
FROM devices
WHERE id = $1;

-- name: UpdateDevice :exec
UPDATE devices
SET model_id = $1,
    location_id = $2,
    serial_number = $3,
    inventory_number = $4,
    status = $5,
    installed_at = $6,
    description = $7
WHERE id = $8;

  -- name: DeleteDevice :exec
DELETE FROM devices
WHERE id = $1;

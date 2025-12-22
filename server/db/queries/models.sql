SELECT m.id,
       m.vendor_id,
       v.name AS vendor_name,
       m.name,
       COALESCE(m.device_type, '') AS device_type
FROM models m
JOIN vendors v ON v.id = m.vendor_id
ORDER BY v.name, m.name;

INSERT INTO models(vendor_id, name, device_type)
VALUES($1, $2, $3)
RETURNING id;

UPDATE models
SET vendor_id = $1,
    name = $2,
    device_type = $3
WHERE id = $4;

DELETE FROM models
WHERE id = $1;

SELECT COUNT(*)
FROM models;

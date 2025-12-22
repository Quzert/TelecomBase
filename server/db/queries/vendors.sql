SELECT id, name, COALESCE(country, '') AS country
FROM vendors
ORDER BY name;

INSERT INTO vendors(name, country)
VALUES($1, $2)
RETURNING id;

UPDATE vendors
SET name = $1,
    country = $2
WHERE id = $3;

DELETE FROM vendors
WHERE id = $1;

SELECT COUNT(*)
FROM vendors;

SELECT id
FROM vendors
ORDER BY id
LIMIT 1;

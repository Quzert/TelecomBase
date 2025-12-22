SELECT id, name, COALESCE(note, '') AS note
FROM locations
ORDER BY name;

INSERT INTO locations(name, note)
VALUES($1, $2)
RETURNING id;

UPDATE locations
SET name = $1,
    note = $2
WHERE id = $3;

DELETE FROM locations
WHERE id = $1;

SELECT COUNT(*)
FROM locations;

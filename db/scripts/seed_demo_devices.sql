-- Демо-устройства для TelecomBase (ручной запуск).
-- Добавляет несколько устройств, чтобы главная таблица не была пустой.
-- Идемпотентно: используется WHERE NOT EXISTS по serial_number.

BEGIN;

WITH m AS (
    SELECT m.id AS model_id
    FROM models m
    JOIN vendors v ON v.id = m.vendor_id
    WHERE v.name = 'Cisco' AND m.name = 'ISR 4321'
    LIMIT 1
), l AS (
    SELECT id AS location_id FROM locations WHERE name = 'Main office' LIMIT 1
)
INSERT INTO devices(model_id, location_id, serial_number, inventory_number, status, installed_at, description)
SELECT m.model_id, l.location_id, 'DEMO-SN-001', 'INV-001', 'active', CURRENT_DATE - 120, 'Demo router'
FROM m, l
WHERE NOT EXISTS (SELECT 1 FROM devices WHERE serial_number = 'DEMO-SN-001');

WITH m AS (
    SELECT m.id AS model_id
    FROM models m
    JOIN vendors v ON v.id = m.vendor_id
    WHERE v.name = 'Cisco' AND m.name = 'Catalyst 2960'
    LIMIT 1
), l AS (
    SELECT id AS location_id FROM locations WHERE name = 'Data center' LIMIT 1
)
INSERT INTO devices(model_id, location_id, serial_number, inventory_number, status, installed_at, description)
SELECT m.model_id, l.location_id, 'DEMO-SN-002', 'INV-002', 'active', CURRENT_DATE - 45, 'Demo switch'
FROM m, l
WHERE NOT EXISTS (SELECT 1 FROM devices WHERE serial_number = 'DEMO-SN-002');

WITH m AS (
    SELECT m.id AS model_id
    FROM models m
    JOIN vendors v ON v.id = m.vendor_id
    WHERE v.name = 'Juniper' AND m.name = 'MX480'
    LIMIT 1
), l AS (
    SELECT id AS location_id FROM locations WHERE name = 'Data center' LIMIT 1
)
INSERT INTO devices(model_id, location_id, serial_number, inventory_number, status, installed_at, description)
SELECT m.model_id, l.location_id, 'DEMO-SN-003', 'INV-003', 'active', CURRENT_DATE - 365, 'Demo core router'
FROM m, l
WHERE NOT EXISTS (SELECT 1 FROM devices WHERE serial_number = 'DEMO-SN-003');

WITH m AS (
    SELECT m.id AS model_id
    FROM models m
    JOIN vendors v ON v.id = m.vendor_id
    WHERE v.name = 'Juniper' AND m.name = 'SRX300'
    LIMIT 1
), l AS (
    SELECT id AS location_id FROM locations WHERE name = 'Branch office' LIMIT 1
)
INSERT INTO devices(model_id, location_id, serial_number, inventory_number, status, installed_at, description)
SELECT m.model_id, l.location_id, 'DEMO-SN-004', 'INV-004', 'active', CURRENT_DATE - 200, 'Demo firewall'
FROM m, l
WHERE NOT EXISTS (SELECT 1 FROM devices WHERE serial_number = 'DEMO-SN-004');

WITH m AS (
    SELECT m.id AS model_id
    FROM models m
    JOIN vendors v ON v.id = m.vendor_id
    WHERE v.name = 'Huawei' AND m.name = 'S5720'
    LIMIT 1
), l AS (
    SELECT id AS location_id FROM locations WHERE name = 'Branch office' LIMIT 1
)
INSERT INTO devices(model_id, location_id, serial_number, inventory_number, status, installed_at, description)
SELECT m.model_id, l.location_id, 'DEMO-SN-005', 'INV-005', 'active', CURRENT_DATE - 20, 'Demo access switch'
FROM m, l
WHERE NOT EXISTS (SELECT 1 FROM devices WHERE serial_number = 'DEMO-SN-005');

WITH m AS (
    SELECT m.id AS model_id
    FROM models m
    JOIN vendors v ON v.id = m.vendor_id
    WHERE v.name = 'Huawei' AND m.name = 'AR169'
    LIMIT 1
), l AS (
    SELECT id AS location_id FROM locations WHERE name = 'Main office' LIMIT 1
)
INSERT INTO devices(model_id, location_id, serial_number, inventory_number, status, installed_at, description)
SELECT m.model_id, l.location_id, 'DEMO-SN-006', 'INV-006', 'maintenance', CURRENT_DATE - 10, 'Demo router (maintenance)'
FROM m, l
WHERE NOT EXISTS (SELECT 1 FROM devices WHERE serial_number = 'DEMO-SN-006');

COMMIT;

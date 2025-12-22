-- Демо-данные для TelecomBase (ручной запуск).
-- Добавляет 3 производителя, 6 моделей, 3 локации.
-- Идемпотентно: используется WHERE NOT EXISTS.

BEGIN;

INSERT INTO vendors(name, country)
SELECT 'Cisco', 'US'
WHERE NOT EXISTS (SELECT 1 FROM vendors WHERE name = 'Cisco');

INSERT INTO vendors(name, country)
SELECT 'Juniper', 'US'
WHERE NOT EXISTS (SELECT 1 FROM vendors WHERE name = 'Juniper');

INSERT INTO vendors(name, country)
SELECT 'Huawei', 'CN'
WHERE NOT EXISTS (SELECT 1 FROM vendors WHERE name = 'Huawei');

INSERT INTO locations(name, note)
SELECT 'Main office', 'Default location'
WHERE NOT EXISTS (SELECT 1 FROM locations WHERE name = 'Main office');

INSERT INTO locations(name, note)
SELECT 'Branch office', 'Remote branch'
WHERE NOT EXISTS (SELECT 1 FROM locations WHERE name = 'Branch office');

INSERT INTO locations(name, note)
SELECT 'Data center', 'Rack area'
WHERE NOT EXISTS (SELECT 1 FROM locations WHERE name = 'Data center');

WITH v AS (SELECT id FROM vendors WHERE name = 'Cisco' LIMIT 1)
INSERT INTO models(vendor_id, name, device_type)
SELECT v.id, 'ISR 4321', 'router'
FROM v
WHERE NOT EXISTS (SELECT 1 FROM models m WHERE m.vendor_id = v.id AND m.name = 'ISR 4321');

WITH v AS (SELECT id FROM vendors WHERE name = 'Cisco' LIMIT 1)
INSERT INTO models(vendor_id, name, device_type)
SELECT v.id, 'Catalyst 2960', 'switch'
FROM v
WHERE NOT EXISTS (SELECT 1 FROM models m WHERE m.vendor_id = v.id AND m.name = 'Catalyst 2960');

WITH v AS (SELECT id FROM vendors WHERE name = 'Juniper' LIMIT 1)
INSERT INTO models(vendor_id, name, device_type)
SELECT v.id, 'MX480', 'router'
FROM v
WHERE NOT EXISTS (SELECT 1 FROM models m WHERE m.vendor_id = v.id AND m.name = 'MX480');

WITH v AS (SELECT id FROM vendors WHERE name = 'Juniper' LIMIT 1)
INSERT INTO models(vendor_id, name, device_type)
SELECT v.id, 'SRX300', 'firewall'
FROM v
WHERE NOT EXISTS (SELECT 1 FROM models m WHERE m.vendor_id = v.id AND m.name = 'SRX300');

WITH v AS (SELECT id FROM vendors WHERE name = 'Huawei' LIMIT 1)
INSERT INTO models(vendor_id, name, device_type)
SELECT v.id, 'S5720', 'switch'
FROM v
WHERE NOT EXISTS (SELECT 1 FROM models m WHERE m.vendor_id = v.id AND m.name = 'S5720');

WITH v AS (SELECT id FROM vendors WHERE name = 'Huawei' LIMIT 1)
INSERT INTO models(vendor_id, name, device_type)
SELECT v.id, 'AR169', 'router'
FROM v
WHERE NOT EXISTS (SELECT 1 FROM models m WHERE m.vendor_id = v.id AND m.name = 'AR169');

COMMIT;

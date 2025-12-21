-- Add user approval gate.
-- New users are created as approved=false, and must be approved by admin.

BEGIN;

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS approved BOOLEAN;

UPDATE users
SET approved = TRUE
WHERE approved IS NULL;

ALTER TABLE users
    ALTER COLUMN approved SET DEFAULT FALSE;

ALTER TABLE users
    ALTER COLUMN approved SET NOT NULL;

COMMIT;

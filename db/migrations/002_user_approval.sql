-- Добавление подтверждения пользователя.
-- Новые пользователи создаются с approved=false и должны быть подтверждены администратором.

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

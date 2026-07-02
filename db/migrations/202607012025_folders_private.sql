-- migrate:up
ALTER TABLE folders ADD COLUMN private BOOLEAN NOT NULL DEFAULT TRUE;
-- migrate:down
ALTER TABLE folders DROP COLUMN IF EXISTS private;

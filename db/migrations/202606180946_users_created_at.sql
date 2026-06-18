-- migrate:up
ALTER TABLE auth.users
ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- migrate:down
ALTER TABLE auth.users DROP COLUMN IF EXISTS created_at;

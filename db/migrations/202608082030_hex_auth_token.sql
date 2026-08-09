-- migrate:up
ALTER TABLE auth.sessions
ALTER COLUMN token SET DEFAULT encode(public.gen_random_bytes(32), 'hex');

-- migrate:down
ALTER TABLE auth.sessions
ALTER COLUMN token SET DEFAULT encode(public.gen_random_bytes(32), 'base64');

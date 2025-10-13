-- migrate:up
alter table auth.users
add column oauth_google_name text;

UPDATE auth.users
SET
    oauth_google_name = oauth_google_email,
    oauth_google_email = NULL
WHERE oauth_google_email NOT LIKE '%@%';

-- migrate:down

-- migrate:up
alter table auth.users
add column mod_perms boolean not null default false;

-- migrate:down

-- migrate:up
alter table studysets
add column created_at timestamptz default now();

update studysets
set created_at = updated_at;

-- migrate:down

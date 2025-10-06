-- migrate:up
alter table studysets
add column subject_id text references subjects (id) on delete set null;
-- migrate:down

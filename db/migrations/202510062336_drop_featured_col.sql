-- migrate:up
alter table studysets
drop column featured;
-- migrate:down

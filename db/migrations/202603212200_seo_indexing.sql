-- migrate:up
ALTER TABLE studysets ADD COLUMN seo_indexing_approved BOOLEAN NOT NULL DEFAULT FALSE;
-- migrate:down
ALTER TABLE studysets DROP COLUMN seo_indexing_approved;
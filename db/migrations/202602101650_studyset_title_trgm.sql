-- migrate:up
CREATE INDEX studysets_title_trgm_idx ON studysets USING GIN (lower(title) gin_trgm_ops);

-- migrate:down
DROP INDEX studysets_title_trgm_idx;

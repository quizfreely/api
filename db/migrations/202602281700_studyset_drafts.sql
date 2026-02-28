-- migrate:up
ALTER TABLE public.studysets ADD COLUMN draft boolean NOT NULL DEFAULT false;

-- migrate:down
ALTER TABLE public.studysets DROP COLUMN draft;
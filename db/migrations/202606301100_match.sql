-- migrate:up
CREATE TABLE public.match_activities (
    id uuid DEFAULT gen_random_uuid() NOT NULL PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    duration_ms integer NOT NULL,
    end_timestamp timestamp with time zone DEFAULT now() NOT NULL
);
GRANT SELECT, INSERT, UPDATE, DELETE ON public.match_activities TO quizfreely_api;

CREATE TABLE public.match_activity_studysets (
    match_id uuid NOT NULL REFERENCES public.match_activities(id) ON DELETE CASCADE,
    studyset_id uuid NOT NULL REFERENCES public.studysets(id) ON DELETE CASCADE,
    PRIMARY KEY (match_id, studyset_id)
);
GRANT SELECT, INSERT, UPDATE, DELETE ON match_activity_studysets TO quizfreely_api;

ALTER TABLE public.review_events
ADD COLUMN match_activity_id uuid REFERENCES public.match_activities(id) ON DELETE CASCADE;

ALTER TABLE public.review_events
ALTER COLUMN answer_with DROP NOT NULL;

-- migrate:down
ALTER TABLE public.review_events ALTER COLUMN answer_with SET NOT NULL;
ALTER TABLE public.review_events DROP COLUMN IF EXISTS match_activity_id;
DROP TABLE IF EXISTS public.match_activity_studysets;
DROP TABLE IF EXISTS public.match_activities;


-- migrate:up
DROP TABLE IF EXISTS public.term_confusion_pairs;

CREATE TYPE public.review_activity_type_enum AS ENUM (
    'PRACTICE_TEST',
    'MATCH'
);

CREATE TABLE public.review_events (
    id uuid DEFAULT gen_random_uuid() NOT NULL PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    term_id uuid NOT NULL REFERENCES public.terms(id) ON DELETE CASCADE,
    practice_test_question_id uuid REFERENCES public.practice_test_questions(id) ON DELETE CASCADE,
    correct boolean NOT NULL,
    answer_with public.answer_with_enum NOT NULL,
    timestamp timestamp with time zone DEFAULT now() NOT NULL,
    answered_term_id uuid REFERENCES public.terms(id) ON DELETE CASCADE,
    practice_test_question_type public.question_type,
    review_activity_type public.review_activity_type_enum NOT NULL,
    answered_string text
);
GRANT SELECT, INSERT, UPDATE, DELETE ON public.review_events TO quizfreely_api;

CREATE INDEX idx_review_events_user_id ON public.review_events(user_id);
CREATE INDEX idx_review_events_term_id ON public.review_events(term_id);
CREATE INDEX idx_review_events_practice_test_question_id ON public.review_events(practice_test_question_id);

-- migrate:down
DROP TABLE IF EXISTS public.review_events;
DROP TYPE IF EXISTS public.review_activity_type_enum;

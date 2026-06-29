-- migrate:up

CREATE TYPE public.question_type AS ENUM ('MCQ', 'TFQ', 'FRQ');

CREATE TABLE public.practice_test_questions (
    id uuid DEFAULT gen_random_uuid() NOT NULL PRIMARY KEY,
    practice_test_id uuid NOT NULL REFERENCES public.practice_tests(id) ON DELETE CASCADE,
    term_id uuid NOT NULL REFERENCES public.terms(id) ON DELETE CASCADE,
    term_snapshot text NOT NULL,
    def_snapshot text NOT NULL,
    type public.question_type NOT NULL,
    answer_with public.answer_with_enum NOT NULL,
    correct boolean NOT NULL,
    position integer NOT NULL,
    data jsonb NOT NULL,
    UNIQUE (practice_test_id, position)
);

-- Migrate data
INSERT INTO public.practice_test_questions (
    practice_test_id, term_id, term_snapshot, def_snapshot, type, answer_with, correct, position, data
)
SELECT
    pt.id,
    (q->COALESCE(sub.type_lower, 'mcq')->'term'->>'id')::uuid,
    q->COALESCE(sub.type_lower, 'mcq')->'term'->>'term',
    q->COALESCE(sub.type_lower, 'mcq')->'term'->>'def',
    sub.type_upper,
    (q->COALESCE(sub.type_lower, 'mcq')->>'answerWith')::public.answer_with_enum,
    CASE
        WHEN sub.type_upper = 'FRQ' THEN
            COALESCE((q->'frq'->>'correct')::boolean, false) OR COALESCE((q->'frq'->>'userMarkedCorrect')::boolean, false)
        ELSE
            COALESCE((q->sub.type_lower->>'correct')::boolean, false)
    END,
    pos::int - 1,
    CASE
        WHEN sub.type_upper = 'MCQ' THEN
            jsonb_build_object(
                'distractors', (
                    SELECT jsonb_agg(
                        jsonb_build_object(
                            'id', d->>'id',
                            'termSnapshot', d->>'term',
                            'defSnapshot', d->>'def'
                        )
                    )
                    FROM jsonb_array_elements(q->'mcq'->'distractors') d
                ),
                'correctChoiceIndex', (q->'mcq'->>'correctChoiceIndex')::int,
                'answeredIndex', (q->'mcq'->>'answeredIndex')::int
            )
        WHEN sub.type_upper = 'TFQ' THEN
            jsonb_build_object(
                'answeredBool', (q->'tfq'->>'answeredBool')::boolean,
                'distractor', CASE
                    WHEN q->'tfq'->'distractor' IS NOT NULL AND q->'tfq'->'distractor' != 'null'::jsonb THEN
                        jsonb_build_object(
                            'id', q->'tfq'->'distractor'->>'id',
                            'termSnapshot', q->'tfq'->'distractor'->>'term',
                            'defSnapshot', q->'tfq'->'distractor'->>'def'
                        )
                    ELSE NULL
                END
            )
        WHEN sub.type_upper = 'FRQ' THEN
            jsonb_build_object(
                'answeredString', q->'frq'->>'answeredString',
                'userMarkedCorrect', COALESCE((q->'frq'->>'userMarkedCorrect')::boolean, false)
            )
    END
FROM public.practice_tests pt
CROSS JOIN LATERAL jsonb_array_elements(pt.questions) WITH ORDINALITY AS q_arr(q, pos)
CROSS JOIN LATERAL (
    SELECT
        CASE
            WHEN q ? 'mcq' THEN 'mcq'
            WHEN q ? 'tfq' THEN 'tfq'
            WHEN q ? 'frq' THEN 'frq'
        END as type_lower,
        CASE
            WHEN q ? 'mcq' THEN 'MCQ'::public.question_type
            WHEN q ? 'tfq' THEN 'TFQ'::public.question_type
            WHEN q ? 'frq' THEN 'FRQ'::public.question_type
        END as type_upper
) sub;

ALTER TABLE public.practice_tests DROP COLUMN questions;
DROP TABLE IF EXISTS public.practice_test_question_terms;
DROP TABLE IF EXISTS public.practice_test_distractor_terms;
DROP TABLE IF EXISTS public.term_progress_history;
ALTER TABLE public.term_progress DROP COLUMN IF EXISTS term_leitner_system_box;
ALTER TABLE public.term_progress DROP COLUMN IF EXISTS def_leitner_system_box;

GRANT SELECT, INSERT, UPDATE, DELETE ON public.practice_test_questions TO quizfreely_api;

-- migrate:down
DROP TABLE public.practice_test_questions;
DROP TYPE public.question_type;

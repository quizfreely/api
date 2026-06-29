-- migrate:up

-- Rename termSnapshot -> term and defSnapshot -> def in the data jsonb column
-- for MCQ distractors array items
UPDATE public.practice_test_questions
SET data = jsonb_set(
    data,
    '{distractors}',
    (
        SELECT jsonb_agg(
            jsonb_build_object(
                'id', d->>'id',
                'term', d->>'termSnapshot',
                'def', d->>'defSnapshot'
            )
        )
        FROM jsonb_array_elements(data->'distractors') d
    )
)
WHERE data ? 'distractors';

-- for TFQ distractor object
UPDATE public.practice_test_questions
SET data = jsonb_set(
    data,
    '{distractor}',
    jsonb_build_object(
        'id', data->'distractor'->>'id',
        'term', data->'distractor'->>'termSnapshot',
        'def', data->'distractor'->>'defSnapshot'
    )
)
WHERE data ? 'distractor';

-- migrate:down

-- Rename term -> termSnapshot and def -> defSnapshot in the data jsonb column
-- for MCQ distractors array items
UPDATE public.practice_test_questions
SET data = jsonb_set(
    data,
    '{distractors}',
    (
        SELECT jsonb_agg(
            jsonb_build_object(
                'id', d->>'id',
                'termSnapshot', d->>'term',
                'defSnapshot', d->>'def'
            )
        )
        FROM jsonb_array_elements(data->'distractors') d
    )
)
WHERE data ? 'distractors';

-- for TFQ distractor object
UPDATE public.practice_test_questions
SET data = jsonb_set(
    data,
    '{distractor}',
    jsonb_build_object(
        'id', data->'distractor'->>'id',
        'termSnapshot', data->'distractor'->>'term',
        'defSnapshot', data->'distractor'->>'def'
    )
)
WHERE data ? 'distractor';

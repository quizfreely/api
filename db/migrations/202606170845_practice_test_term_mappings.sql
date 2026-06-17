-- migrate:up
CREATE TABLE practice_test_question_terms (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    practice_test_id uuid NOT NULL REFERENCES practice_tests(id) ON DELETE CASCADE,
    term_id uuid NOT NULL REFERENCES terms(id) ON DELETE CASCADE
);

CREATE TABLE practice_test_distractor_terms (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    practice_test_id uuid NOT NULL REFERENCES practice_tests(id) ON DELETE CASCADE,
    term_id uuid NOT NULL REFERENCES terms(id) ON DELETE CASCADE
);

CREATE INDEX idx_ptqt_term_id ON practice_test_question_terms(term_id);
CREATE INDEX idx_ptqt_practice_test_id ON practice_test_question_terms(practice_test_id);
CREATE INDEX idx_ptdt_term_id ON practice_test_distractor_terms(term_id);
CREATE INDEX idx_ptdt_practice_test_id ON practice_test_distractor_terms(practice_test_id);

GRANT SELECT, INSERT, UPDATE, DELETE ON practice_test_question_terms TO quizfreely_api;
GRANT SELECT, INSERT, UPDATE, DELETE ON practice_test_distractor_terms TO quizfreely_api;

-- Populate existing data for MCQ questions
INSERT INTO practice_test_question_terms (practice_test_id, term_id)
SELECT pt.id, (q->'mcq'->'term'->>'id')::uuid
FROM practice_tests pt, jsonb_array_elements(pt.questions) AS q
WHERE jsonb_typeof(q->'mcq'->'term') = 'object'
  AND q->'mcq'->'term'->>'id' ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
  AND EXISTS (SELECT 1 FROM terms WHERE id = (q->'mcq'->'term'->>'id')::uuid);

-- Populate existing data for True/False questions
INSERT INTO practice_test_question_terms (practice_test_id, term_id)
SELECT pt.id, (q->'trueFalseQuestion'->'term'->>'id')::uuid
FROM practice_tests pt, jsonb_array_elements(pt.questions) AS q
WHERE jsonb_typeof(q->'trueFalseQuestion'->'term') = 'object'
  AND q->'trueFalseQuestion'->'term'->>'id' ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
  AND EXISTS (SELECT 1 FROM terms WHERE id = (q->'trueFalseQuestion'->'term'->>'id')::uuid);

-- Populate existing data for FRQ questions
INSERT INTO practice_test_question_terms (practice_test_id, term_id)
SELECT pt.id, (q->'frq'->'term'->>'id')::uuid
FROM practice_tests pt, jsonb_array_elements(pt.questions) AS q
WHERE jsonb_typeof(q->'frq'->'term') = 'object'
  AND q->'frq'->'term'->>'id' ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
  AND EXISTS (SELECT 1 FROM terms WHERE id = (q->'frq'->'term'->>'id')::uuid);

-- Populate existing data for MCQ distractors
INSERT INTO practice_test_distractor_terms (practice_test_id, term_id)
SELECT pt.id, (d->>'id')::uuid
FROM practice_tests pt,
     jsonb_array_elements(pt.questions) AS q,
     jsonb_array_elements(CASE WHEN jsonb_typeof(q->'mcq'->'distractors') = 'array' THEN q->'mcq'->'distractors' ELSE '[]'::jsonb END) AS d
WHERE d->>'id' ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
  AND EXISTS (SELECT 1 FROM terms WHERE id = (d->>'id')::uuid);

-- Populate existing data for True/False distractors
INSERT INTO practice_test_distractor_terms (practice_test_id, term_id)
SELECT pt.id, (q->'trueFalseQuestion'->'distractor'->>'id')::uuid
FROM practice_tests pt, jsonb_array_elements(pt.questions) AS q
WHERE jsonb_typeof(q->'trueFalseQuestion'->'distractor') = 'object'
  AND q->'trueFalseQuestion'->'distractor'->>'id' ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
  AND EXISTS (SELECT 1 FROM terms WHERE id = (q->'trueFalseQuestion'->'distractor'->>'id')::uuid);

-- migrate:down
DROP TABLE practice_test_distractor_terms;
DROP TABLE practice_test_question_terms;

-- migrate:up
ALTER TABLE practice_tests DROP COLUMN IF EXISTS studyset_id;

-- migrate:down
ALTER TABLE practice_tests 
ADD COLUMN studyset_id uuid REFERENCES studysets(id) ON DELETE CASCADE;

UPDATE practice_tests pt
SET studyset_id = subquery.studyset_id
FROM (
    SELECT DISTINCT ON (qt.practice_test_id) 
        qt.practice_test_id, 
        t.studyset_id
    FROM practice_test_question_terms qt
    JOIN terms t ON qt.term_id = t.id
    ORDER BY qt.practice_test_id, t.studyset_id ASC
) subquery
WHERE pt.id = subquery.practice_test_id;

DELETE FROM practice_tests 
WHERE studyset_id IS NULL;

ALTER TABLE practice_tests 
ALTER COLUMN studyset_id SET NOT NULL;


-- migrate:up
CREATE TABLE practice_test_studysets (
    practice_test_id uuid NOT NULL REFERENCES practice_tests(id) ON DELETE CASCADE,
    studyset_id uuid NOT NULL REFERENCES studysets(id) ON DELETE CASCADE,
    PRIMARY KEY (practice_test_id, studyset_id)
);

CREATE INDEX idx_pts_studyset_id ON practice_test_studysets(studyset_id);

GRANT SELECT, INSERT, UPDATE, DELETE ON practice_test_studysets TO quizfreely_api;

-- Populate existing data
INSERT INTO practice_test_studysets (practice_test_id, studyset_id)
SELECT DISTINCT mapping.practice_test_id, t.studyset_id
FROM (
    SELECT practice_test_id, term_id FROM practice_test_question_terms
    UNION
    SELECT practice_test_id, term_id FROM practice_test_distractor_terms
) mapping
JOIN terms t ON mapping.term_id = t.id;

-- migrate:down
DROP TABLE practice_test_studysets;

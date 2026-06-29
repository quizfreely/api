package model

import "time"

type ReviewEvent struct {
	ID                        string                  `db:"id"`
	UserID                    string                  `db:"user_id"`
	TermID                    string                  `db:"term_id"`
	PracticeTestQuestionID    *string                 `db:"practice_test_question_id"`
	Correct                   bool                    `db:"correct"`
	AnswerWith                AnswerWith              `db:"answer_with"`
	Timestamp                 time.Time               `db:"timestamp"`
	AnsweredTermID            *string                 `db:"answered_term_id"`
	PracticeTestQuestionType  *string                 `db:"practice_test_question_type"`
	ReviewActivityType        ReviewActivityTypeEnum  `db:"review_activity_type"`
	AnsweredString            *string                 `db:"answered_string"`
}

type ReviewActivityTypeEnum string

const (
	ReviewActivityTypePracticeTest ReviewActivityTypeEnum = "PRACTICE_TEST"
	ReviewActivityTypeMatch        ReviewActivityTypeEnum = "MATCH"
)

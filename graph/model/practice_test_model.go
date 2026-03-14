// NOTE: filename can't be `practice_test.go` because that ends with `_test.go`

package model

type PracticeTest struct {
	ID               *string     `json:"id,omitempty" db:"id"`
	StudysetID       *string     `json:"studysetId,omitempty" db:"studyset_id"`
	Timestamp        *string     `json:"timestamp,omitempty" db:"timestamp"`
	QuestionsCorrect *int32      `json:"questionsCorrect,omitempty" db:"questions_correct"`
	QuestionsTotal   *int32      `json:"questionsTotal,omitempty" db:"questions_total"`
	Questions        []*Question `json:"questions,omitempty" db:"questions"`
}

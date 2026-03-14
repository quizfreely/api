package model

type TermConfusionPair struct {
	ID             *string     `json:"id,omitempty" db:"id"`
	TermID         *string     `json:"termId,omitempty" db:"term_id"`
	Term           *Term       `json:"term,omitempty"`
	ConfusedTermID *string     `json:"confusedTermId,omitempty" db:"confused_term_id"`
	ConfusedTerm   *Term       `json:"confusedTerm,omitempty"`
	AnsweredWith   *AnswerWith `json:"answeredWith,omitempty" db:"answered_with"`
	ConfusedCount  *int32      `json:"confusedCount,omitempty" db:"confused_count"`
	LastConfusedAt *string     `json:"lastConfusedAt,omitempty" db:"last_confused_at"`
}

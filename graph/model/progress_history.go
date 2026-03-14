package model

type TermProgressHistory struct {
	ID                 *string `json:"id,omitempty" db:"id"`
	TermID             *string `json:"termId,omitempty" db:"term_id"`
	Timestamp          *string `json:"timestamp,omitempty" db:"timestamp"`
	TermCorrectCount   *int32  `json:"termCorrectCount,omitempty" db:"term_correct_count"`
	TermIncorrectCount *int32  `json:"termIncorrectCount,omitempty" db:"term_incorrect_count"`
	DefCorrectCount    *int32  `json:"defCorrectCount,omitempty" db:"def_correct_count"`
	DefIncorrectCount  *int32  `json:"defIncorrectCount,omitempty" db:"def_incorrect_count"`
}

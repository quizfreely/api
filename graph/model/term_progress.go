package model

type TermProgress struct {
	ID                   *string `json:"id,omitempty" db:"id"`
	TermFirstReviewedAt  *string `json:"termFirstReviewedAt,omitempty" db:"term_first_reviewed_at"`
	TermLastReviewedAt   *string `json:"termLastReviewedAt,omitempty" db:"term_last_reviewed_at"`
	TermReviewCount      *int32  `json:"termReviewCount,omitempty" db:"term_review_count"`
	DefFirstReviewedAt   *string `json:"defFirstReviewedAt,omitempty" db:"def_first_reviewed_at"`
	DefLastReviewedAt    *string `json:"defLastReviewedAt,omitempty" db:"def_last_reviewed_at"`
	DefReviewCount       *int32  `json:"defReviewCount,omitempty" db:"def_review_count"`
	TermCorrectCount     int32   `json:"termCorrectCount" db:"term_correct_count"`
	TermIncorrectCount   int32   `json:"termIncorrectCount" db:"term_incorrect_count"`
	DefCorrectCount      int32   `json:"defCorrectCount" db:"def_correct_count"`
	DefIncorrectCount    int32   `json:"defIncorrectCount" db:"def_incorrect_count"`
	TermLeitnerSystemBox *int32  `json:"termLeitnerSystemBox,omitempty" db:"term_leitner_system_box"`
	DefLeitnerSystemBox  *int32  `json:"defLeitnerSystemBox,omitempty" db:"def_leitner_system_box"`
}

package model

type TermProgress struct {
	ID                   *string `json:"id,omitempty"`
	TermID                   *string `json:"termId,omitempty"`
	Timestamp            *string     `json:"timestamp,omitempty"`
	TermCorrectCount     *int32  `json:"termCorrectCount,omitempty"`
	TermIncorrectCount   *int32  `json:"termIncorrectCount,omitempty"`
	DefCorrectCount      *int32  `json:"defCorrectCount,omitempty"`
	DefIncorrectCount    *int32  `json:"defIncorrectCount,omitempty"`
}

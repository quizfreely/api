package model

type SearchQuery struct {
	Query   *string `json:"query,omitempty" db:"keyword"`
	Subject *string `json:"subject,omitempty" db:"subject_id"`
}

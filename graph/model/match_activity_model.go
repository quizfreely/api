package model

type MatchActivity struct {
	ID               *string    `json:"id,omitempty" db:"id"`
	DurationMs       int32      `json:"durationMs"`
	EndTimestamp     string     `json:"endTimestamp"`
	TermIds          []string   `json:"termIds"`
	IncorrectPairIds [][]string `json:"incorrectPairIds"`
}

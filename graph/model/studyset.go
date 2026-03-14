package model

type Studyset struct {
	ID         *string  `json:"id,omitempty" db:"id"`
	Title      *string  `json:"title,omitempty" db:"title"`
	Draft      *bool    `json:"draft,omitempty" db:"draft"`
	Private    *bool    `json:"private,omitempty" db:"private"`
	SubjectID  *string  `json:"subjectId,omitempty" db:"subject_id"`
	Subject    *Subject `json:"subject,omitempty"`
	CreatedAt  *string  `json:"createdAt,omitempty" db:"created_at"`
	UpdatedAt  *string  `json:"updatedAt,omitempty" db:"updated_at"`
	UserID     *string  `json:"userId,omitempty" db:"user_id"`
	User       *User    `json:"user,omitempty"`
	Terms      []*Term  `json:"terms,omitempty"`
	TermsCount *int32   `json:"termsCount,omitempty"`
	Saved      *bool    `json:"saved,omitempty"`
	Folder     *Folder  `json:"folder,omitempty"`
}

package model

type Subject struct {
	ID            *string             `json:"id,omitempty" db:"id"`
	Name          *string             `json:"name,omitempty" db:"name"`
	Category      *SubjectCategory    `json:"category,omitempty" db:"category"`
	Studysets     *StudysetConnection `json:"studysets"`
	StudysetCount *int32              `json:"studysetCount"`
}

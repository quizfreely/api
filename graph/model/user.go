package model

type User struct {
	ID            *string             `json:"id,omitempty" db:"id"`
	Username      *string             `json:"username,omitempty" db:"username"`
	DisplayName   *string             `json:"displayName,omitempty" db:"display_name"`
	Studysets     *StudysetConnection `json:"studysets"`
	StudysetCount *int32              `json:"studysetCount"`
}

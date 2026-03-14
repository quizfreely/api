package model

type Folder struct {
	ID             *string             `json:"id,omitempty" db:"id"`
	Name           *string             `json:"name,omitempty" db:"name"`
	Studysets      *StudysetConnection `json:"studysets"`
	StudysetDrafts *StudysetConnection `json:"studysetDrafts"`
	StudysetCount  *int32              `json:"studysetCount"`
	User           *User               `json:"user,omitempty"`
}

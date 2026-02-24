package model

type Term struct {
	ID           *string `json:"id,omitempty"`
	Term         *string `json:"term,omitempty"`
	Def          *string `json:"def,omitempty"`
	StudysetID   *string
	TermImageURL *string       `json:"termImageUrl,omitempty"`
	DefImageURL  *string       `json:"defImageUrl,omitempty"`
	SortOrder    *int32        `json:"sortOrder,omitempty"`
	Progress     *TermProgress `json:"progress,omitempty"`
	CreatedAt    *string       `json:"createdAt,omitempty"`
	UpdatedAt    *string       `json:"updatedAt,omitempty"`
}

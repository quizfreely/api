package model

type Term struct {
	ID           *string       `json:"id,omitempty" db:"id"`
	Term         *string       `json:"term,omitempty" db:"term"`
	Def          *string       `json:"def,omitempty" db:"def"`
	TermImageURL *string       `json:"termImageUrl,omitempty" db:"term_image_url"`
	DefImageURL  *string       `json:"defImageUrl,omitempty" db:"def_image_url"`
	StudysetID   *string       `db:"studyset_id"`
	SortOrder    *int32        `json:"sortOrder,omitempty" db:"sort_order"`
	Progress     *TermProgress `json:"progress,omitempty"`
	CreatedAt    *string       `json:"createdAt,omitempty" db:"created_at"`
	UpdatedAt    *string       `json:"updatedAt,omitempty" db:"updated_at"`
}

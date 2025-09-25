package models

// CategoryRequest is used for creating/updating categories
type CategoryRequest struct {
	Name     string `json:"name" validate:"required"`
	ParentID *int64 `json:"parent_id"` // nullable
}

// CategoryUpdateRequest allows partial updates
type CategoryUpdateRequest struct {
	Name     string `json:"name"`
	ParentID *int64 `json:"parent_id"`
}

// CategoryResponse is returned to the client
type CategoryResponse struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	ParentID *int64 `json:"parent_id,omitempty"`
}

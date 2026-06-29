package pages

// RolPagesRequest represents the request payload for RolPages
type RolPagesRequest struct {
	Name   string `json:"name" validate:"required,max=20"`
	Icon   string `json:"icon" validate:"max=100"`
	URL    string `json:"url" validate:"required"`
	Level  int16  `json:"level" validate:"required,min=0"`
	Sort   int16  `json:"sort" validate:"required,min=0"`
	Parent *int64 `json:"parent"`
	Active bool   `json:"active"`
}

// RolPagesResponse represents the response payload for RolPages
type RolPagesResponse struct {
	Id       int64               `json:"id"`
	Name     string              `json:"name"`
	Icon     string              `json:"icon"`
	URL      string              `json:"url"`
	Level    int16               `json:"level"`
	Sort     int16               `json:"sort"`
	Parent   *int64              `json:"parent,omitempty"`
	Active   bool                `json:"active"`
	Children []*RolPagesResponse `json:"children,omitempty"`
}

// RolPagesTreeResponse represents hierarchical menu structure
type RolPagesTreeResponse struct {
	Id       int64                   `json:"id"`
	Name     string                  `json:"name"`
	Icon     string                  `json:"icon"`
	URL      string                  `json:"url"`
	Level    int16                   `json:"level"`
	Sort     int16                   `json:"sort"`
	Active   bool                    `json:"active"`
	Children []*RolPagesTreeResponse `json:"children,omitempty"`
}

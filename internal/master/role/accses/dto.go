package role

// RolPagesRequest represents the request payload for RolPages
type RolPagesRequest struct {
	Name   string `json:"name"`
	Icon   string `json:"icon"`
	URL    string `json:"url"`
	Level  int16  `json:"level" validate:"required"`
	Sort   int16  `json:"sort"`
	Parent *int64 `json:"parent"`
	Active bool   `json:"active"`
}

type RolPagesResponse struct {
	Id     int64  `json:"id"`
	Name   string `json:"name"`
	Icon   string `json:"icon"`
	URL    string `json:"url"`
	Level  int16  `json:"level"`
	Sort   int16  `json:"sort"`
	Parent *int64 `json:"parent"`
	Active bool   `json:"active"`
}

// RoleAccessResponse represents the response structure for role access with hierarchical menu
type RoleAccessResponse struct {
	Roles  []string   `json:"roles"`
	Access []MenuItem `json:"access"`
}

// MenuItem represents a menu item in the access hierarchy
type MenuItem struct {
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	Icon     string      `json:"icon"`
	URL      string      `json:"url"`
	Level    string      `json:"level"`
	Children []MenuChild `json:"children"`
}

// MenuChild represents a child menu item
type MenuChild struct {
	Name  string `json:"name"`
	Icon  string `json:"icon"`
	URL   string `json:"url"`
	Level string `json:"level"`
}

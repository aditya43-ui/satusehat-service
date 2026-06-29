package permission

// RolPermissionRequest represents the request payload for RolPermission
type RolPermissionRequest struct {
	Create        bool   `json:"create"`
	Read          bool   `json:"read"`
	Update        bool   `json:"update"`
	Disable       bool   `json:"disable"`
	Delete        bool   `json:"delete"`
	Active        bool   `json:"active"`
	FkRolPagesId  int64  `json:"fk_rol_pages_id"`
	RoleKeycloak  string `json:"role_keycloak"`
	GroupKeycloak string `json:"group_keycloak"`
}

// PermissionTreeResponse represents the complete permission tree response
type PermissionTreeResponse struct {
	Role   string                  `json:"role"`
	Group  []string                `json:"group"`
	Access []*PermissionAccessItem `json:"access"`
}

// PermissionAccessItem represents individual page with permission
type PermissionAccessItem struct {
	Id         int64                   `json:"id"`
	Name       string                  `json:"name"`
	Icon       string                  `json:"icon"`
	URL        string                  `json:"url"`
	Level      int16                   `json:"level"`
	Sort       int16                   `json:"sort"`
	Active     bool                    `json:"active"`
	Permission *PermissionDetail       `json:"permission"`
	Children   []*PermissionAccessItem `json:"children,omitempty"`
}

// PermissionResponse represents permission details
type PermissionResponse struct {
	Id      int64 `json:"id"`
	Create  bool  `json:"create"`
	Read    bool  `json:"read"`
	Update  bool  `json:"update"`
	Delete  bool  `json:"delete"`
	Disable bool  `json:"disable"`
}

// RolPermissionResponse represents the response payload for RolPermission
type RolPermissionResponse struct {
	Id             int64   `json:"id"`
	Create         bool    `json:"create"`
	Read           bool    `json:"read"`
	Update         bool    `json:"update"`
	Disable        bool    `json:"disable"`
	Delete         bool    `json:"delete"`
	Active         bool    `json:"active"`
	FkRolPagesId   int64   `json:"fk_rol_pages_id"`
	RoleMasterName *string `json:"role_master"`
	RoleKeycloak   string  `json:"-"`
	GroupKeycloak  string  `json:"group_keycloak"`
}

// PermissionDetail represents permission details for a page
type PermissionDetail struct {
	Create  bool `json:"create"`
	Read    bool `json:"read"`
	Update  bool `json:"update"`
	Delete  bool `json:"delete"`
	Disable bool `json:"disable"`
}

// PageAccess represents page access with permission
type PageAccess struct {
	Id         int64             `json:"id"`
	Name       string            `json:"name"`
	Icon       string            `json:"icon"`
	URL        string            `json:"url"`
	Group      string            `json:"group_keycloak"`
	Level      int16             `json:"level"`
	Sort       int16             `json:"sort"`
	Active     bool              `json:"active"`
	Permission *PermissionDetail `json:"permission,omitempty"`
	Children   []*PageAccess     `json:"children,omitempty"`
}

// RolePermissionTreeResponse represents the complete role permission tree response
type RolePermissionTreeResponse struct {
	Success bool               `json:"success"`
	Data    RolePermissionData `json:"data"`
}

// RolePermissionData contains role, group, and access information
type RolePermissionData struct {
	Role string `json:"role"`
	// Group  []string          `json:"group"`
	Access []*AccessTreeItem `json:"access"`
}

// AccessTreeItem represents a single item in the access tree
type AccessTreeItem struct {
	ID         int64               `json:"id"`
	Name       string              `json:"name"`
	Icon       string              `json:"icon"`
	URL        string              `json:"url"`
	Group      string              `json:"group"`
	Level      int                 `json:"level"`
	Sort       int                 `json:"sort"`
	Active     bool                `json:"active"`
	Children   []*AccessTreeItem   `json:"children,omitempty"`
	Permission *PermissionResponse `json:"permission,omitempty"`
}

// RolePageAccessResponse represents the structured role page access response
type RolePageAccessResponse struct {
	RoleID   string            `json:"role_id"`
	DataRole []*AccessTreeItem `json:"dataRole"`
	DataAll  []*AccessTreeItem `json:"dataAll"`
}

package master

import "time"

// RoleAccessRolMasterRequest represents the request payload for RoleAccessRolMaster
type RoleMasterRequest struct {
	Id        *int64     `json:"id"`
	Name      *string    `json:"name"`
	Active    *bool      `json:"active"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	Slug      *string    `json:"slug"`
}

// RoleMasterResponse represents the response payload for RoleMaster
type RoleMasterResponse struct {
	Id        int64      `json:"id"`
	Name      *string    `json:"name"`
	Active    *bool      `json:"active"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	Slug      *string    `json:"slug"`
}

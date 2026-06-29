package permission

import (
	"time"
)

// RolPermission entity represents the rol_permission table in the database
type RolPermission struct {
	Id             int64      `json:"id" db:"id" gorm:"primaryKey;autoIncrement"`
	Create         bool       `json:"create" db:"create"`
	Read           bool       `json:"read" db:"read"`
	Update         bool       `json:"update" db:"update"`
	Disable        bool       `json:"disable" db:"disable"`
	Delete         bool       `json:"delete" db:"delete"`
	Active         bool       `json:"active" db:"active"`
	FkRolPagesId   int64      `json:"fk_rol_pages_id" db:"fk_rol_pages_id"`
	RoleMasterName *string    `json:"role_master" db:"role_master_name" gorm:"-"`
	RoleKeycloak   string     `json:"-" db:"role_keycloak"`
	GroupKeycloak  string     `json:"group_keycloak" db:"group_keycloak"`
	CreatedAt      *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at" db:"updated_at"`
}

// TableName specifies the table name for RolPermission
func (RolPermission) TableName() string {
	return "role_access.rol_permission"
}

// RolPermissionWithPage struct untuk hasil JOIN dengan pages
type RolPermissionWithPage struct {
	// Fields dari RolPermission
	Id             int64      `json:"id" db:"id"`
	Create         bool       `json:"create" db:"create"`
	Read           bool       `json:"read" db:"read"`
	Update         bool       `json:"update" db:"update"`
	Disable        bool       `json:"disable" db:"disable"`
	Delete         bool       `json:"delete" db:"delete"`
	Active         bool       `json:"active" db:"active"`
	FkRolPagesId   int64      `json:"fk_rol_pages_id" db:"fk_rol_pages_id"`
	RoleMasterName *string    `json:"role_master" db:"role_master_name" db:"-"`
	RoleKeycloak   string     `json:"-" db:"role_keycloak"`
	GroupKeycloak  string     `json:"group_keycloak" db:"group_keycloak"`
	CreatedAt      *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at" db:"updated_at"`

	// Fields dari RolPages (hasil JOIN)
	PageName     string `json:"page_name" db:"page_name"`
	PageUrl      string `json:"page_url" db:"page_url"`
	PageGroup    string `json:"page_group" db:"page_group"`
	PageLevel    int    `json:"page_level" db:"page_level"`
	PageSort     int    `json:"page_sort" db:"page_sort"`
	PageActive   bool   `json:"page_active" db:"page_active"`
	PageIcon     string `json:"page_icon" db:"page_icon"`
	PageParentId *int64 `json:"page_parent" db:"page_parent"`
}

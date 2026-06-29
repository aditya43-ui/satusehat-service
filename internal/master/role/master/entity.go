package master

import (
	"time"
)

// RoleMaster entity represents the role_access.rol_master table in the database
type RoleMaster struct {
	Id        int64      `json:"id" db:"id" gorm:"primaryKey;autoIncrement"`
	Name      *string    `json:"name" db:"name"`
	Active    *bool      `json:"active" db:"active"`
	CreatedAt *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
	Slug      *string    `json:"slug" db:"slug"`
}

// TableName specifies the table name for RoleMaster
func (RoleMaster) TableName() string {
	return "role_access.rol_master"
}

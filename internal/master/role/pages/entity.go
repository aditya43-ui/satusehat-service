package pages

import (
	"time"
)

// RolPages entity represents the rol_pages table in the database
type RolPages struct {
	Id        int64      `json:"id" db:"id" gorm:"primaryKey;autoIncrement"`
	Name      string     `json:"name" db:"name"`
	Icon      string     `json:"icon" db:"icon"`
	URL       string     `json:"url" db:"url"`
	Level     int16      `json:"level" db:"level"`
	Sort      int16      `json:"sort" db:"sort"`
	Parent    *int64     `json:"parent" db:"parent"`
	Active    bool       `json:"active" db:"active"`
	CreatedAt *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`

	// Relasi untuk hirarki menu
	Children  []RolPages `json:"children" gorm:"-"`
	ParentRef *RolPages  `json:"parent_ref,omitempty" gorm:"-" db:"-"`
}

// TableName specifies the table name for RolPages
func (RolPages) TableName() string {
	return "role_access.rol_pages"
}

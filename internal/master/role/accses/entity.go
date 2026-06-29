package role

// RolPages entity represents the rol_pages table in the database
type RolPages struct {
	Id     int64  `json:"id" db:"id" gorm:"primaryKey;autoIncrement"`
	Name   string `json:"name" db:"name"`
	Icon   string `json:"icon" db:"icon"`
	URL    string `json:"url" db:"url"`
	Level  int16  `json:"level" db:"level"`
	Sort   int16  `json:"sort" db:"sort"`
	Parent *int64 `json:"parent" db:"parent"`
	Active bool   `json:"active" db:"active"`
}

// TableName specifies the table name for RolPages
func (RolPages) TableName() string {
	return "role_access.rol_pages"
}

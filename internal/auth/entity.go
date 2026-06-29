package auth

import "time"

type User struct {
	Id        int64      `json:"id" db:"id" gorm:"primaryKey"`
	Email     string     `json:"email" db:"email"`
	Password  string     `json:"-" db:"password"` // Hidden dari response JSON
	RoleID    *int64     `json:"role_id" db:"role_id"`
	Active    bool       `json:"active" db:"active"`
	CreatedAt *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
}

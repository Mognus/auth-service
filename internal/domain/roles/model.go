package roles

import "time"

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
	RoleGuest UserRole = "guest"
)

type Role struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:50;uniqueIndex;not null" json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Role) TableName() string {
	return "roles"
}

func DefaultRoles() []Role {
	return []Role{
		{Name: string(RoleAdmin)},
		{Name: string(RoleUser)},
		{Name: string(RoleGuest)},
	}
}

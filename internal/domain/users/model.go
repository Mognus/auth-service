package users

import (
	"time"

	"auth-service/internal/domain/roles"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Email     string     `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Password  string     `gorm:"size:255;not null" json:"-"`
	FirstName string     `gorm:"size:100" json:"firstName"`
	LastName  string     `gorm:"size:100" json:"lastName"`
	RoleID    uint       `gorm:"not null;default:2;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"roleId"`
	Role      roles.Role `gorm:"foreignKey:RoleID" json:"role"`
	// Pointer keeps explicit false values from being dropped when GORM sees the DB default.
	Active    *bool     `gorm:"not null;default:true" json:"active"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) BeforeSave(tx *gorm.DB) error {
	if u.Password != "" && u.Password[0] != '$' {
		hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.Password = string(hashed)
	}
	return nil
}

func (u *User) CheckPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) == nil
}

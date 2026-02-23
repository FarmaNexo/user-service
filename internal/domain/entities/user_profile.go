// internal/domain/entities/user_profile.go
package entities

import (
	"time"

	"github.com/google/uuid"
)

// UserProfile representa el perfil de un usuario
type UserProfile struct {
	UserID      uuid.UUID  `gorm:"type:uuid;primaryKey" json:"user_id"`
	FullName    string     `gorm:"type:varchar(255);not null" json:"full_name"`
	Phone       string     `gorm:"type:varchar(20)" json:"phone,omitempty"`
	Bio         string     `gorm:"type:text" json:"bio,omitempty"`
	AvatarURL   string     `gorm:"type:varchar(500)" json:"avatar_url,omitempty"`
	DateOfBirth *time.Time `gorm:"type:date" json:"date_of_birth,omitempty"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName retorna el nombre de la tabla con schema
func (UserProfile) TableName() string {
	return "\"user\".profiles"
}

// NewUserProfile crea un perfil de usuario con valores iniciales
func NewUserProfile(userID uuid.UUID, fullName string) *UserProfile {
	now := time.Now()
	return &UserProfile{
		UserID:    userID,
		FullName:  fullName,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// HasAvatar verifica si el usuario tiene avatar
func (p *UserProfile) HasAvatar() bool {
	return p.AvatarURL != ""
}

// ClearAvatar elimina la URL del avatar
func (p *UserProfile) ClearAvatar() {
	p.AvatarURL = ""
	p.UpdatedAt = time.Now()
}

// SetAvatar establece la URL del avatar
func (p *UserProfile) SetAvatar(url string) {
	p.AvatarURL = url
	p.UpdatedAt = time.Now()
}

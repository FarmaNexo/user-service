// internal/domain/entities/preferences.go
package entities

import (
	"time"

	"github.com/google/uuid"
)

// Preferences representa las preferencias del usuario
type Preferences struct {
	UserID                uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	Language              string    `gorm:"type:varchar(10);not null;default:'es'" json:"language"`
	Theme                 string    `gorm:"type:varchar(20);not null;default:'light'" json:"theme"`
	NotificationsEnabled  bool      `gorm:"not null;default:true" json:"notifications_enabled"`
	EmailNotifications    bool      `gorm:"not null;default:true" json:"email_notifications"`
	SMSNotifications      bool      `gorm:"not null;default:false" json:"sms_notifications"`
	CreatedAt             time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt             time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName retorna el nombre de la tabla con schema
func (Preferences) TableName() string {
	return "\"user\".preferences"
}

// NewDefaultPreferences crea preferencias con valores por defecto
func NewDefaultPreferences(userID uuid.UUID) *Preferences {
	now := time.Now()
	return &Preferences{
		UserID:               userID,
		Language:             "es",
		Theme:                "light",
		NotificationsEnabled: true,
		EmailNotifications:   true,
		SMSNotifications:     false,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

// IsValidLanguage verifica si el idioma es válido
func IsValidLanguage(lang string) bool {
	return lang == "es" || lang == "en"
}

// IsValidTheme verifica si el tema es válido
func IsValidTheme(theme string) bool {
	return theme == "light" || theme == "dark"
}

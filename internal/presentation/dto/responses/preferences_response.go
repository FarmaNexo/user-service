// internal/presentation/dto/responses/preferences_response.go
package responses

import (
	"github.com/farmanexo/user-service/internal/domain/entities"
	"github.com/google/uuid"
)

// PreferencesResponse DTO de respuesta de preferencias
type PreferencesResponse struct {
	UserID               uuid.UUID `json:"user_id"`
	Language             string    `json:"language"`
	Theme                string    `json:"theme"`
	NotificationsEnabled bool      `json:"notifications_enabled"`
	EmailNotifications   bool      `json:"email_notifications"`
	SMSNotifications     bool      `json:"sms_notifications"`
}

// NewPreferencesResponse crea un PreferencesResponse desde una entidad Preferences
func NewPreferencesResponse(prefs *entities.Preferences) PreferencesResponse {
	return PreferencesResponse{
		UserID:               prefs.UserID,
		Language:             prefs.Language,
		Theme:                prefs.Theme,
		NotificationsEnabled: prefs.NotificationsEnabled,
		EmailNotifications:   prefs.EmailNotifications,
		SMSNotifications:     prefs.SMSNotifications,
	}
}

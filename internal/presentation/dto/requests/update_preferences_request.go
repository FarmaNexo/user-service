// internal/presentation/dto/requests/update_preferences_request.go
package requests

// UpdatePreferencesRequest DTO para actualizar preferencias
type UpdatePreferencesRequest struct {
	Language             string `json:"language"`
	Theme                string `json:"theme"`
	NotificationsEnabled bool   `json:"notifications_enabled"`
	EmailNotifications   bool   `json:"email_notifications"`
	SMSNotifications     bool   `json:"sms_notifications"`
}

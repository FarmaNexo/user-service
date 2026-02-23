// internal/application/commands/update_preferences_command.go
package commands

// UpdatePreferencesCommand representa el comando para actualizar preferencias
type UpdatePreferencesCommand struct {
	UserID               string
	Language             string
	Theme                string
	NotificationsEnabled bool
	EmailNotifications   bool
	SMSNotifications     bool
}

func (c UpdatePreferencesCommand) GetName() string {
	return "UpdatePreferencesCommand"
}

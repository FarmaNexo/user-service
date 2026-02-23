// internal/application/commands/update_profile_command.go
package commands

// UpdateProfileCommand representa el comando para actualizar el perfil
type UpdateProfileCommand struct {
	UserID      string
	FullName    string
	Phone       string
	Bio         string
	DateOfBirth string // formato: "2006-01-02"
}

func (c UpdateProfileCommand) GetName() string {
	return "UpdateProfileCommand"
}

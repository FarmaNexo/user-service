// internal/application/commands/delete_avatar_command.go
package commands

// DeleteAvatarCommand representa el comando para eliminar un avatar
type DeleteAvatarCommand struct {
	UserID string
}

func (c DeleteAvatarCommand) GetName() string {
	return "DeleteAvatarCommand"
}

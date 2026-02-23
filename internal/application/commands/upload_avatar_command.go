// internal/application/commands/upload_avatar_command.go
package commands

import "io"

// UploadAvatarCommand representa el comando para subir un avatar
type UploadAvatarCommand struct {
	UserID      string
	File        io.Reader
	Filename    string
	ContentType string
	FileSize    int64
}

func (c UploadAvatarCommand) GetName() string {
	return "UploadAvatarCommand"
}

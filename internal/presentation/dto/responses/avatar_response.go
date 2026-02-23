// internal/presentation/dto/responses/avatar_response.go
package responses

// AvatarResponse DTO de respuesta al subir avatar
type AvatarResponse struct {
	AvatarURL string `json:"avatar_url"`
	Message   string `json:"message"`
}

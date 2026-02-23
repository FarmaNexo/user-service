// internal/presentation/dto/requests/update_profile_request.go
package requests

// UpdateProfileRequest DTO para actualizar perfil
type UpdateProfileRequest struct {
	FullName    string `json:"full_name"`
	Phone       string `json:"phone,omitempty"`
	Bio         string `json:"bio,omitempty"`
	DateOfBirth string `json:"date_of_birth,omitempty"` // formato: "2006-01-02"
}

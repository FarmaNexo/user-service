// internal/presentation/dto/responses/profile_response.go
package responses

import (
	"time"

	"github.com/farmanexo/user-service/internal/domain/entities"
	"github.com/google/uuid"
)

// ProfileResponse DTO de respuesta de perfil.
// Email y Role son una proyección sincronizada desde auth-service vía eventos.
type ProfileResponse struct {
	UserID      uuid.UUID  `json:"user_id"`
	Email       string     `json:"email,omitempty"`
	Role        string     `json:"role"`
	FullName    string     `json:"full_name"`
	Phone       string     `json:"phone,omitempty"`
	Bio         string     `json:"bio,omitempty"`
	AvatarURL   string     `json:"avatar_url,omitempty"`
	DateOfBirth *time.Time `json:"date_of_birth,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// NewProfileResponse crea un ProfileResponse desde una entidad UserProfile
func NewProfileResponse(profile *entities.UserProfile) ProfileResponse {
	return ProfileResponse{
		UserID:      profile.UserID,
		Email:       profile.Email,
		Role:        profile.Role,
		FullName:    profile.FullName,
		Phone:       profile.Phone,
		Bio:         profile.Bio,
		AvatarURL:   profile.AvatarURL,
		DateOfBirth: profile.DateOfBirth,
		CreatedAt:   profile.CreatedAt,
		UpdatedAt:   profile.UpdatedAt,
	}
}

// EmptyResponse respuesta vacía para operaciones sin datos de retorno
type EmptyResponse struct{}

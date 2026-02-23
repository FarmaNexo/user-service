// internal/domain/repositories/profile_repository.go
package repositories

import (
	"context"

	"github.com/farmanexo/user-service/internal/domain/entities"
	"github.com/google/uuid"
)

// ProfileRepository define la interfaz para el repositorio de perfiles
type ProfileRepository interface {
	Create(ctx context.Context, profile *entities.UserProfile) error
	FindByUserID(ctx context.Context, userID uuid.UUID) (*entities.UserProfile, error)
	Update(ctx context.Context, profile *entities.UserProfile) error
	Exists(ctx context.Context, userID uuid.UUID) (bool, error)
}

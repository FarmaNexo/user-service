// internal/domain/repositories/preferences_repository.go
package repositories

import (
	"context"

	"github.com/farmanexo/user-service/internal/domain/entities"
	"github.com/google/uuid"
)

// PreferencesRepository define la interfaz para el repositorio de preferencias
type PreferencesRepository interface {
	Create(ctx context.Context, preferences *entities.Preferences) error
	FindByUserID(ctx context.Context, userID uuid.UUID) (*entities.Preferences, error)
	Update(ctx context.Context, preferences *entities.Preferences) error
}

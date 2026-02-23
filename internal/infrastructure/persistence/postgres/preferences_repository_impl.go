// internal/infrastructure/persistence/postgres/preferences_repository_impl.go
package postgres

import (
	"context"
	"fmt"

	"github.com/farmanexo/user-service/internal/domain/entities"
	"github.com/farmanexo/user-service/internal/domain/repositories"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PreferencesRepositoryImpl implementa PreferencesRepository usando PostgreSQL/GORM
type PreferencesRepositoryImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewPreferencesRepository crea una nueva instancia del repositorio
func NewPreferencesRepository(db *gorm.DB, logger *zap.Logger) *PreferencesRepositoryImpl {
	return &PreferencesRepositoryImpl{
		db:     db,
		logger: logger,
	}
}

func (r *PreferencesRepositoryImpl) Create(ctx context.Context, preferences *entities.Preferences) error {
	result := r.db.WithContext(ctx).Create(preferences)
	if result.Error != nil {
		r.logger.Error("Error creando preferencias",
			zap.String("user_id", preferences.UserID.String()),
			zap.Error(result.Error),
		)
		return fmt.Errorf("error creating preferences: %w", result.Error)
	}
	return nil
}

func (r *PreferencesRepositoryImpl) FindByUserID(ctx context.Context, userID uuid.UUID) (*entities.Preferences, error) {
	var preferences entities.Preferences
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&preferences)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.logger.Error("Error buscando preferencias",
			zap.String("user_id", userID.String()),
			zap.Error(result.Error),
		)
		return nil, fmt.Errorf("error finding preferences: %w", result.Error)
	}
	return &preferences, nil
}

func (r *PreferencesRepositoryImpl) Update(ctx context.Context, preferences *entities.Preferences) error {
	result := r.db.WithContext(ctx).Save(preferences)
	if result.Error != nil {
		r.logger.Error("Error actualizando preferencias",
			zap.String("user_id", preferences.UserID.String()),
			zap.Error(result.Error),
		)
		return fmt.Errorf("error updating preferences: %w", result.Error)
	}
	return nil
}

// Compile-time interface check
var _ repositories.PreferencesRepository = (*PreferencesRepositoryImpl)(nil)

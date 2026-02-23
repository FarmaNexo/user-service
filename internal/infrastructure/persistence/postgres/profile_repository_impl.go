// internal/infrastructure/persistence/postgres/profile_repository_impl.go
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

// ProfileRepositoryImpl implementa ProfileRepository usando PostgreSQL/GORM
type ProfileRepositoryImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewProfileRepository crea una nueva instancia del repositorio
func NewProfileRepository(db *gorm.DB, logger *zap.Logger) *ProfileRepositoryImpl {
	return &ProfileRepositoryImpl{
		db:     db,
		logger: logger,
	}
}

func (r *ProfileRepositoryImpl) Create(ctx context.Context, profile *entities.UserProfile) error {
	result := r.db.WithContext(ctx).Create(profile)
	if result.Error != nil {
		r.logger.Error("Error creando perfil de usuario",
			zap.String("user_id", profile.UserID.String()),
			zap.Error(result.Error),
		)
		return fmt.Errorf("error creating profile: %w", result.Error)
	}

	r.logger.Debug("Perfil creado exitosamente",
		zap.String("user_id", profile.UserID.String()),
	)
	return nil
}

func (r *ProfileRepositoryImpl) FindByUserID(ctx context.Context, userID uuid.UUID) (*entities.UserProfile, error) {
	var profile entities.UserProfile
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.logger.Error("Error buscando perfil",
			zap.String("user_id", userID.String()),
			zap.Error(result.Error),
		)
		return nil, fmt.Errorf("error finding profile: %w", result.Error)
	}
	return &profile, nil
}

func (r *ProfileRepositoryImpl) Update(ctx context.Context, profile *entities.UserProfile) error {
	result := r.db.WithContext(ctx).Save(profile)
	if result.Error != nil {
		r.logger.Error("Error actualizando perfil",
			zap.String("user_id", profile.UserID.String()),
			zap.Error(result.Error),
		)
		return fmt.Errorf("error updating profile: %w", result.Error)
	}
	return nil
}

func (r *ProfileRepositoryImpl) Exists(ctx context.Context, userID uuid.UUID) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&entities.UserProfile{}).Where("user_id = ?", userID).Count(&count)
	if result.Error != nil {
		r.logger.Error("Error verificando existencia de perfil",
			zap.String("user_id", userID.String()),
			zap.Error(result.Error),
		)
		return false, fmt.Errorf("error checking profile existence: %w", result.Error)
	}
	return count > 0, nil
}

// Compile-time interface check
var _ repositories.ProfileRepository = (*ProfileRepositoryImpl)(nil)

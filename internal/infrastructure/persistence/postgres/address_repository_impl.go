// internal/infrastructure/persistence/postgres/address_repository_impl.go
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

// AddressRepositoryImpl implementa AddressRepository usando PostgreSQL/GORM
type AddressRepositoryImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewAddressRepository crea una nueva instancia del repositorio
func NewAddressRepository(db *gorm.DB, logger *zap.Logger) *AddressRepositoryImpl {
	return &AddressRepositoryImpl{
		db:     db,
		logger: logger,
	}
}

func (r *AddressRepositoryImpl) Create(ctx context.Context, address *entities.Address) error {
	result := r.db.WithContext(ctx).Create(address)
	if result.Error != nil {
		r.logger.Error("Error creando dirección",
			zap.String("user_id", address.UserID.String()),
			zap.Error(result.Error),
		)
		return fmt.Errorf("error creating address: %w", result.Error)
	}
	return nil
}

func (r *AddressRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entities.Address, error) {
	var address entities.Address
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&address)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.logger.Error("Error buscando dirección",
			zap.String("id", id.String()),
			zap.Error(result.Error),
		)
		return nil, fmt.Errorf("error finding address: %w", result.Error)
	}
	return &address, nil
}

func (r *AddressRepositoryImpl) FindByUserID(ctx context.Context, userID uuid.UUID) ([]entities.Address, error) {
	var addresses []entities.Address
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("is_default DESC, created_at ASC").Find(&addresses)
	if result.Error != nil {
		r.logger.Error("Error listando direcciones",
			zap.String("user_id", userID.String()),
			zap.Error(result.Error),
		)
		return nil, fmt.Errorf("error finding addresses: %w", result.Error)
	}
	return addresses, nil
}

func (r *AddressRepositoryImpl) Update(ctx context.Context, address *entities.Address) error {
	result := r.db.WithContext(ctx).Save(address)
	if result.Error != nil {
		r.logger.Error("Error actualizando dirección",
			zap.String("id", address.ID.String()),
			zap.Error(result.Error),
		)
		return fmt.Errorf("error updating address: %w", result.Error)
	}
	return nil
}

func (r *AddressRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entities.Address{})
	if result.Error != nil {
		r.logger.Error("Error eliminando dirección",
			zap.String("id", id.String()),
			zap.Error(result.Error),
		)
		return fmt.Errorf("error deleting address: %w", result.Error)
	}
	return nil
}

func (r *AddressRepositoryImpl) ClearDefaultByUserID(ctx context.Context, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&entities.Address{}).
		Where("user_id = ? AND is_default = true", userID).
		Update("is_default", false)
	if result.Error != nil {
		r.logger.Error("Error limpiando dirección por defecto",
			zap.String("user_id", userID.String()),
			zap.Error(result.Error),
		)
		return fmt.Errorf("error clearing default address: %w", result.Error)
	}
	return nil
}

func (r *AddressRepositoryImpl) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&entities.Address{}).Where("user_id = ?", userID).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("error counting addresses: %w", result.Error)
	}
	return count, nil
}

// Compile-time interface check
var _ repositories.AddressRepository = (*AddressRepositoryImpl)(nil)

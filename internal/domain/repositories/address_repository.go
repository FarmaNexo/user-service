// internal/domain/repositories/address_repository.go
package repositories

import (
	"context"

	"github.com/farmanexo/user-service/internal/domain/entities"
	"github.com/google/uuid"
)

// AddressRepository define la interfaz para el repositorio de direcciones
type AddressRepository interface {
	Create(ctx context.Context, address *entities.Address) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Address, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]entities.Address, error)
	Update(ctx context.Context, address *entities.Address) error
	Delete(ctx context.Context, id uuid.UUID) error
	ClearDefaultByUserID(ctx context.Context, userID uuid.UUID) error
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
}

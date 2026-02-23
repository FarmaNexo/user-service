// internal/application/handlers/update_address_handler.go
package handlers

import (
	"context"
	"time"

	"github.com/farmanexo/user-service/internal/application/commands"
	"github.com/farmanexo/user-service/internal/domain/repositories"
	"github.com/farmanexo/user-service/internal/presentation/dto/responses"
	"github.com/farmanexo/user-service/internal/shared/common"
	"github.com/farmanexo/user-service/internal/shared/constants"
	"github.com/farmanexo/user-service/pkg/mediator"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// UpdateAddressHandler maneja el comando UpdateAddressCommand
type UpdateAddressHandler struct {
	addressRepo repositories.AddressRepository
	logger      *zap.Logger
}

func NewUpdateAddressHandler(addressRepo repositories.AddressRepository, logger *zap.Logger) *UpdateAddressHandler {
	return &UpdateAddressHandler{
		addressRepo: addressRepo,
		logger:      logger,
	}
}

func (h *UpdateAddressHandler) Handle(
	ctx context.Context,
	command commands.UpdateAddressCommand,
) (*common.ApiResponse[responses.AddressResponse], error) {
	h.logger.Info("Actualizando dirección",
		zap.String("user_id", command.UserID),
		zap.String("address_id", command.AddressID),
	)

	userID, err := uuid.Parse(command.UserID)
	if err != nil {
		return common.BadRequestResponse[responses.AddressResponse](
			constants.CodeValidationError, "ID de usuario inválido",
		), nil
	}

	addressID, err := uuid.Parse(command.AddressID)
	if err != nil {
		return common.BadRequestResponse[responses.AddressResponse](
			constants.CodeValidationError, "ID de dirección inválido",
		), nil
	}

	address, err := h.addressRepo.FindByID(ctx, addressID)
	if err != nil {
		h.logger.Error("Error buscando dirección", zap.Error(err))
		return common.InternalServerErrorResponse[responses.AddressResponse](
			"Error obteniendo dirección",
		), nil
	}

	if address == nil {
		return common.NotFoundResponse[responses.AddressResponse](
			constants.GetDescription(constants.CodeAddressNotFound),
		), nil
	}

	// Verificar que la dirección pertenece al usuario
	if address.UserID != userID {
		return common.ForbiddenResponse[responses.AddressResponse](
			"No tienes permiso para modificar esta dirección",
		), nil
	}

	// Si se marca como default, limpiar las otras
	if command.IsDefault && !address.IsDefault {
		if err := h.addressRepo.ClearDefaultByUserID(ctx, userID); err != nil {
			h.logger.Error("Error limpiando dirección por defecto", zap.Error(err))
			return common.InternalServerErrorResponse[responses.AddressResponse](
				"Error procesando dirección",
			), nil
		}
	}

	// Actualizar campos
	address.Label = command.Label
	address.Street = command.Street
	address.City = command.City
	address.State = command.State
	address.PostalCode = command.PostalCode
	if command.Country != "" {
		address.Country = command.Country
	}
	address.IsDefault = command.IsDefault
	address.Latitude = command.Latitude
	address.Longitude = command.Longitude
	address.UpdatedAt = time.Now()

	if err := h.addressRepo.Update(ctx, address); err != nil {
		h.logger.Error("Error actualizando dirección", zap.Error(err))
		return common.InternalServerErrorResponse[responses.AddressResponse](
			"Error actualizando dirección",
		), nil
	}

	addrResp := responses.NewAddressResponse(address)
	resp := common.OkResponse(addrResp)
	resp.AddMessageWithType(constants.CodeAddressUpdated,
		constants.GetDescription(constants.CodeAddressUpdated),
		constants.MessageTypeSuccess,
	)
	return resp, nil
}

var _ mediator.RequestHandler[commands.UpdateAddressCommand, responses.AddressResponse] = (*UpdateAddressHandler)(nil)

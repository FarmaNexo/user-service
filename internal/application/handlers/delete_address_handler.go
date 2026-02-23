// internal/application/handlers/delete_address_handler.go
package handlers

import (
	"context"

	"github.com/farmanexo/user-service/internal/application/commands"
	"github.com/farmanexo/user-service/internal/domain/repositories"
	"github.com/farmanexo/user-service/internal/presentation/dto/responses"
	"github.com/farmanexo/user-service/internal/shared/common"
	"github.com/farmanexo/user-service/internal/shared/constants"
	"github.com/farmanexo/user-service/pkg/mediator"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DeleteAddressHandler maneja el comando DeleteAddressCommand
type DeleteAddressHandler struct {
	addressRepo repositories.AddressRepository
	logger      *zap.Logger
}

func NewDeleteAddressHandler(addressRepo repositories.AddressRepository, logger *zap.Logger) *DeleteAddressHandler {
	return &DeleteAddressHandler{
		addressRepo: addressRepo,
		logger:      logger,
	}
}

func (h *DeleteAddressHandler) Handle(
	ctx context.Context,
	command commands.DeleteAddressCommand,
) (*common.ApiResponse[responses.EmptyResponse], error) {
	h.logger.Info("Eliminando dirección",
		zap.String("user_id", command.UserID),
		zap.String("address_id", command.AddressID),
	)

	userID, err := uuid.Parse(command.UserID)
	if err != nil {
		return common.BadRequestResponse[responses.EmptyResponse](
			constants.CodeValidationError, "ID de usuario inválido",
		), nil
	}

	addressID, err := uuid.Parse(command.AddressID)
	if err != nil {
		return common.BadRequestResponse[responses.EmptyResponse](
			constants.CodeValidationError, "ID de dirección inválido",
		), nil
	}

	address, err := h.addressRepo.FindByID(ctx, addressID)
	if err != nil {
		h.logger.Error("Error buscando dirección", zap.Error(err))
		return common.InternalServerErrorResponse[responses.EmptyResponse](
			"Error obteniendo dirección",
		), nil
	}

	if address == nil {
		return common.NotFoundResponse[responses.EmptyResponse](
			constants.GetDescription(constants.CodeAddressNotFound),
		), nil
	}

	// Verificar que la dirección pertenece al usuario
	if address.UserID != userID {
		return common.ForbiddenResponse[responses.EmptyResponse](
			"No tienes permiso para eliminar esta dirección",
		), nil
	}

	if err := h.addressRepo.Delete(ctx, addressID); err != nil {
		h.logger.Error("Error eliminando dirección", zap.Error(err))
		return common.InternalServerErrorResponse[responses.EmptyResponse](
			"Error eliminando dirección",
		), nil
	}

	resp := common.OkResponse(responses.EmptyResponse{})
	resp.AddMessageWithType(constants.CodeAddressDeleted,
		constants.GetDescription(constants.CodeAddressDeleted),
		constants.MessageTypeSuccess,
	)
	return resp, nil
}

var _ mediator.RequestHandler[commands.DeleteAddressCommand, responses.EmptyResponse] = (*DeleteAddressHandler)(nil)

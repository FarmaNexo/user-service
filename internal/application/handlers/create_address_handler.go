// internal/application/handlers/create_address_handler.go
package handlers

import (
	"context"

	"github.com/farmanexo/user-service/internal/application/commands"
	"github.com/farmanexo/user-service/internal/domain/entities"
	"github.com/farmanexo/user-service/internal/domain/events"
	"github.com/farmanexo/user-service/internal/domain/repositories"
	"github.com/farmanexo/user-service/internal/domain/services"
	"github.com/farmanexo/user-service/internal/presentation/dto/responses"
	"github.com/farmanexo/user-service/internal/shared/common"
	"github.com/farmanexo/user-service/internal/shared/constants"
	"github.com/farmanexo/user-service/pkg/mediator"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const maxAddressesPerUser = 10

// CreateAddressHandler maneja el comando CreateAddressCommand
type CreateAddressHandler struct {
	addressRepo    repositories.AddressRepository
	eventPublisher services.EventPublisher
	logger         *zap.Logger
}

func NewCreateAddressHandler(
	addressRepo repositories.AddressRepository,
	eventPublisher services.EventPublisher,
	logger *zap.Logger,
) *CreateAddressHandler {
	return &CreateAddressHandler{
		addressRepo:    addressRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

func (h *CreateAddressHandler) Handle(
	ctx context.Context,
	command commands.CreateAddressCommand,
) (*common.ApiResponse[responses.AddressResponse], error) {
	h.logger.Info("Creando dirección", zap.String("user_id", command.UserID))

	userID, err := uuid.Parse(command.UserID)
	if err != nil {
		return common.BadRequestResponse[responses.AddressResponse](
			constants.CodeValidationError, "ID de usuario inválido",
		), nil
	}

	// Verificar límite de direcciones
	count, err := h.addressRepo.CountByUserID(ctx, userID)
	if err != nil {
		h.logger.Error("Error contando direcciones", zap.Error(err))
		return common.InternalServerErrorResponse[responses.AddressResponse](
			"Error verificando direcciones",
		), nil
	}

	if count >= maxAddressesPerUser {
		return common.BadRequestResponse[responses.AddressResponse](
			constants.CodeMaxAddresses,
			constants.GetDescription(constants.CodeMaxAddresses),
		), nil
	}

	// Si es default, limpiar las otras
	if command.IsDefault {
		if err := h.addressRepo.ClearDefaultByUserID(ctx, userID); err != nil {
			h.logger.Error("Error limpiando dirección por defecto", zap.Error(err))
			return common.InternalServerErrorResponse[responses.AddressResponse](
				"Error procesando dirección",
			), nil
		}
	}

	address := entities.NewAddress(
		userID,
		command.Label,
		command.Street,
		command.City,
		command.State,
		command.PostalCode,
		command.Country,
		command.IsDefault,
	)
	address.Latitude = command.Latitude
	address.Longitude = command.Longitude

	if err := h.addressRepo.Create(ctx, address); err != nil {
		h.logger.Error("Error creando dirección", zap.Error(err))
		return common.InternalServerErrorResponse[responses.AddressResponse](
			"Error creando dirección",
		), nil
	}

	// Publicar evento (fire-and-forget)
	go func() {
		if err := h.eventPublisher.Publish(context.Background(),
			events.NewAddressCreatedEvent(command.UserID, address.ID.String())); err != nil {
			h.logger.Warn("Error publicando evento dirección", zap.Error(err))
		}
	}()

	addrResp := responses.NewAddressResponse(address)
	return common.CreatedResponse(addrResp), nil
}

var _ mediator.RequestHandler[commands.CreateAddressCommand, responses.AddressResponse] = (*CreateAddressHandler)(nil)

// internal/application/handlers/get_address_by_id_handler.go
package handlers

import (
	"context"

	"github.com/farmanexo/user-service/internal/application/queries"
	"github.com/farmanexo/user-service/internal/domain/repositories"
	"github.com/farmanexo/user-service/internal/presentation/dto/responses"
	"github.com/farmanexo/user-service/internal/shared/common"
	"github.com/farmanexo/user-service/internal/shared/constants"
	"github.com/farmanexo/user-service/pkg/mediator"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GetAddressByIDHandler resuelve una dirección por ID. Aplica ownership:
// si la dirección pertenece a otro usuario devuelve 403, NO 404, para
// distinguir "no existe" de "no es tuya" en logs y debugging.
type GetAddressByIDHandler struct {
	addressRepo repositories.AddressRepository
	logger      *zap.Logger
}

func NewGetAddressByIDHandler(addressRepo repositories.AddressRepository, logger *zap.Logger) *GetAddressByIDHandler {
	return &GetAddressByIDHandler{
		addressRepo: addressRepo,
		logger:      logger,
	}
}

func (h *GetAddressByIDHandler) Handle(
	ctx context.Context,
	query queries.GetAddressByIDQuery,
) (*common.ApiResponse[responses.AddressResponse], error) {
	userID, err := uuid.Parse(query.UserID)
	if err != nil {
		return common.BadRequestResponse[responses.AddressResponse](
			constants.CodeValidationError, "ID de usuario inválido",
		), nil
	}

	addressID, err := uuid.Parse(query.AddressID)
	if err != nil {
		return common.BadRequestResponse[responses.AddressResponse](
			constants.CodeValidationError, "ID de dirección inválido",
		), nil
	}

	address, err := h.addressRepo.FindByID(ctx, addressID)
	if err != nil {
		h.logger.Error("Error buscando dirección",
			zap.String("address_id", query.AddressID),
			zap.Error(err),
		)
		return common.InternalServerErrorResponse[responses.AddressResponse](
			"Error obteniendo dirección",
		), nil
	}

	if address == nil {
		return common.NotFoundResponse[responses.AddressResponse](
			constants.GetDescription(constants.CodeAddressNotFound),
		), nil
	}

	if address.UserID != userID {
		h.logger.Warn("Intento de acceso a dirección de otro usuario",
			zap.String("requested_address_id", query.AddressID),
			zap.String("requesting_user_id", query.UserID),
		)
		return common.ForbiddenResponse[responses.AddressResponse](
			"No tienes permiso para ver esta dirección",
		), nil
	}

	return common.OkResponse(responses.NewAddressResponse(address)), nil
}

var _ mediator.RequestHandler[queries.GetAddressByIDQuery, responses.AddressResponse] = (*GetAddressByIDHandler)(nil)

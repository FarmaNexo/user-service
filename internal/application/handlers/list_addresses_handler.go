// internal/application/handlers/list_addresses_handler.go
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

// ListAddressesHandler maneja la query ListAddressesQuery
type ListAddressesHandler struct {
	addressRepo repositories.AddressRepository
	logger      *zap.Logger
}

func NewListAddressesHandler(addressRepo repositories.AddressRepository, logger *zap.Logger) *ListAddressesHandler {
	return &ListAddressesHandler{
		addressRepo: addressRepo,
		logger:      logger,
	}
}

func (h *ListAddressesHandler) Handle(
	ctx context.Context,
	query queries.ListAddressesQuery,
) (*common.ApiResponse[responses.AddressListResponse], error) {
	h.logger.Info("Listando direcciones", zap.String("user_id", query.UserID))

	userID, err := uuid.Parse(query.UserID)
	if err != nil {
		return common.BadRequestResponse[responses.AddressListResponse](
			constants.CodeValidationError, "ID de usuario inválido",
		), nil
	}

	addresses, err := h.addressRepo.FindByUserID(ctx, userID)
	if err != nil {
		h.logger.Error("Error listando direcciones", zap.Error(err))
		return common.InternalServerErrorResponse[responses.AddressListResponse](
			"Error obteniendo direcciones",
		), nil
	}

	listResp := responses.NewAddressListResponse(addresses)
	return common.OkResponse(listResp), nil
}

var _ mediator.RequestHandler[queries.ListAddressesQuery, responses.AddressListResponse] = (*ListAddressesHandler)(nil)

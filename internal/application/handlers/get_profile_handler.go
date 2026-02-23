// internal/application/handlers/get_profile_handler.go
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

// GetProfileHandler maneja la query GetProfileQuery
type GetProfileHandler struct {
	profileRepo repositories.ProfileRepository
	logger      *zap.Logger
}

func NewGetProfileHandler(profileRepo repositories.ProfileRepository, logger *zap.Logger) *GetProfileHandler {
	return &GetProfileHandler{
		profileRepo: profileRepo,
		logger:      logger,
	}
}

func (h *GetProfileHandler) Handle(
	ctx context.Context,
	query queries.GetProfileQuery,
) (*common.ApiResponse[responses.ProfileResponse], error) {
	h.logger.Info("Obteniendo perfil de usuario", zap.String("user_id", query.UserID))

	userID, err := uuid.Parse(query.UserID)
	if err != nil {
		return common.BadRequestResponse[responses.ProfileResponse](
			constants.CodeValidationError, "ID de usuario inválido",
		), nil
	}

	profile, err := h.profileRepo.FindByUserID(ctx, userID)
	if err != nil {
		h.logger.Error("Error buscando perfil", zap.Error(err))
		return common.InternalServerErrorResponse[responses.ProfileResponse](
			"Error obteniendo perfil",
		), nil
	}

	if profile == nil {
		return common.NotFoundResponse[responses.ProfileResponse](
			constants.GetDescription(constants.CodeProfileNotFound),
		), nil
	}

	profileResp := responses.NewProfileResponse(profile)
	resp := common.OkResponse(profileResp)
	resp.AddMessageWithType(constants.CodeProfileRetrieved,
		constants.GetDescription(constants.CodeProfileRetrieved),
		constants.MessageTypeSuccess,
	)
	return resp, nil
}

var _ mediator.RequestHandler[queries.GetProfileQuery, responses.ProfileResponse] = (*GetProfileHandler)(nil)

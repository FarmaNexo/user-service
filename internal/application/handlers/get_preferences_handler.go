// internal/application/handlers/get_preferences_handler.go
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

// GetPreferencesHandler maneja la query GetPreferencesQuery
type GetPreferencesHandler struct {
	prefsRepo repositories.PreferencesRepository
	logger    *zap.Logger
}

func NewGetPreferencesHandler(prefsRepo repositories.PreferencesRepository, logger *zap.Logger) *GetPreferencesHandler {
	return &GetPreferencesHandler{
		prefsRepo: prefsRepo,
		logger:    logger,
	}
}

func (h *GetPreferencesHandler) Handle(
	ctx context.Context,
	query queries.GetPreferencesQuery,
) (*common.ApiResponse[responses.PreferencesResponse], error) {
	h.logger.Info("Obteniendo preferencias", zap.String("user_id", query.UserID))

	userID, err := uuid.Parse(query.UserID)
	if err != nil {
		return common.BadRequestResponse[responses.PreferencesResponse](
			constants.CodeValidationError, "ID de usuario inválido",
		), nil
	}

	prefs, err := h.prefsRepo.FindByUserID(ctx, userID)
	if err != nil {
		h.logger.Error("Error buscando preferencias", zap.Error(err))
		return common.InternalServerErrorResponse[responses.PreferencesResponse](
			"Error obteniendo preferencias",
		), nil
	}

	if prefs == nil {
		return common.NotFoundResponse[responses.PreferencesResponse](
			"Preferencias no encontradas",
		), nil
	}

	prefsResp := responses.NewPreferencesResponse(prefs)
	resp := common.OkResponse(prefsResp)
	resp.AddMessageWithType(constants.CodePreferencesLoaded,
		constants.GetDescription(constants.CodePreferencesLoaded),
		constants.MessageTypeSuccess,
	)
	return resp, nil
}

var _ mediator.RequestHandler[queries.GetPreferencesQuery, responses.PreferencesResponse] = (*GetPreferencesHandler)(nil)

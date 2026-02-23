// internal/application/handlers/update_preferences_handler.go
package handlers

import (
	"context"
	"time"

	"github.com/farmanexo/user-service/internal/application/commands"
	"github.com/farmanexo/user-service/internal/domain/entities"
	"github.com/farmanexo/user-service/internal/domain/repositories"
	"github.com/farmanexo/user-service/internal/presentation/dto/responses"
	"github.com/farmanexo/user-service/internal/shared/common"
	"github.com/farmanexo/user-service/internal/shared/constants"
	"github.com/farmanexo/user-service/pkg/mediator"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// UpdatePreferencesHandler maneja el comando UpdatePreferencesCommand
type UpdatePreferencesHandler struct {
	prefsRepo repositories.PreferencesRepository
	logger    *zap.Logger
}

func NewUpdatePreferencesHandler(prefsRepo repositories.PreferencesRepository, logger *zap.Logger) *UpdatePreferencesHandler {
	return &UpdatePreferencesHandler{
		prefsRepo: prefsRepo,
		logger:    logger,
	}
}

func (h *UpdatePreferencesHandler) Handle(
	ctx context.Context,
	command commands.UpdatePreferencesCommand,
) (*common.ApiResponse[responses.PreferencesResponse], error) {
	h.logger.Info("Actualizando preferencias", zap.String("user_id", command.UserID))

	// Validar language
	if !entities.IsValidLanguage(command.Language) {
		return common.BadRequestResponse[responses.PreferencesResponse](
			constants.CodeInvalidLanguage,
			constants.GetDescription(constants.CodeInvalidLanguage),
		), nil
	}

	// Validar theme
	if !entities.IsValidTheme(command.Theme) {
		return common.BadRequestResponse[responses.PreferencesResponse](
			constants.CodeInvalidTheme,
			constants.GetDescription(constants.CodeInvalidTheme),
		), nil
	}

	userID, err := uuid.Parse(command.UserID)
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

	// Actualizar campos
	prefs.Language = command.Language
	prefs.Theme = command.Theme
	prefs.NotificationsEnabled = command.NotificationsEnabled
	prefs.EmailNotifications = command.EmailNotifications
	prefs.SMSNotifications = command.SMSNotifications
	prefs.UpdatedAt = time.Now()

	if err := h.prefsRepo.Update(ctx, prefs); err != nil {
		h.logger.Error("Error actualizando preferencias", zap.Error(err))
		return common.InternalServerErrorResponse[responses.PreferencesResponse](
			"Error actualizando preferencias",
		), nil
	}

	prefsResp := responses.NewPreferencesResponse(prefs)
	resp := common.OkResponse(prefsResp)
	resp.AddMessageWithType(constants.CodePreferencesUpdate,
		constants.GetDescription(constants.CodePreferencesUpdate),
		constants.MessageTypeSuccess,
	)
	return resp, nil
}

var _ mediator.RequestHandler[commands.UpdatePreferencesCommand, responses.PreferencesResponse] = (*UpdatePreferencesHandler)(nil)

// internal/application/handlers/update_profile_handler.go
package handlers

import (
	"context"
	"time"

	"github.com/farmanexo/user-service/internal/application/commands"
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

// UpdateProfileHandler maneja el comando UpdateProfileCommand
type UpdateProfileHandler struct {
	profileRepo    repositories.ProfileRepository
	eventPublisher services.EventPublisher
	logger         *zap.Logger
}

func NewUpdateProfileHandler(
	profileRepo repositories.ProfileRepository,
	eventPublisher services.EventPublisher,
	logger *zap.Logger,
) *UpdateProfileHandler {
	return &UpdateProfileHandler{
		profileRepo:    profileRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

func (h *UpdateProfileHandler) Handle(
	ctx context.Context,
	command commands.UpdateProfileCommand,
) (*common.ApiResponse[responses.ProfileResponse], error) {
	h.logger.Info("Actualizando perfil", zap.String("user_id", command.UserID))

	userID, err := uuid.Parse(command.UserID)
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

	// Actualizar campos
	if command.FullName != "" {
		profile.FullName = command.FullName
	}
	profile.Phone = command.Phone
	profile.Bio = command.Bio

	if command.DateOfBirth != "" {
		dob, err := time.Parse("2006-01-02", command.DateOfBirth)
		if err != nil {
			return common.BadRequestResponse[responses.ProfileResponse](
				constants.CodeValidationError, "Formato de fecha inválido. Use: YYYY-MM-DD",
			), nil
		}
		profile.DateOfBirth = &dob
	}

	profile.UpdatedAt = time.Now()

	if err := h.profileRepo.Update(ctx, profile); err != nil {
		h.logger.Error("Error actualizando perfil", zap.Error(err))
		return common.InternalServerErrorResponse[responses.ProfileResponse](
			"Error actualizando perfil",
		), nil
	}

	// Publicar evento (fire-and-forget)
	go h.publishEvent(context.Background(), events.NewProfileUpdatedEvent(command.UserID))

	profileResp := responses.NewProfileResponse(profile)
	resp := common.OkResponse(profileResp)
	resp.AddMessageWithType(constants.CodeProfileUpdated,
		constants.GetDescription(constants.CodeProfileUpdated),
		constants.MessageTypeSuccess,
	)
	return resp, nil
}

func (h *UpdateProfileHandler) publishEvent(ctx context.Context, event events.UserEvent) {
	if err := h.eventPublisher.Publish(ctx, event); err != nil {
		h.logger.Warn("Error publicando evento",
			zap.String("event_type", event.EventType),
			zap.Error(err),
		)
	}
}

var _ mediator.RequestHandler[commands.UpdateProfileCommand, responses.ProfileResponse] = (*UpdateProfileHandler)(nil)

// internal/application/handlers/delete_avatar_handler.go
package handlers

import (
	"context"
	"fmt"
	"strings"

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

// DeleteAvatarHandler maneja el comando DeleteAvatarCommand
type DeleteAvatarHandler struct {
	profileRepo    repositories.ProfileRepository
	fileStorage    services.FileStorage
	eventPublisher services.EventPublisher
	logger         *zap.Logger
}

func NewDeleteAvatarHandler(
	profileRepo repositories.ProfileRepository,
	fileStorage services.FileStorage,
	eventPublisher services.EventPublisher,
	logger *zap.Logger,
) *DeleteAvatarHandler {
	return &DeleteAvatarHandler{
		profileRepo:    profileRepo,
		fileStorage:    fileStorage,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

func (h *DeleteAvatarHandler) Handle(
	ctx context.Context,
	command commands.DeleteAvatarCommand,
) (*common.ApiResponse[responses.EmptyResponse], error) {
	h.logger.Info("Eliminando avatar", zap.String("user_id", command.UserID))

	userID, err := uuid.Parse(command.UserID)
	if err != nil {
		return common.BadRequestResponse[responses.EmptyResponse](
			constants.CodeValidationError, "ID de usuario inválido",
		), nil
	}

	profile, err := h.profileRepo.FindByUserID(ctx, userID)
	if err != nil {
		h.logger.Error("Error buscando perfil", zap.Error(err))
		return common.InternalServerErrorResponse[responses.EmptyResponse](
			"Error obteniendo perfil",
		), nil
	}

	if profile == nil {
		return common.NotFoundResponse[responses.EmptyResponse](
			constants.GetDescription(constants.CodeProfileNotFound),
		), nil
	}

	if !profile.HasAvatar() {
		resp := common.OkResponse(responses.EmptyResponse{})
		resp.AddMessageWithType(constants.CodeSuccess, "El usuario no tiene avatar", constants.MessageTypeInformation)
		return resp, nil
	}

	// Extraer key del avatar URL para eliminar de S3
	key := h.extractS3Key(profile.AvatarURL, command.UserID)
	if key != "" {
		if err := h.fileStorage.Delete(ctx, avatarsBucket, key); err != nil {
			h.logger.Warn("Error eliminando avatar de S3 (continuando)",
				zap.String("key", key),
				zap.Error(err),
			)
		}
	}

	// Limpiar avatar del perfil
	profile.ClearAvatar()
	if err := h.profileRepo.Update(ctx, profile); err != nil {
		h.logger.Error("Error actualizando perfil", zap.Error(err))
		return common.InternalServerErrorResponse[responses.EmptyResponse](
			"Error actualizando perfil",
		), nil
	}

	// Publicar evento (fire-and-forget)
	go func() {
		if err := h.eventPublisher.Publish(context.Background(),
			events.NewAvatarChangedEvent(command.UserID, "deleted")); err != nil {
			h.logger.Warn("Error publicando evento avatar", zap.Error(err))
		}
	}()

	resp := common.OkResponse(responses.EmptyResponse{})
	resp.AddMessageWithType(constants.CodeAvatarDeleted,
		constants.GetDescription(constants.CodeAvatarDeleted),
		constants.MessageTypeSuccess,
	)
	return resp, nil
}

func (h *DeleteAvatarHandler) extractS3Key(avatarURL, userID string) string {
	// Buscar el path "users/{userID}/avatar.xxx" en la URL
	prefix := fmt.Sprintf("users/%s/", userID)
	idx := strings.Index(avatarURL, prefix)
	if idx >= 0 {
		return avatarURL[idx:]
	}
	return ""
}

var _ mediator.RequestHandler[commands.DeleteAvatarCommand, responses.EmptyResponse] = (*DeleteAvatarHandler)(nil)

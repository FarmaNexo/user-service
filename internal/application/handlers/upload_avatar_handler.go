// internal/application/handlers/upload_avatar_handler.go
package handlers

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
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

const (
	maxAvatarSize  = 5 * 1024 * 1024 // 5MB
	avatarWidth    = 512
	avatarHeight   = 512
	avatarsBucket  = "farmanexo-avatars"
)

var allowedContentTypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/webp": true,
}

// UploadAvatarHandler maneja el comando UploadAvatarCommand
type UploadAvatarHandler struct {
	profileRepo    repositories.ProfileRepository
	fileStorage    services.FileStorage
	eventPublisher services.EventPublisher
	logger         *zap.Logger
}

func NewUploadAvatarHandler(
	profileRepo repositories.ProfileRepository,
	fileStorage services.FileStorage,
	eventPublisher services.EventPublisher,
	logger *zap.Logger,
) *UploadAvatarHandler {
	return &UploadAvatarHandler{
		profileRepo:    profileRepo,
		fileStorage:    fileStorage,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

func (h *UploadAvatarHandler) Handle(
	ctx context.Context,
	command commands.UploadAvatarCommand,
) (*common.ApiResponse[responses.AvatarResponse], error) {
	h.logger.Info("Subiendo avatar", zap.String("user_id", command.UserID))

	// Validar tipo de archivo
	if !allowedContentTypes[command.ContentType] {
		return common.BadRequestResponse[responses.AvatarResponse](
			constants.CodeInvalidFileType,
			constants.GetDescription(constants.CodeInvalidFileType),
		), nil
	}

	// Validar tamaño
	if command.FileSize > maxAvatarSize {
		return common.BadRequestResponse[responses.AvatarResponse](
			constants.CodeFileTooLarge,
			constants.GetDescription(constants.CodeFileTooLarge),
		), nil
	}

	userID, err := uuid.Parse(command.UserID)
	if err != nil {
		return common.BadRequestResponse[responses.AvatarResponse](
			constants.CodeValidationError, "ID de usuario inválido",
		), nil
	}

	// Buscar perfil
	profile, err := h.profileRepo.FindByUserID(ctx, userID)
	if err != nil {
		h.logger.Error("Error buscando perfil", zap.Error(err))
		return common.InternalServerErrorResponse[responses.AvatarResponse](
			"Error obteniendo perfil",
		), nil
	}

	if profile == nil {
		return common.NotFoundResponse[responses.AvatarResponse](
			constants.GetDescription(constants.CodeProfileNotFound),
		), nil
	}

	// Redimensionar imagen
	resizedBuf, contentType, err := h.resizeImage(command)
	if err != nil {
		h.logger.Error("Error redimensionando imagen", zap.Error(err))
		return common.InternalServerErrorResponse[responses.AvatarResponse](
			"Error procesando la imagen",
		), nil
	}

	// Determinar extensión
	ext := "jpg"
	if strings.Contains(command.ContentType, "png") {
		ext = "png"
	} else if strings.Contains(command.ContentType, "webp") {
		ext = "webp"
	}

	// Subir a S3
	key := fmt.Sprintf("users/%s/avatar.%s", command.UserID, ext)
	avatarURL, err := h.fileStorage.Upload(ctx, avatarsBucket, key, resizedBuf, contentType)
	if err != nil {
		h.logger.Error("Error subiendo avatar a S3", zap.Error(err))
		return common.InternalServerErrorResponse[responses.AvatarResponse](
			constants.GetDescription(constants.CodeStorageError),
		), nil
	}

	// Actualizar perfil
	profile.SetAvatar(avatarURL)
	if err := h.profileRepo.Update(ctx, profile); err != nil {
		h.logger.Error("Error actualizando perfil con avatar", zap.Error(err))
		return common.InternalServerErrorResponse[responses.AvatarResponse](
			"Error actualizando perfil",
		), nil
	}

	// Publicar evento (fire-and-forget)
	go func() {
		if err := h.eventPublisher.Publish(context.Background(),
			events.NewAvatarChangedEvent(command.UserID, "uploaded")); err != nil {
			h.logger.Warn("Error publicando evento avatar", zap.Error(err))
		}
	}()

	avatarResp := responses.AvatarResponse{
		AvatarURL: avatarURL,
		Message:   constants.GetDescription(constants.CodeAvatarUploaded),
	}

	resp := common.OkResponse(avatarResp)
	resp.AddMessageWithType(constants.CodeAvatarUploaded,
		constants.GetDescription(constants.CodeAvatarUploaded),
		constants.MessageTypeSuccess,
	)
	return resp, nil
}

func (h *UploadAvatarHandler) resizeImage(command commands.UploadAvatarCommand) (*bytes.Buffer, string, error) {
	img, err := imaging.Decode(command.File)
	if err != nil {
		return nil, "", fmt.Errorf("error decoding image: %w", err)
	}

	// Redimensionar manteniendo aspect ratio y recortando al centro
	resized := imaging.Fill(img, avatarWidth, avatarHeight, imaging.Center, imaging.Lanczos)

	buf := new(bytes.Buffer)
	ext := strings.ToLower(filepath.Ext(command.Filename))

	switch ext {
	case ".png":
		err = imaging.Encode(buf, resized, imaging.PNG)
		return buf, "image/png", err
	default:
		err = imaging.Encode(buf, resized, imaging.JPEG, imaging.JPEGQuality(85))
		return buf, "image/jpeg", err
	}
}

var _ mediator.RequestHandler[commands.UploadAvatarCommand, responses.AvatarResponse] = (*UploadAvatarHandler)(nil)

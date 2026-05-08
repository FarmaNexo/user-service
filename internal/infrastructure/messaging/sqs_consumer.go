// internal/infrastructure/messaging/sqs_consumer.go
package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/farmanexo/user-service/internal/domain/entities"
	"github.com/farmanexo/user-service/internal/domain/events"
	"github.com/farmanexo/user-service/internal/domain/repositories"
	"github.com/farmanexo/user-service/pkg/config"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// SQSEventConsumer consume eventos de la cola de Auth Service
type SQSEventConsumer struct {
	sqsClient   *sqs.Client
	queueURL    string
	profileRepo repositories.ProfileRepository
	prefsRepo   repositories.PreferencesRepository
	addressRepo repositories.AddressRepository
	logger      *zap.Logger
	stopCh      chan struct{}
}

// NewSQSEventConsumer crea una nueva instancia del consumer
func NewSQSEventConsumer(
	awsCfg config.AWSConfig,
	sqsCfg config.SQSConfig,
	profileRepo repositories.ProfileRepository,
	prefsRepo repositories.PreferencesRepository,
	addressRepo repositories.AddressRepository,
	logger *zap.Logger,
) (*SQSEventConsumer, error) {
	optFns := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(awsCfg.Region),
	}

	if awsCfg.Endpoint != "" {
		optFns = append(optFns, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("test", "test", ""),
		))
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(), optFns...)
	if err != nil {
		return nil, fmt.Errorf("error cargando configuración AWS: %w", err)
	}

	sqsOptFns := []func(*sqs.Options){}
	if awsCfg.Endpoint != "" {
		sqsOptFns = append(sqsOptFns, func(o *sqs.Options) {
			o.BaseEndpoint = &awsCfg.Endpoint
		})
	}

	sqsClient := sqs.NewFromConfig(cfg, sqsOptFns...)

	logger.Info("SQS EventConsumer inicializado",
		zap.String("queue_url", sqsCfg.AuthEventsQueueURL),
	)

	return &SQSEventConsumer{
		sqsClient:   sqsClient,
		queueURL:    sqsCfg.AuthEventsQueueURL,
		profileRepo: profileRepo,
		prefsRepo:   prefsRepo,
		addressRepo: addressRepo,
		logger:      logger,
		stopCh:      make(chan struct{}),
	}, nil
}

// Start inicia el consumer en un goroutine
func (c *SQSEventConsumer) Start(ctx context.Context) {
	c.logger.Info("Iniciando SQS EventConsumer",
		zap.String("queue_url", c.queueURL),
	)

	go c.pollMessages(ctx)
}

// Stop detiene el consumer
func (c *SQSEventConsumer) Stop() {
	close(c.stopCh)
	c.logger.Info("SQS EventConsumer detenido")
}

func (c *SQSEventConsumer) pollMessages(ctx context.Context) {
	for {
		select {
		case <-c.stopCh:
			return
		case <-ctx.Done():
			return
		default:
			c.receiveAndProcess(ctx)
		}
	}
}

func (c *SQSEventConsumer) receiveAndProcess(ctx context.Context) {
	input := &sqs.ReceiveMessageInput{
		QueueUrl:            &c.queueURL,
		MaxNumberOfMessages: 10,
		WaitTimeSeconds:     20, // Long polling
	}

	result, err := c.sqsClient.ReceiveMessage(ctx, input)
	if err != nil {
		c.logger.Warn("Error recibiendo mensajes de SQS",
			zap.Error(err),
		)
		time.Sleep(5 * time.Second)
		return
	}

	for _, message := range result.Messages {
		c.processMessage(ctx, message)
	}
}

func (c *SQSEventConsumer) processMessage(ctx context.Context, message sqstypes.Message) {
	if message.Body == nil {
		return
	}

	var authEvent events.AuthEvent
	if err := json.Unmarshal([]byte(*message.Body), &authEvent); err != nil {
		c.logger.Error("Error deserializando evento",
			zap.Error(err),
			zap.String("body", *message.Body),
		)
		c.deleteMessage(ctx, message)
		return
	}

	c.logger.Info("Evento recibido de Auth Service",
		zap.String("event_type", authEvent.EventType),
		zap.String("user_id", authEvent.UserID),
	)

	switch authEvent.EventType {
	case events.AuthEventUserRegistered:
		c.handleUserRegistered(ctx, authEvent)
	case events.AuthEventUserEmailChanged:
		c.handleUserEmailChanged(ctx, authEvent)
	case events.AuthEventUserRoleChanged:
		c.handleUserRoleChanged(ctx, authEvent)
	case events.AuthEventUserDeleted:
		c.handleUserDeleted(ctx, authEvent)
	default:
		c.logger.Debug("Evento ignorado",
			zap.String("event_type", authEvent.EventType),
		)
	}

	// Eliminar mensaje de la cola después de procesarlo
	c.deleteMessage(ctx, message)
}

func (c *SQSEventConsumer) handleUserRegistered(ctx context.Context, event events.AuthEvent) {
	userID, err := uuid.Parse(event.UserID)
	if err != nil {
		c.logger.Error("Error parseando user_id del evento",
			zap.String("user_id", event.UserID),
			zap.Error(err),
		)
		return
	}

	// Verificar si ya existe el perfil (idempotencia)
	exists, err := c.profileRepo.Exists(ctx, userID)
	if err != nil {
		c.logger.Error("Error verificando perfil existente",
			zap.String("user_id", event.UserID),
			zap.Error(err),
		)
		return
	}

	if exists {
		c.logger.Debug("Perfil ya existe, ignorando evento duplicado",
			zap.String("user_id", event.UserID),
		)
		return
	}

	// Usar full_name del evento si existe, sino usar email como nombre (fallback defensivo).
	fullName := event.FullName
	if fullName == "" {
		fullName = event.Email
	}

	// Crear perfil inicial. Email y Role son la proyección read-side de auth-service.
	profile := entities.NewUserProfile(userID, fullName)
	profile.Email = event.Email
	if event.Role != "" {
		profile.Role = event.Role
	}
	if event.Phone != "" {
		profile.Phone = event.Phone
	}
	if err := c.profileRepo.Create(ctx, profile); err != nil {
		c.logger.Error("Error creando perfil inicial",
			zap.String("user_id", event.UserID),
			zap.Error(err),
		)
		return
	}

	// Crear preferencias con valores por defecto
	preferences := entities.NewDefaultPreferences(userID)
	if err := c.prefsRepo.Create(ctx, preferences); err != nil {
		c.logger.Error("Error creando preferencias iniciales",
			zap.String("user_id", event.UserID),
			zap.Error(err),
		)
		return
	}

	c.logger.Info("Perfil y preferencias iniciales creados para nuevo usuario",
		zap.String("user_id", event.UserID),
		zap.String("full_name", fullName),
	)
}

// handleUserEmailChanged sincroniza la proyección local del email cuando cambia en auth-service.
func (c *SQSEventConsumer) handleUserEmailChanged(ctx context.Context, event events.AuthEvent) {
	userID, err := uuid.Parse(event.UserID)
	if err != nil {
		c.logger.Error("Error parseando user_id del evento",
			zap.String("user_id", event.UserID), zap.Error(err))
		return
	}
	if event.Email == "" {
		c.logger.Warn("USER_EMAIL_CHANGED sin email", zap.String("user_id", event.UserID))
		return
	}

	profile, err := c.profileRepo.FindByUserID(ctx, userID)
	if err != nil {
		c.logger.Error("Error buscando perfil para actualizar email",
			zap.String("user_id", event.UserID), zap.Error(err))
		return
	}
	profile.Email = event.Email
	if err := c.profileRepo.Update(ctx, profile); err != nil {
		c.logger.Error("Error actualizando email del perfil",
			zap.String("user_id", event.UserID), zap.Error(err))
		return
	}
	c.logger.Info("Email del perfil sincronizado desde auth-service",
		zap.String("user_id", event.UserID))
}

// handleUserRoleChanged sincroniza la proyección local del role cuando cambia en auth-service.
func (c *SQSEventConsumer) handleUserRoleChanged(ctx context.Context, event events.AuthEvent) {
	userID, err := uuid.Parse(event.UserID)
	if err != nil {
		c.logger.Error("Error parseando user_id del evento",
			zap.String("user_id", event.UserID), zap.Error(err))
		return
	}
	if event.Role == "" {
		c.logger.Warn("USER_ROLE_CHANGED sin role", zap.String("user_id", event.UserID))
		return
	}

	profile, err := c.profileRepo.FindByUserID(ctx, userID)
	if err != nil {
		c.logger.Error("Error buscando perfil para actualizar role",
			zap.String("user_id", event.UserID), zap.Error(err))
		return
	}
	profile.Role = event.Role
	if err := c.profileRepo.Update(ctx, profile); err != nil {
		c.logger.Error("Error actualizando role del perfil",
			zap.String("user_id", event.UserID), zap.Error(err))
		return
	}
	c.logger.Info("Role del perfil sincronizado desde auth-service",
		zap.String("user_id", event.UserID), zap.String("new_role", event.Role))
}

// handleUserDeleted reacciona al ejercicio del derecho ARCO de cancelación del usuario.
// Anonimiza la proyección del perfil (email, nombre, teléfono, bio, avatar, fecha de
// nacimiento), elimina todas las direcciones registradas y resetea las preferencias a
// valores por defecto. Es idempotente: si no hay perfil o ya fue anonimizado, no hace nada.
func (c *SQSEventConsumer) handleUserDeleted(ctx context.Context, event events.AuthEvent) {
	userID, err := uuid.Parse(event.UserID)
	if err != nil {
		c.logger.Error("Error parseando user_id del evento USER_DELETED",
			zap.String("user_id", event.UserID), zap.Error(err))
		return
	}

	// 1. Anonimizar perfil (si existe)
	profile, err := c.profileRepo.FindByUserID(ctx, userID)
	if err != nil {
		c.logger.Warn("Perfil no encontrado al procesar USER_DELETED (posiblemente ya limpio)",
			zap.String("user_id", event.UserID), zap.Error(err))
	} else if profile != nil {
		profile.Email = fmt.Sprintf("deleted-%s@deleted.local", userID.String())
		profile.FullName = "USUARIO ELIMINADO"
		profile.Phone = ""
		profile.Bio = ""
		profile.AvatarURL = ""
		profile.DateOfBirth = nil
		if err := c.profileRepo.Update(ctx, profile); err != nil {
			c.logger.Error("Error anonimizando perfil en USER_DELETED",
				zap.String("user_id", event.UserID), zap.Error(err))
			return
		}
	}

	// 2. Eliminar todas las direcciones del usuario
	addresses, err := c.addressRepo.FindByUserID(ctx, userID)
	if err != nil {
		c.logger.Error("Error listando direcciones para eliminar",
			zap.String("user_id", event.UserID), zap.Error(err))
	} else {
		for _, addr := range addresses {
			if err := c.addressRepo.Delete(ctx, addr.ID); err != nil {
				c.logger.Warn("Error eliminando dirección (continúa con el resto)",
					zap.String("user_id", event.UserID),
					zap.String("address_id", addr.ID.String()),
					zap.Error(err))
			}
		}
	}

	// 3. Resetear preferencias a valores por defecto
	defaultPrefs := entities.NewDefaultPreferences(userID)
	if err := c.prefsRepo.Update(ctx, defaultPrefs); err != nil {
		// No bloqueante: las preferencias son cosméticas y pueden recrearse
		c.logger.Warn("Error reseteando preferencias (continúa el flujo)",
			zap.String("user_id", event.UserID), zap.Error(err))
	}

	c.logger.Info("Usuario anonimizado en user-service por USER_DELETED",
		zap.String("user_id", event.UserID),
		zap.Int("addresses_deleted", len(addresses)))
}

func (c *SQSEventConsumer) deleteMessage(ctx context.Context, message sqstypes.Message) {
	input := &sqs.DeleteMessageInput{
		QueueUrl:      &c.queueURL,
		ReceiptHandle: message.ReceiptHandle,
	}

	_, err := c.sqsClient.DeleteMessage(ctx, input)
	if err != nil {
		c.logger.Warn("Error eliminando mensaje de SQS",
			zap.Error(err),
		)
	}
}

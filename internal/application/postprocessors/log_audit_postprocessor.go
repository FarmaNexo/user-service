// internal/application/postprocessors/log_audit_postprocessor.go
package postprocessors

import (
	"context"
	"time"

	"github.com/farmanexo/user-service/internal/application/commands"
	"github.com/farmanexo/user-service/internal/application/queries"
	"github.com/farmanexo/user-service/pkg/mediator"
	"go.uber.org/zap"
)

// LogAuditPostProcessor registra eventos de auditoría
type LogAuditPostProcessor struct {
	logger *zap.Logger
}

func NewLogAuditPostProcessor(logger *zap.Logger) *LogAuditPostProcessor {
	return &LogAuditPostProcessor{logger: logger}
}

func (p *LogAuditPostProcessor) Process(ctx context.Context, request interface{}, response interface{}) error {
	userID := p.getUserIDFromContext(ctx)
	correlationID := mediator.GetCorrelationID(ctx)
	isSuccess := p.checkSuccess(response)

	switch request.(type) {
	case commands.UpdateProfileCommand, *commands.UpdateProfileCommand:
		p.logAudit("PROFILE_UPDATED", userID, correlationID, isSuccess)
	case commands.UploadAvatarCommand, *commands.UploadAvatarCommand:
		p.logAudit("AVATAR_UPLOADED", userID, correlationID, isSuccess)
	case commands.DeleteAvatarCommand, *commands.DeleteAvatarCommand:
		p.logAudit("AVATAR_DELETED", userID, correlationID, isSuccess)
	case commands.CreateAddressCommand, *commands.CreateAddressCommand:
		p.logAudit("ADDRESS_CREATED", userID, correlationID, isSuccess)
	case commands.UpdateAddressCommand, *commands.UpdateAddressCommand:
		p.logAudit("ADDRESS_UPDATED", userID, correlationID, isSuccess)
	case commands.DeleteAddressCommand, *commands.DeleteAddressCommand:
		p.logAudit("ADDRESS_DELETED", userID, correlationID, isSuccess)
	case commands.UpdatePreferencesCommand, *commands.UpdatePreferencesCommand:
		p.logAudit("PREFERENCES_UPDATED", userID, correlationID, isSuccess)
	case queries.GetProfileQuery, *queries.GetProfileQuery:
		p.logAudit("PROFILE_VIEWED", userID, correlationID, isSuccess)
	default:
		p.logger.Debug("Post-processor: comando sin auditoría configurada")
	}

	return nil
}

func (p *LogAuditPostProcessor) logAudit(eventType, userID, correlationID string, success bool) {
	p.logger.Info("AUDIT",
		zap.String("event_type", eventType),
		zap.Bool("success", success),
		zap.String("correlation_id", correlationID),
		zap.String("user_id", userID),
		zap.Time("timestamp", time.Now()),
	)
}

func (p *LogAuditPostProcessor) checkSuccess(response interface{}) bool {
	if resp, ok := response.(interface{ IsValid() bool }); ok {
		return resp.IsValid()
	}
	return false
}

func (p *LogAuditPostProcessor) getUserIDFromContext(ctx context.Context) string {
	userID, _ := mediator.GetUserID(ctx)
	if userID == "" {
		return "ANONYMOUS"
	}
	return userID
}

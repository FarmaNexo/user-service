// internal/domain/services/event_publisher.go
package services

import (
	"context"

	"github.com/farmanexo/user-service/internal/domain/events"
)

// EventPublisher define la interfaz para publicar eventos de usuario
type EventPublisher interface {
	Publish(ctx context.Context, event events.UserEvent) error
}

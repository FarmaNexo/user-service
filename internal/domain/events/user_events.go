// internal/domain/events/user_events.go
package events

import "time"

// Tipos de eventos de usuario
const (
	EventProfileUpdated = "PROFILE_UPDATED"
	EventAvatarChanged  = "AVATAR_CHANGED"
	EventAddressCreated = "ADDRESS_CREATED"
)

// UserEvent representa un evento del User Service
type UserEvent struct {
	EventType string            `json:"event_type"`
	UserID    string            `json:"user_id"`
	Timestamp time.Time         `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// NewProfileUpdatedEvent crea un evento de perfil actualizado
func NewProfileUpdatedEvent(userID string) UserEvent {
	return UserEvent{
		EventType: EventProfileUpdated,
		UserID:    userID,
		Timestamp: time.Now(),
		Metadata: map[string]string{
			"source": "user-service",
		},
	}
}

// NewAvatarChangedEvent crea un evento de avatar cambiado
func NewAvatarChangedEvent(userID string, action string) UserEvent {
	return UserEvent{
		EventType: EventAvatarChanged,
		UserID:    userID,
		Timestamp: time.Now(),
		Metadata: map[string]string{
			"source": "user-service",
			"action": action,
		},
	}
}

// NewAddressCreatedEvent crea un evento de dirección creada
func NewAddressCreatedEvent(userID string, addressID string) UserEvent {
	return UserEvent{
		EventType: EventAddressCreated,
		UserID:    userID,
		Timestamp: time.Now(),
		Metadata: map[string]string{
			"source":     "user-service",
			"address_id": addressID,
		},
	}
}

// AuthEvent representa un evento recibido desde Auth Service
type AuthEvent struct {
	EventType string            `json:"event_type"`
	UserID    string            `json:"user_id"`
	Email     string            `json:"email,omitempty"`
	FullName  string            `json:"full_name,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

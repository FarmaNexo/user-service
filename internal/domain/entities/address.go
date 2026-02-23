// internal/domain/entities/address.go
package entities

import (
	"time"

	"github.com/google/uuid"
)

// Address representa una dirección del usuario
type Address struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index:idx_addresses_user_id" json:"user_id"`
	Label      string    `gorm:"type:varchar(50)" json:"label,omitempty"`
	Street     string    `gorm:"type:varchar(255);not null" json:"street"`
	City       string    `gorm:"type:varchar(100);not null" json:"city"`
	State      string    `gorm:"type:varchar(100)" json:"state,omitempty"`
	PostalCode string    `gorm:"type:varchar(20)" json:"postal_code,omitempty"`
	Country    string    `gorm:"type:varchar(100);not null;default:'Perú'" json:"country"`
	IsDefault  bool      `gorm:"not null;default:false" json:"is_default"`
	Latitude   *float64  `gorm:"type:decimal(10,8)" json:"latitude,omitempty"`
	Longitude  *float64  `gorm:"type:decimal(11,8)" json:"longitude,omitempty"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName retorna el nombre de la tabla con schema
func (Address) TableName() string {
	return "\"user\".addresses"
}

// NewAddress crea una nueva dirección
func NewAddress(userID uuid.UUID, label, street, city, state, postalCode, country string, isDefault bool) *Address {
	if country == "" {
		country = "Perú"
	}
	now := time.Now()
	return &Address{
		ID:         uuid.New(),
		UserID:     userID,
		Label:      label,
		Street:     street,
		City:       city,
		State:      state,
		PostalCode: postalCode,
		Country:    country,
		IsDefault:  isDefault,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

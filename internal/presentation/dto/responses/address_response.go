// internal/presentation/dto/responses/address_response.go
package responses

import (
	"time"

	"github.com/farmanexo/user-service/internal/domain/entities"
	"github.com/google/uuid"
)

// AddressResponse DTO de respuesta de dirección
type AddressResponse struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	Label      string    `json:"label,omitempty"`
	Street     string    `json:"street"`
	City       string    `json:"city"`
	State      string    `json:"state,omitempty"`
	PostalCode string    `json:"postal_code,omitempty"`
	Country    string    `json:"country"`
	IsDefault  bool      `json:"is_default"`
	Latitude   *float64  `json:"latitude,omitempty"`
	Longitude  *float64  `json:"longitude,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// NewAddressResponse crea un AddressResponse desde una entidad Address
func NewAddressResponse(address *entities.Address) AddressResponse {
	return AddressResponse{
		ID:         address.ID,
		UserID:     address.UserID,
		Label:      address.Label,
		Street:     address.Street,
		City:       address.City,
		State:      address.State,
		PostalCode: address.PostalCode,
		Country:    address.Country,
		IsDefault:  address.IsDefault,
		Latitude:   address.Latitude,
		Longitude:  address.Longitude,
		CreatedAt:  address.CreatedAt,
		UpdatedAt:  address.UpdatedAt,
	}
}

// AddressListResponse DTO de respuesta de lista de direcciones
type AddressListResponse struct {
	Addresses []AddressResponse `json:"addresses"`
	Total     int               `json:"total"`
}

// NewAddressListResponse crea un AddressListResponse desde un slice de entidades
func NewAddressListResponse(addresses []entities.Address) AddressListResponse {
	list := make([]AddressResponse, len(addresses))
	for i, addr := range addresses {
		list[i] = NewAddressResponse(&addr)
	}
	return AddressListResponse{
		Addresses: list,
		Total:     len(list),
	}
}
